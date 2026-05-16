package service

import (
	"errors"
	"strings"
	"time"

	"github.com/dujiao-next/internal/constants"
	"github.com/dujiao-next/internal/models"
	"github.com/dujiao-next/internal/payment/paypal"
	"github.com/dujiao-next/internal/payment/provider"
	"github.com/dujiao-next/internal/payment/stripe"
	"github.com/dujiao-next/internal/payment/wechatpay"

	"github.com/shopspring/decimal"
)

// HandleSyncCallback 处理同步 form callback（alipay/epay/epusdt/bepusdt/tokenpay/okpay）。
// 通过 Registry 找到 adapter 的 CallbackVerifier 能力解析并验签 form/body，然后调 HandleCallback。
// channel 必须由 caller 加载好传入（handler 负责找到 payment→channel 并验证类型）。
//
// P1.2c Task 7: alipay callback handler pilot；Task 8/9 将同模式复用此方法。
func (s *PaymentService) HandleSyncCallback(
	channel *models.PaymentChannel,
	form map[string][]string,
	body []byte,
) (*models.Payment, error) {
	if channel == nil {
		return nil, ErrPaymentChannelNotFound
	}
	if s.paymentProviderRegistry == nil {
		return nil, ErrPaymentProviderNotSupported
	}

	p, ok := s.paymentProviderRegistry.Lookup(channel.ProviderType, channel.ChannelType)
	if !ok {
		return nil, ErrPaymentProviderNotSupported
	}
	verifier, ok := p.(provider.CallbackVerifier)
	if !ok {
		return nil, ErrPaymentProviderNotSupported
	}

	result, err := verifier.VerifyCallback(channel.ConfigJSON, form, body)
	if err != nil {
		return nil, mapProviderErrorToService(err)
	}

	// WebhookResult = CallbackResult（类型别名），findWebhookPayment 直接复用。
	payment, err := s.findWebhookPayment(channel.ID, result)
	if err != nil {
		return nil, err
	}

	payload := models.JSON{}
	if result.Payload != nil {
		payload = result.Payload
	}

	callbackInput := PaymentCallbackInput{
		PaymentID:   payment.ID,
		ChannelID:   channel.ID,
		Status:      result.Status,
		ProviderRef: pickFirstNonEmpty(result.ProviderRef, payment.ProviderRef),
		Amount:      result.Amount,
		Currency:    strings.ToUpper(strings.TrimSpace(result.Currency)),
		PaidAt:      result.PaidAt,
		Payload:     payload,
	}
	return s.HandleCallback(callbackInput)
}

// HandlePaypalWebhook 处理 PayPal webhook。
// P1.2c Task 5: 退化为 thin wrapper，通过 handleWebhookViaRegistry 路由解析。
func (s *PaymentService) HandlePaypalWebhook(input WebhookCallbackInput) (*models.Payment, string, error) {
	return s.handleWebhookViaRegistry(
		input,
		constants.PaymentProviderOfficial,
		constants.PaymentChannelTypePaypal,
	)
}

