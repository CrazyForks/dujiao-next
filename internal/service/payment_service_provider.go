package service

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/dujiao-next/internal/constants"
	"github.com/dujiao-next/internal/models"
	"github.com/dujiao-next/internal/payment/alipay"
	"github.com/dujiao-next/internal/payment/bepusdt"
	"github.com/dujiao-next/internal/payment/epay"
	"github.com/dujiao-next/internal/payment/epusdt"
	"github.com/dujiao-next/internal/payment/okpay"
	"github.com/dujiao-next/internal/payment/paypal"
	"github.com/dujiao-next/internal/payment/provider"
	"github.com/dujiao-next/internal/payment/stripe"
	"github.com/dujiao-next/internal/payment/tokenpay"
	"github.com/dujiao-next/internal/payment/wechatpay"

	"github.com/shopspring/decimal"
)

// appendExchangeInfo 将 payment.Amount 更新为转换后金额（与网关实际交互的金额），
// 原始金额记录到 ProviderPayload 用于审计追踪。
func appendExchangeInfo(payment *models.Payment, convertedAmount, exchangeRate, originalAmount, originalCurrency string) {
	if d, err := decimal.NewFromString(convertedAmount); err == nil {
		payment.Amount = models.Money{Decimal: d}
	}
	if payment.ProviderPayload == nil {
		payment.ProviderPayload = models.JSON{}
	}
	payment.ProviderPayload["exchange_rate"] = strings.TrimSpace(exchangeRate)
	payment.ProviderPayload["original_amount"] = originalAmount
	payment.ProviderPayload["original_currency"] = originalCurrency
}

func (s *PaymentService) applyProviderPayment(input CreatePaymentInput, order *models.Order, channel *models.PaymentChannel, payment *models.Payment) (err error) {
	providerType := strings.ToLower(strings.TrimSpace(channel.ProviderType))
	channelType := strings.ToLower(strings.TrimSpace(channel.ChannelType))
	gatewayCtx, cancel := detachOutboundRequestContext(input.Context)
	defer cancel()
	payment.GatewayOrderNo = resolveGatewayOrderNo(channel, payment)
	providerOrderNo := resolveProviderOrderNo(order.OrderNo, payment)
	log := paymentLogger(
		"order_id", order.ID,
		"order_no", order.OrderNo,
		"gateway_order_no", payment.GatewayOrderNo,
		"payment_id", payment.ID,
		"channel_id", channel.ID,
		"provider_type", providerType,
		"channel_type", channelType,
		"interaction_mode", channel.InteractionMode,
	)
	defer func() {
		if err != nil {
			log.Errorw("payment_provider_apply_failed", "error", err)
			return
		}
		log.Infow("payment_provider_apply_success")
	}()
	if s.paymentProviderRegistry == nil {
		return ErrPaymentProviderNotSupported
	}

	// C1b: reject guard — P1.2b adapter wrapper 尚未实现 currency conversion（P1.2c 修复）。
	// 针对 official/epay provider，如果 ConfigJSON 配置了 target_currency + exchange_rate（非 1:1），
	// 拒绝创建支付，避免静默 money loss（exchange_rate ≠ 1 时网关会将 CNY 金额当目标货币处理）。
	// okpay 的 exchange_rate 由 okpay native 包自行处理，不受此拦截。
	// bepusdt/epusdt/tokenpay 无汇率转换概念，不受影响。
	switch providerType {
	case constants.PaymentProviderOfficial, constants.PaymentProviderEpay:
		if needsCurrencyConversion(channel.ConfigJSON) {
			return fmt.Errorf("%w: channel exchange_rate conversion requires P1.2c, blocked to prevent money loss", ErrPaymentChannelConfigInvalid)
		}
	}

	p, ok := s.paymentProviderRegistry.Lookup(channel.ProviderType, channel.ChannelType)
	if !ok {
		return ErrPaymentProviderNotSupported
	}

	// 构造 provider.CreateInput。
	// NotifyURL / ReturnURL 留空：各 adapter/native 包均实现 "input值 || cfg值" fallback，
	// 空值时自动读取 channel.ConfigJSON 里配置的 notify_url / return_url。
	// P1.2c 会把 returnURL tracking marker 和 currency conversion 下沉到 adapter wrapper。
	extra := models.JSON{}
	if interactionMode := strings.TrimSpace(channel.InteractionMode); interactionMode != "" {
		extra["interaction_mode"] = interactionMode
	}
	// order_user_key 是 tokenpay 必须的稳定用户标识符；其他 adapter 忽略此字段。
	extra["order_user_key"] = resolveTokenPayOrderUserKey(order)

	createInput := provider.CreateInput{
		PaymentID:   payment.ID,
		OrderID:     order.ID,
		OrderNo:     providerOrderNo,
		Subject:     buildOrderSubject(order),
		Amount:      payment.Amount,
		Currency:    payment.Currency,
		ClientIP:    strings.TrimSpace(input.ClientIP),
		ChannelType: channel.ChannelType,
		Extra:       extra,
		// NotifyURL / ReturnURL 留空，由各 adapter 从 cfg 读取
	}

	result, err := p.CreatePayment(gatewayCtx, channel.ConfigJSON, createInput)
	if err != nil {
		return mapProviderErrorToService(err)
	}

	// 把 result 写回 payment 字段
	payment.PayURL = strings.TrimSpace(result.RedirectURL)
	payment.QRCode = strings.TrimSpace(result.QRCodeURL)
	if result.ProviderRef != "" {
		payment.ProviderRef = result.ProviderRef
	}
	// 确保 ProviderRef 始终有值（各 adapter 可能返回空，如 wechat CreatePayment 阶段）
	if payment.ProviderRef == "" {
		payment.ProviderRef = order.OrderNo
	}
	if result.Payload != nil {
		payment.ProviderPayload = result.Payload
	}
	if result.CurrencySent != "" {
		payment.Currency = result.CurrencySent
	}
	payment.Status = constants.PaymentStatusPending
	payment.UpdatedAt = time.Now()

	if err := s.paymentRepo.Update(payment); err != nil {
		return ErrPaymentUpdateFailed
	}
	return nil
}

