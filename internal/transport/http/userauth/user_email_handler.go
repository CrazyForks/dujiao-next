package userauthhttp

import (
	"errors"

	"github.com/dujiao-next/internal/dto"
	"github.com/dujiao-next/internal/http/handlers/shared"
	"github.com/dujiao-next/internal/http/response"
	"github.com/dujiao-next/internal/i18n"
	"github.com/dujiao-next/internal/models"

	"github.com/gin-gonic/gin"
)

var (
	ErrInvalidEmail               = errors.New("invalid email")
	ErrEmailChangeInvalid         = errors.New("email change invalid")
	ErrEmailChangeExists          = errors.New("email change exists")
	ErrVerifyCodeInvalid          = errors.New("verify code invalid")
	ErrVerifyCodeExpired          = errors.New("verify code expired")
	ErrVerifyCodeTooFrequent      = errors.New("verify code too frequent")
	ErrVerifyCodeAttemptsExceeded = errors.New("verify code attempts exceeded")
	ErrEmailServiceDisabled       = errors.New("email service disabled")
	ErrEmailServiceNotConfigured  = errors.New("email service not configured")
	ErrEmailRecipientRejected     = errors.New("email recipient rejected")
)

// UserEmailService 是更换邮箱端点所需的最小端口。
type UserEmailService interface {
	SendChangeEmailCode(userID uint, kind, newEmail, locale string) error
	ChangeEmail(userID uint, newEmail, oldCode, newCode string) (*models.User, error)
	ResolveEmailChangeMode(user *models.User) (string, error)
	ResolvePasswordChangeMode(user *models.User) (string, error)
}

// UserEmailHandler 处理当前用户更换邮箱 HTTP 请求。
type UserEmailHandler struct {
	service UserEmailService
}

func NewUserEmailHandler(service UserEmailService) *UserEmailHandler {
	if service == nil {
		panic("user email handler: service is nil")
	}
	return &UserEmailHandler{service: service}
}

// ChangeEmailSendCodeRequest 更换邮箱验证码请求。
type ChangeEmailSendCodeRequest struct {
	Kind     string `json:"kind" binding:"required"`
	NewEmail string `json:"new_email"`
}

// SendChangeEmailCode 发送更换邮箱验证码。
func (h *UserEmailHandler) SendChangeEmailCode(c *gin.Context) {
	id, ok := shared.GetUserID(c)
	if !ok {
		return
	}

	var req ChangeEmailSendCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		shared.RespondBindError(c, err)
		return
	}

	locale := i18n.ResolveLocale(c)
	if err := h.service.SendChangeEmailCode(id, req.Kind, req.NewEmail, locale); err != nil {
		switch {
		case errors.Is(err, ErrInvalidEmail):
			shared.RespondError(c, response.CodeBadRequest, "error.email_invalid", nil)
		case errors.Is(err, ErrEmailChangeInvalid):
			shared.RespondError(c, response.CodeBadRequest, "error.email_change_invalid", nil)
		case errors.Is(err, ErrEmailChangeExists):
			shared.RespondError(c, response.CodeBadRequest, "error.email_change_exists", nil)
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

// ChangeEmailRequest 更换邮箱请求。
type ChangeEmailRequest struct {
	NewEmail string `json:"new_email" binding:"required"`
	OldCode  string `json:"old_code"`
	NewCode  string `json:"new_code" binding:"required"`
}

// ChangeEmail 更换邮箱。
func (h *UserEmailHandler) ChangeEmail(c *gin.Context) {
	id, ok := shared.GetUserID(c)
	if !ok {
		return
	}

	var req ChangeEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		shared.RespondBindError(c, err)
		return
	}

	user, err := h.service.ChangeEmail(id, req.NewEmail, req.OldCode, req.NewCode)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidEmail):
			shared.RespondError(c, response.CodeBadRequest, "error.email_invalid", nil)
		case errors.Is(err, ErrEmailChangeInvalid):
			shared.RespondError(c, response.CodeBadRequest, "error.email_change_invalid", nil)
		case errors.Is(err, ErrEmailChangeExists):
			shared.RespondError(c, response.CodeBadRequest, "error.email_change_exists", nil)
		case errors.Is(err, ErrVerifyCodeInvalid):
			shared.RespondError(c, response.CodeBadRequest, "error.verify_code_invalid", nil)
		case errors.Is(err, ErrVerifyCodeExpired):
			shared.RespondError(c, response.CodeBadRequest, "error.verify_code_expired", nil)
		case errors.Is(err, ErrVerifyCodeAttemptsExceeded):
			shared.RespondError(c, response.CodeBadRequest, "error.verify_code_attempts_exceeded", nil)
		case errors.Is(err, ErrUserNotFound):
			shared.RespondError(c, response.CodeNotFound, "error.user_not_found", nil)
		default:
			shared.RespondError(c, response.CodeInternal, "error.email_change_failed", err)
		}
		return
	}

	profile, respErr := h.changeEmailProfileResponse(user)
	if respErr != nil {
		shared.RespondError(c, response.CodeInternal, "error.email_change_failed", respErr)
		return
	}
	response.Success(c, profile)
}

func (h *UserEmailHandler) changeEmailProfileResponse(user *models.User) (dto.UserProfileResp, error) {
	emailMode, err := h.service.ResolveEmailChangeMode(user)
	if err != nil {
		return dto.UserProfileResp{}, err
	}
	passwordMode, err := h.service.ResolvePasswordChangeMode(user)
	if err != nil {
		return dto.UserProfileResp{}, err
	}
	return dto.NewUserProfileResp(user, emailMode, passwordMode), nil
}
