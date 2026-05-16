package provider

import (
	"context"
	"errors"
	"testing"

	"github.com/dujiao-next/internal/constants"
	"github.com/dujiao-next/internal/models"
	"github.com/dujiao-next/internal/payment/epay"
)

func TestEpayAdapter_Type(t *testing.T) {
	a := NewEpayAdapter()
	want := constants.PaymentProviderEpay + ":"
	if got := a.Type(); got != want {
		t.Fatalf("Type() = %q, want %q", got, want)
	}
}

func TestEpayAdapter_ValidateConfig_UnsupportedChannel(t *testing.T) {
	a := NewEpayAdapter()
	err := a.ValidateConfig(models.JSON{}, "no-such-channel-type")
	if err == nil {
		t.Fatalf("expected error for unsupported channel")
	}
	if !errors.Is(err, ErrUnsupportedChannel) {
		t.Fatalf("expected wrapped ErrUnsupportedChannel, got %v", err)
	}
}

func TestEpayAdapter_CreatePayment_ConfigInvalidMapped(t *testing.T) {
	a := NewEpayAdapter()
	// 传一个 epay.IsSupportedChannelType 接受的 channelType(让校验过),
	// 但 config 空导致 ParseConfig/ValidateConfig 失败
	_, err := a.CreatePayment(context.Background(), models.JSON{}, CreateInput{
		OrderNo:     "ORDER_1",
		Currency:    "CNY",
		ChannelType: "alipay", // epay 支持 alipay/wxpay/qqpay
	})
	if err == nil {
		t.Fatalf("expected error from empty config")
	}
	if !errors.Is(err, ErrConfigInvalid) {
		t.Fatalf("expected wrapped ErrConfigInvalid, got %v", err)
	}
}

func TestEpayAdapter_MapEpayError(t *testing.T) {
	cases := []struct {
		name string
		in   error
		want error
	}{
		{"config", epay.ErrConfigInvalid, ErrConfigInvalid},
		{"channel_type→unsupported", epay.ErrChannelTypeNotOK, ErrUnsupportedChannel},
		{"sign_generate→config", epay.ErrSignatureGenerate, ErrConfigInvalid},
		{"sign_invalid→signature", epay.ErrSignatureInvalid, ErrSignatureInvalid},
		{"request", epay.ErrRequestFailed, ErrRequestFailed},
		{"response", epay.ErrResponseInvalid, ErrResponseInvalid},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := mapEpayError(tc.in)
			if !errors.Is(got, tc.want) {
				t.Fatalf("mapEpayError(%v) errors.Is %v = false, want true", tc.in, tc.want)
			}
		})
	}
}
