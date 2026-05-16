package provider

import (
	"context"
	"errors"
	"testing"

	"github.com/dujiao-next/internal/constants"
	"github.com/dujiao-next/internal/models"
	"github.com/dujiao-next/internal/payment/okpay"
)

func TestOkpayAdapter_Type(t *testing.T) {
	a := NewOkpayAdapter()
	want := constants.PaymentProviderOkpay + ":"
	if got := a.Type(); got != want {
		t.Fatalf("Type() = %q, want %q", got, want)
	}
}

func TestOkpayAdapter_ValidateConfig_UnsupportedChannel(t *testing.T) {
	a := NewOkpayAdapter()
	err := a.ValidateConfig(models.JSON{}, "no-such-channel-type")
	if err == nil {
		t.Fatalf("expected error for unsupported channel")
	}
	if !errors.Is(err, ErrUnsupportedChannel) {
		t.Fatalf("expected wrapped ErrUnsupportedChannel, got %v", err)
	}
}

func TestOkpayAdapter_CreatePayment_ConfigInvalidMapped(t *testing.T) {
	a := NewOkpayAdapter()
	// 用 okpay 真实支持的 channelType("usdt" / "trx")
	_, err := a.CreatePayment(context.Background(), models.JSON{}, CreateInput{
		OrderNo:     "ORDER_1",
		Currency:    "USDT",
		ChannelType: "usdt",
	})
	if err == nil {
		t.Fatalf("expected error from empty config")
	}
	if !errors.Is(err, ErrConfigInvalid) {
		t.Fatalf("expected wrapped ErrConfigInvalid, got %v", err)
	}
}

func TestOkpayAdapter_MapOkpayError(t *testing.T) {
	cases := []struct {
		name string
		in   error
		want error
	}{
		{"config", okpay.ErrConfigInvalid, ErrConfigInvalid},
		{"request", okpay.ErrRequestFailed, ErrRequestFailed},
		{"response", okpay.ErrResponseInvalid, ErrResponseInvalid},
		{"signature", okpay.ErrSignatureInvalid, ErrSignatureInvalid},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := mapOkpayError(tc.in)
			if !errors.Is(got, tc.want) {
				t.Fatalf("mapOkpayError(%v) errors.Is %v = false, want true", tc.in, tc.want)
			}
		})
	}
}
