package shared

import (
	"errors"
	"net/http"
	"strings"

	"github.com/dujiao-next/internal/i18n"
	telegramauthapp "github.com/dujiao-next/internal/modules/identity/telegramauth/application"
	userauthapp "github.com/dujiao-next/internal/modules/identity/userauth/application"
	"github.com/dujiao-next/internal/platform/http/ginutil"
	"github.com/dujiao-next/internal/platform/http/response"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// ChannelSuccess 返回渠道 API 成功响应。
func ChannelSuccess(c *gin.Context, data interface{}) {
	response.ChannelSuccess(c, data)
}

// ChannelError 返回渠道 API 错误响应，并在有原始错误时记录日志。
func ChannelError(c *gin.Context, httpCode, code int, errorCode, key string, err error) {
	locale := i18n.ResolveLocale(c)
	msg := i18n.T(locale, key)
	if err != nil {
		ginutil.RequestLog(c).Errorw("channel_handler_error",
			"http_code", httpCode,
			"code", code,
			"error_code", errorCode,
			"message", msg,
			"error", err,
		)
	}
	response.ChannelError(c, httpCode, code, msg, errorCode)
}

// ChannelBindError 返回渠道 API 参数绑定错误。
func ChannelBindError(c *gin.Context, err error) {
	locale := i18n.ResolveLocale(c)

	var ve validator.ValidationErrors
	if errors.As(err, &ve) {
		details := make([]string, 0, len(ve))
		for _, fe := range ve {
			details = append(details, formatChannelFieldError(locale, fe))
		}
		msg := strings.Join(details, "; ")
		ginutil.RequestLog(c).Warnw("channel_bind_validation_error", "details", msg, "error", err)
		response.ChannelError(c, http.StatusBadRequest, response.CodeBadRequest, msg, "validation_error")
		return
	}

	msg := i18n.T(locale, "error.bad_request")
	ginutil.RequestLog(c).Warnw("channel_bind_error", "message", msg, "error", err)
	response.ChannelError(c, http.StatusBadRequest, response.CodeBadRequest, msg, "validation_error")
}

// ChannelIdentityError 映射并返回渠道身份相关错误。
func ChannelIdentityError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, telegramauthapp.ErrTelegramAuthPayloadInvalid):
		ChannelError(c, http.StatusBadRequest, response.CodeBadRequest, "validation_error", "error.bad_request", nil)
	case errors.Is(err, userauthapp.ErrInvalidEmail):
		ChannelError(c, http.StatusBadRequest, response.CodeBadRequest, "validation_error", "error.email_invalid", nil)
	case errors.Is(err, userauthapp.ErrNotFound):
		ChannelError(c, http.StatusNotFound, response.CodeNotFound, "user_not_found", "error.user_not_found", nil)
	case errors.Is(err, userauthapp.ErrVerifyCodeInvalid):
		ChannelError(c, http.StatusBadRequest, response.CodeBadRequest, "verify_code_invalid", "error.verify_code_invalid", nil)
	case errors.Is(err, userauthapp.ErrVerifyCodeExpired):
		ChannelError(c, http.StatusBadRequest, response.CodeBadRequest, "verify_code_expired", "error.verify_code_expired", nil)
	case errors.Is(err, userauthapp.ErrVerifyCodeAttemptsExceeded):
		ChannelError(c, http.StatusBadRequest, response.CodeBadRequest, "verify_code_invalid", "error.verify_code_attempts_exceeded", nil)
	case errors.Is(err, userauthapp.ErrUserDisabled):
		ChannelError(c, http.StatusUnauthorized, response.CodeUnauthorized, "user_disabled", "error.user_disabled", nil)
	case errors.Is(err, userauthapp.ErrUserOAuthIdentityExists):
		ChannelError(c, http.StatusBadRequest, response.CodeBadRequest, "channel_identity_conflict", "error.telegram_bind_conflict", nil)
	case errors.Is(err, userauthapp.ErrUserOAuthAlreadyBound):
		ChannelError(c, http.StatusBadRequest, response.CodeBadRequest, "channel_identity_conflict", "error.telegram_already_bound", nil)
	default:
		ChannelError(c, http.StatusInternalServerError, response.CodeInternal, "internal_error", "error.internal_error", err)
	}
}

// ChannelUserIDValue 兼容 channel_user_id / telegram_user_id。
func ChannelUserIDValue(primary, legacy string) string {
	if value := strings.TrimSpace(primary); value != "" {
		return value
	}
	return strings.TrimSpace(legacy)
}

func formatChannelFieldError(locale string, fe validator.FieldError) string {
	field := fe.Field()
	tag := fe.Tag()
	param := fe.Param()

	customKey := "validation." + field + "." + tag
	if msg := i18n.T(locale, customKey); msg != customKey {
		return msg
	}

	ruleKey := "validation.rule." + tag
	if ruleMsg := i18n.T(locale, ruleKey); ruleMsg != ruleKey {
		if param != "" {
			return field + ": " + i18n.Sprintf(locale, ruleKey, param)
		}
		return field + ": " + ruleMsg
	}

	if param != "" {
		return field + ": " + tag + "=" + param
	}
	return field + ": " + tag
}
