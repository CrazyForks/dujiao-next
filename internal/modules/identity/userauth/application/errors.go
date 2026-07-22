package application

import "errors"

var (
	ErrNotFound                    = errors.New("user not found")
	ErrInvalidCredentials          = errors.New("invalid credentials")
	ErrInvalidPassword             = errors.New("invalid password")
	ErrEmailExists                 = errors.New("email exists")
	ErrEmailNotVerified            = errors.New("email not verified")
	ErrUserDisabled                = errors.New("user disabled")
	ErrInvalidEmail                = errors.New("invalid email")
	ErrInvalidVerifyPurpose        = errors.New("invalid verify purpose")
	ErrAgreementRequired           = errors.New("agreement required")
	ErrVerifyCodeInvalid           = errors.New("verify code invalid")
	ErrVerifyCodeExpired           = errors.New("verify code expired")
	ErrVerifyCodeTooFrequent       = errors.New("verify code too frequent")
	ErrVerifyCodeAttemptsExceeded  = errors.New("verify code attempts exceeded")
	ErrEmailServiceNotConfigured   = errors.New("email service not configured")
	ErrUserOAuthIdentityExists     = errors.New("user oauth identity exists")
	ErrUserOAuthAlreadyBound       = errors.New("user oauth already bound")
	ErrUserOAuthNotBound           = errors.New("user oauth not bound")
	ErrTelegramUnbindRequiresEmail = errors.New("telegram unbind requires real email")
	ErrProfileEmpty                = errors.New("profile empty")
	ErrEmailChangeInvalid          = errors.New("email change invalid")
	ErrEmailChangeExists           = errors.New("email change exists")
	ErrRegistrationDisabled        = errors.New("registration disabled")
)
