package provider

import (
	"context"
	"errors"
	"testing"

	"github.com/dujiao-next/internal/constants"
	"github.com/dujiao-next/internal/models"
	"github.com/dujiao-next/internal/payment/paypal"
)

func TestPaypalAdapter_Type(t *testing.T) {
	a := NewPaypalAdapter()
	want := constants.PaymentProviderOfficial + ":" + constants.PaymentChannelTypePaypal
	if got := a.Type(); got != want {
		t.Fatalf("Type() = %q, want %q", got, want)
	}
}

func TestPaypalAdapter_ValidateConfig_EmptyRejected(t *testing.T) {
	a := NewPaypalAdapter()
	err := a.ValidateConfig(models.JSON{}, "")
	if err == nil {
		t.Fatalf("expected error from empty config")
	}
	if !errors.Is(err, ErrConfigInvalid) {
		t.Fatalf("expected wrapped ErrConfigInvalid, got %v", err)
	}
}

func TestPaypalAdapter_CreatePayment_ConfigInvalidMapped(t *testing.T) {
	a := NewPaypalAdapter()
	_, err := a.CreatePayment(context.Background(), models.JSON{}, CreateInput{
		OrderNo:  "ORDER_1",
		Currency: "USD",
	})
	if err == nil {
		t.Fatalf("expected error from empty config")
	}
	if !errors.Is(err, ErrConfigInvalid) {
		t.Fatalf("expected wrapped ErrConfigInvalid, got %v", err)
	}
}

func TestPaypalAdapter_MapPaypalError(t *testing.T) {
	cases := []struct {
		name string
		in   error
		want error
	}{
		{"config", paypal.ErrConfigInvalid, ErrConfigInvalid},
		{"auth", paypal.ErrAuthFailed, ErrAuthFailed},
		{"request", paypal.ErrRequestFailed, ErrRequestFailed},
		{"response", paypal.ErrResponseInvalid, ErrResponseInvalid},
		{"webhook→signature", paypal.ErrWebhookVerifyFailed, ErrSignatureInvalid},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := mapPaypalError(tc.in)
			if !errors.Is(got, tc.want) {
				t.Fatalf("mapPaypalError(%v) errors.Is %v = false, want true", tc.in, tc.want)
			}
		})
	}
}
