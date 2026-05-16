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
