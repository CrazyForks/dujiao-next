package userauthhttp

import (
	"errors"
	"strings"

	"github.com/dujiao-next/internal/constants"
	"github.com/dujiao-next/internal/http/handlers/shared"
	"github.com/dujiao-next/internal/http/response"
	"github.com/dujiao-next/internal/i18n"
	"github.com/dujiao-next/internal/modules/captcha"

	"github.com/gin-gonic/gin"
)

var (
	ErrInvalidVerifyPurpose  = errors.New("invalid verify purpose")
	ErrEmailExists           = errors.New("email exists")
	ErrEmailDomainNotAllowed = errors.New("email domain not allowed")
)

// CaptchaVerifier 是发送验证码所需的验证码端口。
type CaptchaVerifier interface {
	Verify(scene string, payload shared.CaptchaPayloadRequest, clientIP string) error
}

// UserVerifySettings 是发送验证码所需的设置端口。
type UserVerifySettings interface {
	GetEmailVerificationEnabled(defaultValue bool) (bool, error)
	GetRegistrationEnabled(defaultValue bool) (bool, error)
}

// UserVerifyAuth 是发送邮箱验证码端口。
type UserVerifyAuth interface {
	SendVerifyCode(email, purpose, locale string) error
}

// UserVerifyHandler 处理公开的发送邮箱验证码 HTTP 请求。
type UserVerifyHandler struct {
	settings UserVerifySettings
	captcha  CaptchaVerifier
	auth     UserVerifyAuth
}

func NewUserVerifyHandler(settings UserVerifySettings, captcha CaptchaVerifier, auth UserVerifyAuth) *UserVerifyHandler {
	if settings == nil {
		panic("user verify handler: settings is nil")
	}
	if auth == nil {
		panic("user verify handler: auth is nil")
	}
	return &UserVerifyHandler{settings: settings, captcha: captcha, auth: auth}
}

// UserSendVerifyCodeRequest 发送验证码请求。
type UserSendVerifyCodeRequest struct {
	Email          string                       `json:"email" binding:"required"`
	Purpose        string                       `json:"purpose" binding:"required"`
	CaptchaPayload shared.CaptchaPayloadRequest `json:"captcha_payload"`
}

// SendUserVerifyCode 发送用户邮箱验证码。
func (h *UserVerifyHandler) SendUserVerifyCode(c *gin.Context) {
	var req UserSendVerifyCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		shared.RespondBindError(c, err)
		return
	}

	purpose := strings.ToLower(strings.TrimSpace(req.Purpose))

	emailVerificationEnabled, err := h.settings.GetEmailVerificationEnabled(true)
	if err != nil {
		shared.RespondError(c, response.CodeInternal, "error.send_verify_code_failed", err)
		return
	}
	if !emailVerificationEnabled {
		shared.RespondError(c, response.CodeForbidden, "error.email_verification_disabled", nil)
		return
	}

	if purpose == constants.VerifyPurposeRegister {
		registrationEnabled, err := h.settings.GetRegistrationEnabled(true)
		if err != nil {
			shared.RespondError(c, response.CodeInternal, "error.send_verify_code_failed", err)
			return
		}
		if !registrationEnabled {
			shared.RespondError(c, response.CodeForbidden, "error.registration_disabled", nil)
			return
		}
	}

	captchaScene := ""
	switch purpose {
	case constants.VerifyPurposeRegister:
		captchaScene = constants.CaptchaSceneRegisterSendCode
	case constants.VerifyPurposeReset:
		captchaScene = constants.CaptchaSceneResetSendCode
	}
	if captchaScene != "" && h.captcha != nil {
		if captchaErr := h.captcha.Verify(captchaScene, req.CaptchaPayload, c.ClientIP()); captchaErr != nil {
			respondCaptchaError(c, captchaErr)
			return
		}
	}

	locale := i18n.ResolveLocale(c)
	if err := h.auth.SendVerifyCode(req.Email, req.Purpose, locale); err != nil {
		switch {
		case errors.Is(err, ErrInvalidEmail):
			shared.RespondError(c, response.CodeBadRequest, "error.email_invalid", nil)
		case errors.Is(err, ErrInvalidVerifyPurpose):
			shared.RespondError(c, response.CodeBadRequest, "error.verify_purpose_invalid", nil)
		case errors.Is(err, ErrEmailExists):
			shared.RespondError(c, response.CodeBadRequest, "error.email_exists", nil)
		case errors.Is(err, ErrEmailDomainNotAllowed):
			shared.RespondError(c, response.CodeBadRequest, "error.email_domain_not_allowed", nil)
		case errors.Is(err, ErrUserNotFound):
			shared.RespondError(c, response.CodeNotFound, "error.user_not_found", nil)
		case errors.Is(err, ErrVerifyCodeTooFrequent):
			shared.RespondError(c, response.CodeTooManyRequests, "error.verify_code_too_frequent", nil)
		case errors.Is(err, ErrEmailRecipientRejected):
			shared.RespondError(c, response.CodeBadRequest, "error.email_recipient_not_found", nil)
		case errors.Is(err, ErrEmailServiceDisabled),
			errors.Is(err, ErrEmailServiceNotConfigured):
			shared.RespondError(c, response.CodeInternal, "error.email_service_not_configured", err)
		default:
			shared.RespondError(c, response.CodeInternal, "error.send_verify_code_failed", err)
		}
		return
	}

	response.Success(c, gin.H{"sent": true})
}

func respondCaptchaError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, captcha.ErrRequired):
		shared.RespondError(c, response.CodeBadRequest, "error.captcha_required", nil)
	case errors.Is(err, captcha.ErrInvalid):
		shared.RespondError(c, response.CodeBadRequest, "error.captcha_invalid", nil)
	case errors.Is(err, captcha.ErrConfigInvalid):
		shared.RespondError(c, response.CodeInternal, "error.captcha_config_invalid", err)
	default:
		shared.RespondError(c, response.CodeInternal, "error.captcha_verify_failed", err)
	}
}
