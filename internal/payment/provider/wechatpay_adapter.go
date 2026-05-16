package provider

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/dujiao-next/internal/constants"
	"github.com/dujiao-next/internal/models"
	"github.com/dujiao-next/internal/payment/wechatpay"

	"github.com/shopspring/decimal"
)

// wechatpayAdapter 是 wechatpay 网关的 Provider/Capturer/Webhooker 实现。
// 它仅做参数适配 + 错误映射，所有业务逻辑仍委托给 internal/payment/wechatpay/ 包级函数。
type wechatpayAdapter struct{}

// NewWechatpayAdapter 实例化 wechatpay adapter。
func NewWechatpayAdapter() Provider { return &wechatpayAdapter{} }

// 编译期断言 wechatpayAdapter 实现了三个 capability interface。
var (
	_ Provider  = (*wechatpayAdapter)(nil)
	_ Capturer  = (*wechatpayAdapter)(nil)
	_ Webhooker = (*wechatpayAdapter)(nil)
)

// Type 返回 provider 标识。
func (a *wechatpayAdapter) Type() string {
	return constants.PaymentProviderOfficial + ":" + constants.PaymentChannelTypeWechat
}

// parseConfig 解析并验证 wechatpay Config，把 wechatpay.ErrConfigInvalid 等映射为 provider.ErrXxx。
// 4 个公开方法共用，避免每个都重复样板。
// interactionMode 参数用于调用 wechatpay.ValidateConfig。当为空字符串时，wechatpay 包会校验并拒绝
// ("interaction_mode is not supported")，wrapper 会正确地映射为 ErrConfigInvalid。
func (a *wechatpayAdapter) parseConfig(raw models.JSON, interactionMode string) (*wechatpay.Config, error) {
	cfg, err := wechatpay.ParseConfig(raw)
	if err != nil {
		return nil, mapWechatpayError(err)
	}
	if err := wechatpay.ValidateConfig(cfg, interactionMode); err != nil {
		return nil, mapWechatpayError(err)
	}
	return cfg, nil
}

// ValidateConfig 验证 channel.ConfigJSON。
// 调用 parseConfig 传空字符串，由 wechatpay.ValidateConfig 反对 interaction_mode 空值
// 并正确地映射为 ErrConfigInvalid。
func (a *wechatpayAdapter) ValidateConfig(raw models.JSON, _ string) error {
	_, err := a.parseConfig(raw, "")
	return err
}

// CreatePayment 创建支付。
func (a *wechatpayAdapter) CreatePayment(ctx context.Context, raw models.JSON, input CreateInput) (*CreateResult, error) {
	// 从 input.Extra 取 interaction_mode（jsapi/native/h5）
	interactionMode, _ := input.Extra["interaction_mode"].(string)
	cfg, err := a.parseConfig(raw, interactionMode)
	if err != nil {
		return nil, err
	}

	native := wechatpay.CreateInput{
		OrderNo:     input.OrderNo,
		Amount:      input.Amount.Decimal.String(),
		Currency:    input.Currency,
		Description: input.Subject,
		ClientIP:    input.ClientIP,
		NotifyURL:   input.NotifyURL,
	}
	result, err := wechatpay.CreatePayment(ctx, cfg, native, interactionMode)
	if err != nil {
		return nil, mapWechatpayError(err)
	}

	// wechat CreatePayment 阶段返回 PrepayID，但不是最终的 transaction_id。
	// 最终 transaction_id 在 Query 或 Webhook 时才出现。所以 ProviderRef 设为空，
	// PrepayID 和 PayURL/QRCode 入 Payload 供上游参考。
	return &CreateResult{
		ProviderRef: "",
		RedirectURL: result.PayURL,
		QRCodeURL:   result.QRCode,
		Payload: models.JSON(map[string]interface{}{
			"prepay_id": result.PrepayID,
			"raw":       result.Raw,
		}),
	}, nil
}

// QueryPayment 主动查询订单状态(实现 Capturer)。
// wechat 的 QueryOrderByOutTradeNo 用商户订单号查询，返回 TransactionID（wechat 的 transaction_id）。
// 调用方传入的 providerRef 实际就是 OrderNo（因为 CreatePayment 阶段没有 transaction_id）。
func (a *wechatpayAdapter) QueryPayment(ctx context.Context, raw models.JSON, providerRef string) (*QueryResult, error) {
	cfg, err := a.parseConfig(raw, "")
	if err != nil {
		return nil, err
	}

	result, err := wechatpay.QueryOrderByOutTradeNo(ctx, cfg, providerRef)
	if err != nil {
		return nil, mapWechatpayError(err)
	}

	// amount 解析失败时返回零值：wrapper 仅做适配，金额异常的语义边界(对账失败 / 网关返回脏数据)
	// 留给上游业务层判定，wrapper 不擅自报错。
	amount := models.Money{}
	if s := strings.TrimSpace(result.Amount); s != "" {
		if parsed, parseErr := decimal.NewFromString(s); parseErr == nil {
			amount = models.NewMoneyFromDecimal(parsed)
		}
	}

	return &QueryResult{
		ProviderRef: result.TransactionID,
		Status:      result.Status,
		Amount:      amount,
		Currency:    strings.ToUpper(strings.TrimSpace(result.Currency)),
		PaidAt:      result.PaidAt,
		Payload:     models.JSON(result.Raw),
	}, nil
}

// ParseWebhook 验签并解析 webhook(实现 Webhooker)。
func (a *wechatpayAdapter) ParseWebhook(ctx context.Context, raw models.JSON, headers map[string]string, body []byte, _ time.Time) (*WebhookResult, error) {
	cfg, err := a.parseConfig(raw, "")
	if err != nil {
		return nil, err
	}

	result, err := wechatpay.VerifyAndDecodeWebhook(ctx, cfg, headers, body)
	if err != nil {
		return nil, mapWechatpayError(err)
	}

	// amount 解析失败时返回零值：wrapper 仅做适配，金额异常的语义边界(对账失败 / 网关返回脏数据)
	// 留给上游业务层判定，wrapper 不擅自报错。
	amount := models.Money{}
	if s := strings.TrimSpace(result.Amount); s != "" {
		if parsed, parseErr := decimal.NewFromString(s); parseErr == nil {
			amount = models.NewMoneyFromDecimal(parsed)
		}
	}

	return &WebhookResult{
		OrderNo:     result.OrderNo,
		ProviderRef: result.TransactionID,
		Status:      result.Status,
		Amount:      amount,
		Currency:    strings.ToUpper(strings.TrimSpace(result.Currency)),
		PaidAt:      result.PaidAt,
		Payload:     models.JSON(result.Raw),
	}, nil
}

func mapWechatpayError(err error) error {
	if err == nil {
		return nil
	}
	switch {
	case errors.Is(err, wechatpay.ErrConfigInvalid):
		return fmt.Errorf("%w: %v", ErrConfigInvalid, err)
	case errors.Is(err, wechatpay.ErrRequestFailed):
		return fmt.Errorf("%w: %v", ErrRequestFailed, err)
	case errors.Is(err, wechatpay.ErrResponseInvalid):
		return fmt.Errorf("%w: %v", ErrResponseInvalid, err)
	case errors.Is(err, wechatpay.ErrSignatureInvalid):
		return fmt.Errorf("%w: %v", ErrSignatureInvalid, err)
	default:
		return err
	}
}