// ValidateChannel 校验支付渠道配置
func (s *PaymentService) ValidateChannel(channel *models.PaymentChannel) error {
	if channel == nil {
		return ErrPaymentChannelConfigInvalid
	}
	feeRate := channel.FeeRate.Decimal.Round(2)
	if feeRate.LessThan(decimal.Zero) || feeRate.GreaterThan(decimal.NewFromInt(100)) {
		return ErrPaymentChannelConfigInvalid
	}
	fixedFee := channel.FixedFee.Decimal.Round(2)
	// decimal(6,2) max value is 9999.99
	if fixedFee.LessThan(decimal.Zero) || fixedFee.GreaterThanOrEqual(decimal.NewFromInt(10000)) {
		return ErrPaymentChannelConfigInvalid
	}
	minAmount := channel.MinAmount.Decimal.Round(2)
	maxAmount := channel.MaxAmount.Decimal.Round(2)
	amountOverflow20_2 := decimal.NewFromInt(1000000000000000000)
	// min/max amount are stored as decimal(20,2), max allowed is 999999999999999999.99.
	if minAmount.LessThan(decimal.Zero) || minAmount.GreaterThanOrEqual(amountOverflow20_2) || maxAmount.LessThan(decimal.Zero) || maxAmount.GreaterThanOrEqual(amountOverflow20_2) {
		return ErrPaymentChannelConfigInvalid
	}
	if maxAmount.GreaterThan(decimal.Zero) && minAmount.GreaterThan(maxAmount) {
		return ErrPaymentChannelConfigInvalid
	}
	providerType := strings.ToLower(strings.TrimSpace(channel.ProviderType))
	switch providerType {
	case constants.PaymentProviderEpay:
		if !epay.IsSupportedChannelType(channel.ChannelType) {
			return fmt.Errorf("%w: unsupported channel_type %s", ErrPaymentChannelConfigInvalid, channel.ChannelType)
		}
		mode := strings.ToLower(strings.TrimSpace(channel.InteractionMode))
		if mode != constants.PaymentInteractionQR && mode != constants.PaymentInteractionRedirect {
			return ErrPaymentChannelConfigInvalid
		}
		cfg, err := epay.ParseConfig(channel.ConfigJSON)
		if err != nil {
			return fmt.Errorf("%w: %v", ErrPaymentChannelConfigInvalid, err)
		}
		if err := epay.ValidateConfig(cfg); err != nil {
			return fmt.Errorf("%w: %v", ErrPaymentChannelConfigInvalid, err)
		}
		return nil
	case constants.PaymentProviderBepusdt:
		if !bepusdt.IsSupportedChannelType(channel.ChannelType) {
			return fmt.Errorf("%w: unsupported channel_type %s", ErrPaymentChannelConfigInvalid, channel.ChannelType)
		}
		if strings.ToLower(strings.TrimSpace(channel.InteractionMode)) != constants.PaymentInteractionRedirect &&
			strings.ToLower(strings.TrimSpace(channel.InteractionMode)) != constants.PaymentInteractionQR {
			return ErrPaymentChannelConfigInvalid
		}
		cfg, err := bepusdt.ParseConfig(channel.ConfigJSON)
		if err != nil {
			return fmt.Errorf("%w: %v", ErrPaymentChannelConfigInvalid, err)
		}
		if err := bepusdt.ValidateConfig(cfg); err != nil {
			return fmt.Errorf("%w: %v", ErrPaymentChannelConfigInvalid, err)
		}
		return nil
	case constants.PaymentProviderEpusdt:
		if strings.ToLower(strings.TrimSpace(channel.InteractionMode)) != constants.PaymentInteractionRedirect {
			return ErrPaymentChannelConfigInvalid
		}
		cfg, err := epusdt.ParseConfig(channel.ConfigJSON)
		if err != nil {
			return fmt.Errorf("%w: %v", ErrPaymentChannelConfigInvalid, err)
		}
		cfg.Normalize()
		if err := epusdt.ValidateConfig(cfg); err != nil {
			return fmt.Errorf("%w: %v", ErrPaymentChannelConfigInvalid, err)
		}
		return nil
	case constants.PaymentProviderOkpay:
		if !okpay.IsSupportedChannelType(channel.ChannelType) {
			return fmt.Errorf("%w: unsupported channel_type %s", ErrPaymentChannelConfigInvalid, channel.ChannelType)
		}
		mode := strings.ToLower(strings.TrimSpace(channel.InteractionMode))
		if mode != constants.PaymentInteractionQR && mode != constants.PaymentInteractionRedirect {
			return ErrPaymentChannelConfigInvalid
		}
		cfg, err := okpay.ParseConfig(channel.ConfigJSON)
		if err != nil {
			return fmt.Errorf("%w: %v", ErrPaymentChannelConfigInvalid, err)
		}
		if strings.TrimSpace(cfg.Coin) == "" {
			cfg.Coin = okpay.ResolveCoin(channel.ChannelType)
		}
		if err := okpay.ValidateConfig(cfg); err != nil {
			return fmt.Errorf("%w: %v", ErrPaymentChannelConfigInvalid, err)
		}
		return nil
	case constants.PaymentProviderTokenpay:
		if strings.ToLower(strings.TrimSpace(channel.InteractionMode)) != constants.PaymentInteractionRedirect &&
			strings.ToLower(strings.TrimSpace(channel.InteractionMode)) != constants.PaymentInteractionQR {
			return ErrPaymentChannelConfigInvalid
		}
		cfg, err := tokenpay.ParseConfig(channel.ConfigJSON)
		if err != nil {
			return fmt.Errorf("%w: %v", ErrPaymentChannelConfigInvalid, err)
		}
		if strings.TrimSpace(cfg.Currency) == "" {
			cfg.Currency = tokenpay.DefaultCurrency
		}
		if strings.TrimSpace(cfg.NotifyURL) == "" {
			return fmt.Errorf("%w: notify_url is required", ErrPaymentChannelConfigInvalid)
		}
		if err := tokenpay.ValidateConfig(cfg); err != nil {
			return fmt.Errorf("%w: %v", ErrPaymentChannelConfigInvalid, err)
		}
		return nil
	case constants.PaymentProviderOfficial:
		channelType := strings.ToLower(strings.TrimSpace(channel.ChannelType))
		switch channelType {
		case constants.PaymentChannelTypePaypal:
			if strings.ToLower(strings.TrimSpace(channel.InteractionMode)) != constants.PaymentInteractionRedirect {
				return ErrPaymentChannelConfigInvalid
			}
			cfg, err := paypal.ParseConfig(channel.ConfigJSON)
			if err != nil {
				return fmt.Errorf("%w: %v", ErrPaymentChannelConfigInvalid, err)
			}
			if err := paypal.ValidateConfig(cfg); err != nil {
				return fmt.Errorf("%w: %v", ErrPaymentChannelConfigInvalid, err)
			}
			return nil
		case constants.PaymentChannelTypeAlipay:
			cfg, err := alipay.ParseConfig(channel.ConfigJSON)
			if err != nil {
				return fmt.Errorf("%w: %v", ErrPaymentChannelConfigInvalid, err)
			}
			if err := alipay.ValidateConfig(cfg, channel.InteractionMode); err != nil {
				return fmt.Errorf("%w: %v", ErrPaymentChannelConfigInvalid, err)
			}
			return nil
		case constants.PaymentChannelTypeWechat:
			cfg, err := wechatpay.ParseConfig(channel.ConfigJSON)
			if err != nil {
				return fmt.Errorf("%w: %v", ErrPaymentChannelConfigInvalid, err)
			}
			if err := wechatpay.ValidateConfig(cfg, channel.InteractionMode); err != nil {
				return fmt.Errorf("%w: %v", ErrPaymentChannelConfigInvalid, err)
			}
			return nil
		case constants.PaymentChannelTypeStripe:
			if strings.ToLower(strings.TrimSpace(channel.InteractionMode)) != constants.PaymentInteractionRedirect {
				return ErrPaymentChannelConfigInvalid
			}
			cfg, err := stripe.ParseConfig(channel.ConfigJSON)
			if err != nil {
				return fmt.Errorf("%w: %v", ErrPaymentChannelConfigInvalid, err)
			}
			if err := stripe.ValidateConfig(cfg); err != nil {
				return fmt.Errorf("%w: %v", ErrPaymentChannelConfigInvalid, err)
			}
			return nil
		default:
			return ErrPaymentProviderNotSupported
		}
	default:
		return ErrPaymentProviderNotSupported
	}
}

