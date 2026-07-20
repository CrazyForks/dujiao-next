package provider

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/shopspring/decimal"

	"github.com/dujiao-next/internal/constants"
	"github.com/dujiao-next/internal/models"
)

func TestDujiaoPayAdapter_Type(t *testing.T) {
	a := NewDujiaoPayAdapter()
	want := constants.PaymentProviderDujiaoPay + ":"
	if got := a.Type(); got != want {
		t.Fatalf("Type() = %q, want %q", got, want)
	}
}

func TestDujiaoPayAdapter_ValidateConfig_UnsupportedToken(t *testing.T) {
	a := NewDujiaoPayAdapter()
	err := a.ValidateConfig(models.JSON{
		"api_base_url":   "https://api.example.com",
		"api_key_id":     "key-1",
		"api_secret":     "secret-1",
		"webhook_secret": "whsec-1",
		"fiat_currency":  "USD",
	}, "doge-usdt")
	if err == nil {
		t.Fatalf("expected unsupported token error")
	}
	if !errors.Is(err, ErrUnsupportedChannel) {
		t.Fatalf("expected ErrUnsupportedChannel, got %v", err)
	}
}

func TestDujiaoPayAdapter_ValidateConfig_CashierMode(t *testing.T) {
	a := NewDujiaoPayAdapter()
	base := models.JSON{
		"api_base_url":    "https://api.example.com",
		"api_key_id":      "key-1",
		"api_secret":      "secret-1",
		"webhook_secret":  "whsec-1",
		"fiat_currency":   "USD",
		"order_mode":      "cashier",
		"allowed_methods": "tron-usdt,base-usdc",
	}

	if err := a.ValidateConfig(base, "dujiaopay"); err != nil {
		t.Fatalf("cashier channel_type dujiaopay should pass: %v", err)
	}
	if err := a.ValidateConfig(base, "tron-usdt"); !errors.Is(err, ErrUnsupportedChannel) {
		t.Fatalf("cashier with token channel_type should fail, got %v", err)
	}

	transaction := models.JSON{
		"api_base_url":   "https://api.example.com",
		"api_key_id":     "key-1",
		"api_secret":     "secret-1",
		"webhook_secret": "whsec-1",
		"fiat_currency":  "USD",
	}
	if err := a.ValidateConfig(transaction, "dujiaopay"); err == nil {
		t.Fatalf("transaction with channel_type dujiaopay should fail")
	}
	if err := a.ValidateConfig(transaction, "tron-usdt"); err != nil {
		t.Fatalf("transaction with token channel_type should pass: %v", err)
	}
}

func TestDujiaoPayAdapter_CreatePaymentCashierRejectsQRMode(t *testing.T) {
	a := NewDujiaoPayAdapter()
	_, err := a.CreatePayment(context.Background(), models.JSON{
		"api_base_url":   "https://api.example.com",
		"api_key_id":     "key-1",
		"api_secret":     "secret-1",
		"webhook_secret": "whsec-1",
		"fiat_currency":  "USD",
		"order_mode":     "cashier",
	}, CreateInput{
		OrderNo:     "PAY-2",
		Amount:      models.NewMoneyFromDecimal(decimal.RequireFromString("10")),
		Currency:    "USD",
		ChannelType: "dujiaopay",
		Extra:       models.JSON{"interaction_mode": constants.PaymentInteractionQR},
	})
	if !errors.Is(err, ErrConfigInvalid) {
		t.Fatalf("cashier + qr should fail with ErrConfigInvalid, got %v", err)
	}
}

func TestDujiaoPayAdapter_CreatePaymentCashierRedirect(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"order_id":"do_2","status":"awaiting_payment","selection_deadline":"2026-06-11T00:15:00Z","checkout_token":"ct_2","checkout_url":"https://pay.example.com/c/ct_2"}`))
	}))
	defer server.Close()

	a := NewDujiaoPayAdapter()
	result, err := a.CreatePayment(context.Background(), models.JSON{
		"api_base_url":    server.URL,
		"api_key_id":      "key-1",
		"api_secret":      "secret-1",
		"webhook_secret":  "whsec-1",
		"fiat_currency":   "USD",
		"order_mode":      "cashier",
		"allowed_methods": "tron-usdt,base-usdc",
	}, CreateInput{
		OrderNo:     "PAY-2",
		Amount:      models.NewMoneyFromDecimal(decimal.RequireFromString("10")),
		Currency:    "USD",
		ChannelType: "dujiaopay",
		Extra:       models.JSON{"interaction_mode": constants.PaymentInteractionRedirect},
	})
	if err != nil {
		t.Fatalf("CreatePayment failed: %v", err)
	}
	if result.ProviderRef != "do_2" {
		t.Fatalf("ProviderRef = %q, want do_2", result.ProviderRef)
	}
	if result.RedirectURL != "https://pay.example.com/c/ct_2" {
		t.Fatalf("RedirectURL = %q", result.RedirectURL)
	}
	if result.Payload["checkout_url"] != "https://pay.example.com/c/ct_2" {
		t.Fatalf("payload checkout_url = %v", result.Payload["checkout_url"])
	}
}

func TestDujiaoPayAdapter_CreatePaymentQRCodeModeUsesWalletAddress(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"order_id":"do_1","chain":"tron","token_id":"tron-usdt","checkout_url":"https://pay.example.com/c/ct_1","pay_address":"TAddr","payable_amount":"10.0001","status":"pending"}`))
	}))
	defer server.Close()

	a := NewDujiaoPayAdapter()
	result, err := a.CreatePayment(context.Background(), models.JSON{
		"api_base_url":   server.URL,
		"api_key_id":     "key-1",
		"api_secret":     "secret-1",
		"webhook_secret": "whsec-1",
		"fiat_currency":  "USD",
	}, CreateInput{
		OrderNo:        "PAY-1",
		Amount:         models.NewMoneyFromDecimal(decimal.RequireFromString("10")),
		Currency:       "USD",
		ChannelType:    "tron-usdt",
		ReturnURLQuery: map[string]string{"biz_type": "order", "order_no": "ORDER-1"},
		Extra:          models.JSON{"interaction_mode": constants.PaymentInteractionQR},
	})
	if err != nil {
		t.Fatalf("CreatePayment failed: %v", err)
	}
	if result.ProviderRef != "do_1" {
		t.Fatalf("ProviderRef = %q, want do_1", result.ProviderRef)
	}
	if result.RedirectURL != "https://pay.example.com/c/ct_1" {
		t.Fatalf("RedirectURL = %q", result.RedirectURL)
	}
	if result.QRCodeURL != "TAddr" {
		t.Fatalf("QRCodeURL = %q", result.QRCodeURL)
	}
	if result.Payload["pay_address"] != "TAddr" {
		t.Fatalf("payload pay_address = %v", result.Payload["pay_address"])
	}
	if result.Payload["chain"] != "tron" {
		t.Fatalf("payload chain = %v", result.Payload["chain"])
	}
	if result.Payload["token_id"] != "tron-usdt" {
		t.Fatalf("payload token_id = %v", result.Payload["token_id"])
	}
}
