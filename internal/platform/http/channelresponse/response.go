package channelresponse

import (
	"errors"
	"net/http"
	"strings"

	"github.com/dujiao-next/internal/i18n"
	"github.com/dujiao-next/internal/platform/http/ginutil"
	apiresponse "github.com/dujiao-next/internal/platform/http/response"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// Success 返回渠道 API 成功响应。
func Success(c *gin.Context, data interface{}) {
	apiresponse.ChannelSuccess(c, data)
}

// Error 返回渠道 API 错误响应，并在有原始错误时记录日志。
func Error(c *gin.Context, httpCode, code int, errorCode, key string, err error) {
	locale := i18n.ResolveLocale(c)
	message := i18n.T(locale, key)
	if err != nil {
		ginutil.RequestLog(c).Errorw("channel_handler_error",
			"http_code", httpCode,
			"code", code,
			"error_code", errorCode,
			"message", message,
			"error", err,
		)
	}
	apiresponse.ChannelError(c, httpCode, code, message, errorCode)
}

// BindError 返回渠道 API 参数绑定错误。
func BindError(c *gin.Context, err error) {
	locale := i18n.ResolveLocale(c)

	var validationErrors validator.ValidationErrors
	if errors.As(err, &validationErrors) {
		details := make([]string, 0, len(validationErrors))
		for _, fieldError := range validationErrors {
			details = append(details, formatFieldError(locale, fieldError))
		}
		message := strings.Join(details, "; ")
		ginutil.RequestLog(c).Warnw("channel_bind_validation_error", "details", message, "error", err)
		apiresponse.ChannelError(c, http.StatusBadRequest, apiresponse.CodeBadRequest, message, "validation_error")
		return
	}

	message := i18n.T(locale, "error.bad_request")
	ginutil.RequestLog(c).Warnw("channel_bind_error", "message", message, "error", err)
	apiresponse.ChannelError(c, http.StatusBadRequest, apiresponse.CodeBadRequest, message, "validation_error")
}

// UserIDValue 在渠道主身份字段为空时读取历史 Telegram 身份字段。
func UserIDValue(primary, fallback string) string {
	if value := strings.TrimSpace(primary); value != "" {
		return value
	}
	return strings.TrimSpace(fallback)
}

func formatFieldError(locale string, fieldError validator.FieldError) string {
	field := fieldError.Field()
	tag := fieldError.Tag()
	parameter := fieldError.Param()

	customKey := "validation." + field + "." + tag
	if message := i18n.T(locale, customKey); message != customKey {
		return message
	}

	ruleKey := "validation.rule." + tag
	if ruleMessage := i18n.T(locale, ruleKey); ruleMessage != ruleKey {
		if parameter != "" {
			return field + ": " + i18n.Sprintf(locale, ruleKey, parameter)
		}
		return field + ": " + ruleMessage
	}

	if parameter != "" {
		return field + ": " + tag + "=" + parameter
	}
	return field + ": " + tag
}
