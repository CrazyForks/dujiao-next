package service

import (
	"errors"
	"strings"
	"testing"

	"github.com/dujiao-next/internal/i18n"
	"github.com/dujiao-next/internal/models"

	"github.com/shopspring/decimal"
)

func TestBuildOrderStatusContent(t *testing.T) {
	tests := []struct {
		name                string
		locale              string
		status              string
		payload             string
		wantSubjectContains []string
		wantBodyContains    []string
	}{
		{
			name:   "paid_zh",
			locale: i18n.LocaleZH,
			status: "paid",
			wantSubjectContains: []string{
				"订单状态更新",
				"已支付",
			},
			wantBodyContains: []string{
				"已收到您的付款",
				"订单号：DJ-PAID",
			},
		},
		{
			name:   "canceled_en",
			locale: i18n.LocaleEN,
			status: "canceled",
			wantSubjectContains: []string{
				"Order status updated",
				"Canceled",
			},
			wantBodyContains: []string{
				"The order has been canceled",
				"Order No: DJ-CANCEL",
			},
		},
		{
			name:    "delivered_with_payload_tw",
			locale:  i18n.LocaleTW,
			status:  "delivered",
			payload: "CODE-A\nCODE-B",
			wantSubjectContains: []string{
				"訂單狀態更新",
				"已交付",
			},
			wantBodyContains: []string{
				"交付內容",
				"CODE-A",
			},
		},
		{
			name:    "delivered_no_payload_en",
			locale:  i18n.LocaleEN,
			status:  "delivered",
			payload: "",
			wantSubjectContains: []string{
				"Order status updated",
				"Delivered",
			},
			wantBodyContains: []string{
				"Delivery completed",
				"Order No: DJ-DELIVER",
			},
		},
		{
			name:    "completed_with_payload_zh",
			locale:  i18n.LocaleZH,
			status:  "completed",
			payload: "AUTO-CODE-001",
			wantSubjectContains: []string{
				"订单状态更新",
				"已完成",
			},
			wantBodyContains: []string{
				"交付内容",
				"AUTO-CODE-001",
			},
		},
		{
			name:   "refunded_zh",
			locale: i18n.LocaleZH,
			status: "refunded",
			wantSubjectContains: []string{
				"订单状态更新",
				"已退款",
			},
			wantBodyContains: []string{
				"退款金额：8.80 USD",
				"退款原因：manual refund",
				"示例站点 的网址：https://example.com",
			},
		},
		{
			name:   "partially_refunded_en",
			locale: i18n.LocaleEN,
			status: "partially_refunded",
			wantSubjectContains: []string{
				"Order status updated",
				"Partially refunded",
			},
			wantBodyContains: []string{
				"Refund Amount: 8.80 USD",
				"Reason for refund: manual refund",
				"Example Site's Site URL: https://example.com",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := OrderStatusEmailInput{
				OrderNo:         pickOrderNo(tt.status),
				Status:          tt.status,
				Amount:          models.NewMoneyFromDecimal(decimal.NewFromFloat(19.8)),
				RefundAmount:    models.NewMoneyFromDecimal(decimal.NewFromFloat(8.8)),
				RefundReason:    "manual refund",
				Currency:        "USD",
				SiteName:        "Example Site",
				SiteURL:         "https://example.com",
				FulfillmentInfo: tt.payload,
			}
			if tt.locale == i18n.LocaleZH {
				input.SiteName = "示例站点"
			}
			subject, body := buildOrderStatusContent(input, tt.locale)
			for _, expected := range tt.wantSubjectContains {
				if !strings.Contains(subject, expected) {
					t.Fatalf("subject missing %q: %s", expected, subject)
				}
			}
			for _, expected := range tt.wantBodyContains {
				if !strings.Contains(body, expected) {
					t.Fatalf("body missing %q: %s", expected, body)
				}
			}
			if strings.Contains(body, "%!") {
				t.Fatalf("body contains fmt placeholder error marker: %s", body)
			}
		})
	}
}

func pickOrderNo(status string) string {
	switch status {
	case "paid":
		return "DJ-PAID"
	case "canceled":
		return "DJ-CANCEL"
	case "refunded":
		return "DJ-REFUND"
	case "partially_refunded":
		return "DJ-PART-REFUND"
	default:
		return "DJ-DELIVER"
	}
}

func TestIsEmailRecipientRejected(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "smtp_550_no_such_recipient",
			err:  errors.New("550 No such recipient here"),
			want: true,
		},
		{
			name: "smtp_user_unknown",
			err:  errors.New("SMTP 5.1.1 user unknown"),
			want: true,
		},
		{
			name: "smtp_550_mailbox_unavailable",
			err:  errors.New("550 mailbox unavailable"),
			want: true,
		},
		{
			name: "network_timeout",
			err:  errors.New("dial tcp timeout"),
			want: false,
		},
		{
			name: "nil_error",
			err:  nil,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isEmailRecipientRejected(tt.err); got != tt.want {
				t.Fatalf("isEmailRecipientRejected() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNormalizeEmailSendError(t *testing.T) {
	rejected := errors.New("550 No such recipient here")
	if got := normalizeEmailSendError(rejected); !errors.Is(got, ErrEmailRecipientRejected) {
		t.Fatalf("normalizeEmailSendError() expected ErrEmailRecipientRejected, got %v", got)
	}

	networkErr := errors.New("dial tcp timeout")
	if got := normalizeEmailSendError(networkErr); !errors.Is(got, networkErr) {
		t.Fatalf("normalizeEmailSendError() should keep original error, got %v", got)
	}

	if got := normalizeEmailSendError(nil); got != nil {
		t.Fatalf("normalizeEmailSendError(nil) should be nil, got %v", got)
	}
}

func TestSendTextEmailSkipTelegramPlaceholder(t *testing.T) {
	service := &EmailService{}
	if err := service.sendTextEmail("telegram_6059928735@login.local", "subject", "body"); err != nil {
		t.Fatalf("sendTextEmail() should skip telegram placeholder email, got %v", err)
	}
}

func TestBuildOrderStatusContentFromTemplateIncludesSiteBrand(t *testing.T) {
	tmpl := OrderEmailTemplateDefaultSetting()
	tmpl.Templates.Paid.ZHCN.Subject = "订单通知 {{site_name}}"
	tmpl.Templates.Paid.ZHCN.Body = "订单号：{{order_no}}\n站点：{{site_name}} {{site_url}}"

	input := OrderStatusEmailInput{
		OrderNo:         "DJ-SITE-001",
		Status:          "paid",
		Amount:          models.NewMoneyFromDecimal(decimal.NewFromInt(10)),
		Currency:        "CNY",
		SiteName:        " 示例站点 ",
		SiteURL:         " https://example.com/shop ",
		IsGuest:         false,
		FulfillmentInfo: "",
	}

	subject, body := buildOrderStatusContentFromTemplate(input, i18n.LocaleZH, tmpl)

	if !strings.Contains(subject, "示例站点") {
		t.Fatalf("subject should contain site_name, got: %s", subject)
	}
	if !strings.Contains(body, "站点：示例站点 https://example.com/shop") {
		t.Fatalf("body should contain site_name and site_url, got: %s", body)
	}
}
