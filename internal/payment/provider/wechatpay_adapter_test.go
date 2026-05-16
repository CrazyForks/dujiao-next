package provider

import (
	"context"
	"errors"
	"testing"

	"github.com/dujiao-next/internal/constants"
	"github.com/dujiao-next/internal/models"
	"github.com/dujiao-next/internal/payment/wechatpay"
)

func TestWechatpayAdapter_Type(t *testing.T) {
	a := NewWechatpayAdapter()
	want := constants.PaymentProviderOfficial + ":" + constants.PaymentChannelTypeWechat
	if got := a.Type(); got != want {
		t.Fatalf("Type() = %q, want %q", got, want)
	}
}

func TestWechatpayAdapter_ValidateConfig_EmptyRejected(t *testing.T) {
	a := NewWechatpayAdapter()
	// 空 config 传给 ValidateConfig，会由 parseConfig 调 wechatpay.ValidateConfig("")
	// wechatpay 包的 ValidateConfig 会反对空字符串的 interaction_mode，返回 ErrConfigInvalid
	raw := models.JSON{}
	err := a.ValidateConfig(raw, "wechat")
	if err == nil {
		t.Fatalf("expected error for empty config")
	}
	if !errors.Is(err, ErrConfigInvalid) {
		t.Fatalf("expected wrapped ErrConfigInvalid, got %v", err)
	}
}

func TestWechatpayAdapter_CreatePayment_ConfigInvalidMapped(t *testing.T) {
	a := NewWechatpayAdapter()
	raw := models.JSON{} // 空 config
	_, err := a.CreatePayment(context.Background(), raw, CreateInput{
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

func TestWechatpayAdapter_MapWechatpayError(t *testing.T) {
	cases := []struct {
		name string
		in   error
		want error
	}{
		{"config", wechatpay.ErrConfigInvalid, ErrConfigInvalid},
		{"request", wechatpay.ErrRequestFailed, ErrRequestFailed},
		{"response", wechatpay.ErrResponseInvalid, ErrResponseInvalid},
		{"signature", wechatpay.ErrSignatureInvalid, ErrSignatureInvalid},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := mapWechatpayError(tc.in)
			if !errors.Is(got, tc.want) {
				t.Fatalf("mapWechatpayError(%v) errors.Is %v = false, want true", tc.in, tc.want)
			}
		})
	}
}
