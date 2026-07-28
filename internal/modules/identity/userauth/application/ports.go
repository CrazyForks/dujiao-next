package application

import "github.com/dujiao-next/internal/shared/mailbrand"

// VerificationEmailSender defines the only email capability required by user
// authentication. Infrastructure implementations are supplied by composition.
type VerificationEmailSender interface {
	SendVerifyCode(toEmail, code, purpose, locale string, brand mailbrand.Brand) error
}
