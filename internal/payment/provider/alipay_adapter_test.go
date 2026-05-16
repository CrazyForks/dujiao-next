package provider

import (
	"context"
	"errors"
	"testing"

	"github.com/dujiao-next/internal/constants"
	"github.com/dujiao-next/internal/models"
	"github.com/dujiao-next/internal/payment/alipay"
)

func TestAlipayAdapter_Type(t *testing.T) {
	a := NewAlipayAdapter()
	want := constants.PaymentProviderOfficial + ":" + constants.PaymentChannelTypeAlipay
	if got := a.Type(); got != want {
		t.Fatalf("Type() = %q, want %q", got, want)
	}
}

func TestAlipayAdapter_ValidateConfig_EmptyRejected(t *testing.T) {
	a := NewAlipayAdapter()
	err := a.ValidateConfig(models.JSON{}, "")
	if err == nil {
		t.Fatalf("expected error from empty config")
	}
	if !errors.Is(err, ErrConfigInvalid) {
		t.Fatalf("expected wrapped ErrConfigInvalid, got %v", err)
	}
}

func TestAlipayAdapter_CreatePayment_ConfigInvalidMapped(t *testing.T) {
	a := NewAlipayAdapter()
	_, err := a.CreatePayment(context.Background(), models.JSON{}, CreateInput{
		OrderNo:  "ORDER_1",
		Currency: "CNY",
	})
	if err == nil {
		t.Fatalf("expected error from empty config")
	}
	if !errors.Is(err, ErrConfigInvalid) {
		t.Fatalf("expected wrapped ErrConfigInvalid, got %v", err)
	}
}

// TestAlipayAdapter_ValidateConfig_ValidConfig_C3Regression 守护 C3 regression fix:
// ValidateConfig 不再将 interactionMode 参数丢弃传空字符串，而是正确使用传入的 interactionMode。
// service 层 official provider 分支传入 channel.InteractionMode；为空时用 QR 作 default。
// valid alipay config + 合法 interactionMode 必须通过校验，不能因 interaction_mode 为空/无效而被拒绝。
func TestAlipayAdapter_ValidateConfig_ValidConfig_C3Regression(t *testing.T) {
	a := NewAlipayAdapter()
	// 使用 alipay native test 中确认有效的最小配置（QR 模式不要求 return_url）
	raw := models.JSON{
		"app_id":            "2026000000000000",
		"private_key":       "k",
		"alipay_public_key": "p",
		"gateway_url":       "https://openapi.alipay.com/gateway.do",
		"notify_url":        "https://example.com/api/v1/payments/callback",
		"sign_type":         "rsa2",
	}
	// 传 interactionMode=qr（service 层会从 channel.InteractionMode 取值传入）
	err := a.ValidateConfig(raw, constants.PaymentInteractionQR)
	if err != nil {
		t.Fatalf("ValidateConfig() should pass valid alipay config with QR mode, got: %v", err)
	}
}

// TestAlipayAdapter_ValidateConfig_EmptyInteractionModeUsesDefault 验证修复后
// 外部传空 interactionMode 不再让 ValidateConfig 永远失败（C3 修复前的 bug 路径）。
// 原因：wrapper 内部当 interactionMode="" 时用 QR 作 default。
func TestAlipayAdapter_ValidateConfig_EmptyInteractionModeUsesDefault(t *testing.T) {
	a := NewAlipayAdapter()
	raw := models.JSON{
		"app_id":            "2026000000000000",
		"private_key":       "k",
		"alipay_public_key": "p",
		"gateway_url":       "https://openapi.alipay.com/gateway.do",
		"notify_url":        "https://example.com/api/v1/payments/callback",
		"sign_type":         "rsa2",
	}
	// 第二参数传空字符串：
	// C3 修复前，这会把 "" 传给 alipay.ValidateConfig 导致 ErrConfigInvalid；
	// C3 修复后，wrapper 用 QR 作 default，valid config 应当通过。
	err := a.ValidateConfig(raw, "")
	if err != nil {
		t.Fatalf("ValidateConfig() should use QR default when interactionMode is empty, got: %v", err)
	}
}

func TestAlipayAdapter_MapAlipayError(t *testing.T) {
	cases := []struct {
		name string
		in   error
		want error
	}{
		{"config", alipay.ErrConfigInvalid, ErrConfigInvalid},
		{"sign_generate→config", alipay.ErrSignGenerate, ErrConfigInvalid},
		{"request", alipay.ErrRequestFailed, ErrRequestFailed},
		{"response", alipay.ErrResponseInvalid, ErrResponseInvalid},
		{"signature", alipay.ErrSignatureInvalid, ErrSignatureInvalid},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := mapAlipayError(tc.in)
			if !errors.Is(got, tc.want) {
				t.Fatalf("mapAlipayError(%v) errors.Is %v = false, want true", tc.in, tc.want)
			}
		})
	}
}