func mapPaypalStatus(status string) (string, bool) {
	status = strings.ToUpper(strings.TrimSpace(status))
	switch status {
	case "COMPLETED":
		return constants.PaymentStatusSuccess, true
	case "PENDING", "APPROVED", "CREATED", "SAVED":
		return constants.PaymentStatusPending, true
	case "DECLINED", "DENIED", "FAILED", "VOIDED":
		return constants.PaymentStatusFailed, true
	default:
		return "", false
	}
}

func resolveTokenPayOrderUserKey(order *models.Order) string {
	if order == nil {
		return ""
	}
	if order.UserID > 0 {
		return strconv.FormatUint(uint64(order.UserID), 10)
	}
	if guestEmail := strings.TrimSpace(order.GuestEmail); guestEmail != "" {
		return guestEmail
	}
	return strings.TrimSpace(order.OrderNo)
}

// needsCurrencyConversion 检测 channel.ConfigJSON 是否配置了汇率转换（target_currency + exchange_rate 均非空）。
// 用于 C1b reject guard：official/epay adapter 尚未实现 conversion，配置了转换时拒绝创建支付（P1.2c 实现真正 fix）。
// 注意：okpay 的 exchange_rate 语义不同（由 okpay native 包自行处理），不走此路径。
func needsCurrencyConversion(raw models.JSON) bool {
	if raw == nil {
		return false
	}
	targetCurrency, _ := raw["target_currency"].(string)
	exchangeRate, _ := raw["exchange_rate"].(string)
	return strings.TrimSpace(targetCurrency) != "" && strings.TrimSpace(exchangeRate) != ""
}
