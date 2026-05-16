package provider

import (
	"context"
	"errors"
	"testing"

	"github.com/dujiao-next/internal/constants"
	"github.com/dujiao-next/internal/models"
	"github.com/dujiao-next/internal/payment/stripe"
)

func TestStripeAdapter_Type(t *testing.T) {
	a := NewStripeAdapter()
	want := constants.PaymentProviderOfficial + ":" + constants.PaymentChannelTypeStripe
	if got := a.Type(); got != want {
		t.Fatalf("Type() = %q, want %q", got, want)
	}
}

func TestStripeAdapter_ValidateConfig_InvalidIsMapped(t *testing.T) {
	a := NewStripeAdapter()
	// 缺 secret_key，应被 stripe.ValidateConfig 拒绝
	raw := models.JSON{
		"webhook_secret":       "whsec_x",
		"success_url":          "https://example.com/s",
		"cancel_url":           "https://example.com/c",
		"api_base_url":         "https://api.stripe.com",
		"payment_method_types": []any{"card"},
	}
	err := a.ValidateConfig(raw, "stripe")
	if err == nil {
		t.Fatalf("expected error for missing secret_key")
	}
	if !errors.Is(err, ErrConfigInvalid) {
		t.Fatalf("expected wrapped ErrConfigInvalid, got %v", err)
	}
}

func TestStripeAdapter_CreatePayment_ConfigInvalidMapped(t *testing.T) {
	a := NewStripeAdapter()
	raw := models.JSON{} // 空 config
	_, err := a.CreatePayment(context.Background(), raw, CreateInput{
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

func TestStripeAdapter_MapStripeError(t *testing.T) {
	cases := []struct {
		name string
		in   error
		want error
	}{
		{"config", stripe.ErrConfigInvalid, ErrConfigInvalid},
		{"request", stripe.ErrRequestFailed, ErrRequestFailed},
		{"response", stripe.ErrResponseInvalid, ErrResponseInvalid},
		{"signature", stripe.ErrSignatureInvalid, ErrSignatureInvalid},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := mapStripeError(tc.in)
			if !errors.Is(got, tc.want) {
				t.Fatalf("mapStripeError(%v) errors.Is %v = false, want true", tc.in, tc.want)
			}
		})
	}
}
