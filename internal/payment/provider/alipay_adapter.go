package provider

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/dujiao-next/internal/constants"
	"github.com/dujiao-next/internal/models"
	"github.com/dujiao-next/internal/payment/alipay"

	"github.com/shopspring/decimal"
)

// alipayLocation 是 alipay 网关 gmt_payment 字段使用的固定时区(GMT+8)。
// alipay 回调的 gmt_payment 字段没有 timezone marker，必须用 ParseInLocation
// 显式指定；否则 time.Parse 默认 UTC，PaidAt 会偏 8 小时，导致下游对账、
// 订单完成时间标记、webhook payload 全错。
var alipayLocation = time.FixedZone("CST", 8*3600)

// alipayAdapter 是 alipay 网关的 Provider + CallbackVerifier 实现。
// alipay 没有主动查询 API，callback 是同步 form POST（不是 JSON webhook），
// 所以**不**实现 Capturer 和 Webhooker。
type alipayAdapter struct{}

// NewAlipayAdapter 实例化 alipay adapter。
func NewAlipayAdapter() Provider { return &alipayAdapter{} }

// 编译期断言 alipayAdapter 实现了 Provider 和 CallbackVerifier。
var (
	_ Provider         = (*alipayAdapter)(nil)
	_ CallbackVerifier = (*alipayAdapter)(nil)
)

// Type 返回 provider 标识。
func (a *alipayAdapter) Type() string {
	return constants.PaymentProviderOfficial + ":" + constants.PaymentChannelTypeAlipay
}

// parseConfig 解析并验证 alipay Config。interactionMode 影响 ValidateConfig
// 是否要求 return_url（jump 模式必填）。
func (a *alipayAdapter) parseConfig(raw models.JSON, interactionMode string) (*alipay.Config, error) {
	cfg, err := alipay.ParseConfig(raw)
	if err != nil {
		return nil, mapAlipayError(err)
	}
	if err := alipay.ValidateConfig(cfg, interactionMode); err != nil {
		return nil, mapAlipayError(err)
	}
	return cfg, nil
}

// ValidateConfig 验证 channel.ConfigJSON。
func (a *alipayAdapter) ValidateConfig(raw models.JSON, _ string) error {
	// 调用 parseConfig 传空字符串，由 alipay.ValidateConfig 反对 interaction_mode 空值，
	// wrapper 会正确地映射为 ErrConfigInvalid。
	_, err := a.parseConfig(raw, "")
	return err
}

// CreatePayment 创建支付。
func (a *alipayAdapter) CreatePayment(ctx context.Context, raw models.JSON, input CreateInput) (*CreateResult, error) {
	// 从 input.Extra 取 interaction_mode（jump / qr）。
	interactionMode, _ := input.Extra["interaction_mode"].(string)
	cfg, err := a.parseConfig(raw, interactionMode)
	if err != nil {
		return nil, err
	}

	native := alipay.CreateInput{
		OrderNo:   input.OrderNo,
		Amount:    input.Amount.Decimal.String(),
		Subject:   input.Subject,
		NotifyURL: input.NotifyURL,
		ReturnURL: input.ReturnURL,
	}
	result, err := alipay.CreatePayment(ctx, cfg, native, interactionMode)
	if err != nil {
		return nil, mapAlipayError(err)
	}

	return &CreateResult{
		ProviderRef: result.TradeNo,
		RedirectURL: result.PayURL,
		QRCodeURL:   result.QRCode,
		Payload:     models.JSON(result.Raw),
	}, nil
}

// VerifyCallback 实现 CallbackVerifier。alipay 用 form POST，body 参数忽略。
func (a *alipayAdapter) VerifyCallback(raw models.JSON, form map[string][]string, _ []byte) (*CallbackResult, error) {
	cfg, err := alipay.ParseConfig(raw)
	if err != nil {
		return nil, mapAlipayError(err)
	}
	// 注意：这里不调 alipay.ValidateConfig（因为没有 interactionMode 上下文），
	// 直接走 VerifyCallback，失败由 alipay 包内部抛 ErrSignatureInvalid。

	if err := alipay.VerifyCallback(cfg, form); err != nil {
		return nil, mapAlipayError(err)
	}

	orderNo := pickFormValue(form, "out_trade_no")
	providerRef := pickFormValue(form, "trade_no")
	tradeStatus := pickFormValue(form, "trade_status")
	amountStr := pickFormValue(form, "total_amount")
	paidAtRaw := pickFormValue(form, "gmt_payment")

	status := constants.PaymentStatusPending
	if tradeStatus == "TRADE_SUCCESS" || tradeStatus == "TRADE_FINISHED" {
		status = constants.PaymentStatusSuccess
	}

	// amount 解析失败时返回零值：wrapper 仅做适配，金额异常的语义边界由业务层判定。
	amount := models.Money{}
	if s := strings.TrimSpace(amountStr); s != "" {
		if d, parseErr := decimal.NewFromString(s); parseErr == nil {
			amount = models.NewMoneyFromDecimal(d)
		}
	}

	var paidAt *time.Time
	if t, parseErr := time.ParseInLocation("2006-01-02 15:04:05", strings.TrimSpace(paidAtRaw), alipayLocation); parseErr == nil {
		paidAt = &t
	}

	return &CallbackResult{
		OrderNo:     orderNo,
		ProviderRef: providerRef,
		Status:      status,
		Amount:      amount,
		Currency:    "CNY",
		PaidAt:      paidAt,
		Payload:     formToJSON(form),
	}, nil
}

// mapAlipayError 把 alipay 包的 sentinel error 映射为 provider 统一错误。
func mapAlipayError(err error) error {
	if err == nil {
		return nil
	}
	switch {
	case errors.Is(err, alipay.ErrConfigInvalid):
		return fmt.Errorf("%w: %v", ErrConfigInvalid, err)
	case errors.Is(err, alipay.ErrSignGenerate):
		// 签名生成失败 ≈ 配置错误（private_key 不可用）。
		return fmt.Errorf("%w: %v", ErrConfigInvalid, err)
	case errors.Is(err, alipay.ErrRequestFailed):
		return fmt.Errorf("%w: %v", ErrRequestFailed, err)
	case errors.Is(err, alipay.ErrResponseInvalid):
		return fmt.Errorf("%w: %v", ErrResponseInvalid, err)
	case errors.Is(err, alipay.ErrSignatureInvalid):
		return fmt.Errorf("%w: %v", ErrSignatureInvalid, err)
	default:
		return err
	}
}

// pickFormValue 返回 form 第一个非空值；若不存在或全空返回 ""。
// 同包内 sync-callback 类 adapter（alipay/epay/epusdt/bepusdt/tokenpay/okpay）共用。
func pickFormValue(form map[string][]string, key string) string {
	values, ok := form[key]
	if !ok || len(values) == 0 {
		return ""
	}
	return strings.TrimSpace(values[0])
}

// formToJSON 把 form 浅拷贝成 models.JSON（每 key 取首值）用于 Payload 字段。
// 同包内 sync-callback 类 adapter 共用。
func formToJSON(form map[string][]string) models.JSON {
	out := make(models.JSON, len(form))
	for k, v := range form {
		if len(v) == 0 {
			continue
		}
		out[k] = v[0]
	}
	return out
}