// handleWebhookViaRegistry 通过 Registry 路由 webhook 解析。
// Task 5/6: paypal/stripe/wechat 共用此方法。
func (s *PaymentService) handleWebhookViaRegistry(
	input WebhookCallbackInput,
	expectedProviderType string,
	expectedChannelType string,
) (*models.Payment, string, error) {
	log := paymentLogger(
		"provider", expectedChannelType,
		"channel_id", input.ChannelID,
		"body_size", len(input.Body),
	)

	if input.ChannelID == 0 {
		log.Warnw("payment_webhook_invalid_channel_id")
		return nil, "", ErrPaymentInvalid
	}
	channel, err := s.channelRepo.GetByID(input.ChannelID)
	if err != nil {
		log.Errorw("payment_webhook_channel_fetch_failed", "error", err)
		return nil, "", ErrPaymentUpdateFailed
	}
	if channel == nil {
		log.Warnw("payment_webhook_channel_not_found")
		return nil, "", ErrPaymentChannelNotFound
	}

	providerType := strings.ToLower(strings.TrimSpace(channel.ProviderType))
	channelType := strings.ToLower(strings.TrimSpace(channel.ChannelType))
	if providerType != expectedProviderType || channelType != expectedChannelType {
		log.Warnw("payment_webhook_provider_mismatch",
			"provider_type", channel.ProviderType,
			"channel_type", channel.ChannelType,
		)
		return nil, "", ErrPaymentProviderNotSupported
	}

	if s.paymentProviderRegistry == nil {
		return nil, "", ErrPaymentProviderNotSupported
	}
	p, ok := s.paymentProviderRegistry.Lookup(channel.ProviderType, channel.ChannelType)
	if !ok {
		return nil, "", ErrPaymentProviderNotSupported
	}
	webhooker, ok := p.(provider.Webhooker)
	if !ok {
		log.Warnw("payment_webhook_capability_missing",
			"provider_type", channel.ProviderType,
			"channel_type", channel.ChannelType,
		)
		return nil, "", ErrPaymentProviderNotSupported
	}

	ctx, cancel := detachOutboundRequestContext(input.Context)
	defer cancel()

	result, err := webhooker.ParseWebhook(ctx, channel.ConfigJSON, input.Headers, input.Body, time.Now())
	if err != nil {
		log.Warnw("payment_webhook_parse_failed", "error", err)
		return nil, "", mapProviderErrorToService(err)
	}

	log.Infow("payment_webhook_parsed",
		"order_no", result.OrderNo,
		"provider_ref", result.ProviderRef,
		"status", result.Status,
	)

	// status 为空表示 adapter 判断该事件无需处理（不可识别的事件类型），直接忽略。
	if result.Status == "" {
		log.Infow("payment_webhook_status_ignored",
			"order_no", result.OrderNo,
			"provider_ref", result.ProviderRef,
		)
		return nil, "", nil
	}

	payment, err := s.findWebhookPayment(channel.ID, result)
	if err != nil {
		if errors.Is(err, ErrPaymentNotFound) {
			log.Infow("payment_webhook_payment_not_found",
				"order_no", result.OrderNo,
				"provider_ref", result.ProviderRef,
				"status", result.Status,
			)
			return nil, result.Status, nil
		}
		log.Warnw("payment_webhook_payment_lookup_failed",
			"order_no", result.OrderNo,
			"provider_ref", result.ProviderRef,
			"status", result.Status,
			"error", err,
		)
		return nil, result.Status, err
	}

	payload := models.JSON{}
	if result.Payload != nil {
		payload = result.Payload
	}

	callbackInput := PaymentCallbackInput{
		PaymentID:   payment.ID,
		ChannelID:   channel.ID,
		Status:      result.Status,
		ProviderRef: pickFirstNonEmpty(result.ProviderRef, payment.ProviderRef),
		Amount:      result.Amount,
		Currency:    strings.ToUpper(strings.TrimSpace(result.Currency)),
		PaidAt:      result.PaidAt,
		Payload:     payload,
	}

	updated, err := s.HandleCallback(callbackInput)
	if err != nil {
		log.Errorw("payment_webhook_callback_apply_failed",
			"payment_id", payment.ID,
			"order_no", result.OrderNo,
			"provider_ref", result.ProviderRef,
			"status", result.Status,
			"error", err,
		)
		return nil, result.Status, err
	}
	log.Infow("payment_webhook_processed",
		"payment_id", updated.ID,
		"order_no", result.OrderNo,
		"provider_ref", result.ProviderRef,
		"status", updated.Status,
	)
	return updated, result.Status, nil
}

// findWebhookPayment 通过 webhook result 反查 payment。
// 优先用 OrderNo（= GatewayOrderNo，商户单号），次选 ProviderRef（网关流水号）。
func (s *PaymentService) findWebhookPayment(channelID uint, result *provider.WebhookResult) (*models.Payment, error) {
	if result == nil {
		return nil, ErrPaymentNotFound
	}
	if orderNo := strings.TrimSpace(result.OrderNo); orderNo != "" {
		payment, err := s.paymentRepo.GetByGatewayOrderNo(orderNo)
		if err == nil && payment != nil && payment.ChannelID == channelID {
			return payment, nil
		}
	}
	if providerRef := strings.TrimSpace(result.ProviderRef); providerRef != "" {
		payment, err := s.paymentRepo.GetLatestByProviderRef(providerRef)
		if err == nil && payment != nil && payment.ChannelID == channelID {
			return payment, nil
		}
	}
	return nil, ErrPaymentNotFound
}

func buildPaypalCallbackAmount(event *paypal.WebhookEvent, status string) (models.Money, string, error) {
	amount := models.Money{}
	if event == nil {
		return amount, "", ErrPaymentGatewayResponseInvalid
	}

	amountValue, amountCurrency := event.CaptureAmount()
	amountValue = strings.TrimSpace(amountValue)
	amountCurrency = strings.ToUpper(strings.TrimSpace(amountCurrency))

	requiresAmount := strings.EqualFold(strings.TrimSpace(status), constants.PaymentStatusSuccess)
	if requiresAmount {
		if amountValue == "" || amountCurrency == "" {
			return amount, "", ErrPaymentGatewayResponseInvalid
		}
	}

	if amountValue == "" {
		if amountCurrency != "" {
			return amount, "", ErrPaymentGatewayResponseInvalid
		}
		return amount, "", nil
	}

	parsedAmount, err := decimal.NewFromString(amountValue)
	if err != nil || parsedAmount.Cmp(decimal.Zero) <= 0 {
		return amount, "", ErrPaymentGatewayResponseInvalid
	}
	if amountCurrency == "" {
		return amount, "", ErrPaymentGatewayResponseInvalid
	}

	amount = models.NewMoneyFromDecimal(parsedAmount)
	return amount, amountCurrency, nil
}

// HandleWechatWebhook 处理微信支付回调。
// P1.2c Task 6: 退化为 thin wrapper，通过 handleWebhookViaRegistry 路由解析。
func (s *PaymentService) HandleWechatWebhook(input WebhookCallbackInput) (*models.Payment, string, error) {
	return s.handleWebhookViaRegistry(
		input,
		constants.PaymentProviderOfficial,
		constants.PaymentChannelTypeWechat,
	)
}

// HandleStripeWebhook 处理 Stripe webhook。
// P1.2c Task 6: 退化为 thin wrapper，通过 handleWebhookViaRegistry 路由解析。
func (s *PaymentService) HandleStripeWebhook(input WebhookCallbackInput) (*models.Payment, string, error) {
	return s.handleWebhookViaRegistry(
		input,
		constants.PaymentProviderOfficial,
		constants.PaymentChannelTypeStripe,
	)
}

func mapWechatGatewayError(err error) error {
	switch {
	case errors.Is(err, wechatpay.ErrConfigInvalid):
		return ErrPaymentChannelConfigInvalid
	case errors.Is(err, wechatpay.ErrRequestFailed):
		return ErrPaymentGatewayRequestFailed
	case errors.Is(err, wechatpay.ErrSignatureInvalid), errors.Is(err, wechatpay.ErrResponseInvalid):
		return ErrPaymentGatewayResponseInvalid
	default:
		return ErrPaymentGatewayRequestFailed
	}
}

func mapStripeGatewayError(err error) error {
	switch {
	case errors.Is(err, stripe.ErrConfigInvalid):
		return ErrPaymentChannelConfigInvalid
	case errors.Is(err, stripe.ErrRequestFailed):
		return ErrPaymentGatewayRequestFailed
	case errors.Is(err, stripe.ErrSignatureInvalid), errors.Is(err, stripe.ErrResponseInvalid):
		return ErrPaymentGatewayResponseInvalid
	default:
		return ErrPaymentGatewayRequestFailed
	}
}
