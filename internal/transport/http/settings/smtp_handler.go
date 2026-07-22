package settingshttp

import (
	"errors"
	"strings"

	"github.com/dujiao-next/internal/http/handlers/shared"
	"github.com/dujiao-next/internal/http/response"
	settingsmodule "github.com/dujiao-next/internal/modules/settings"

	"github.com/gin-gonic/gin"
)

var (
	ErrSMTPTestInvalidEmail         = errors.New("invalid email")
	ErrSMTPTestRecipientRejected    = errors.New("email recipient rejected")
	ErrSMTPTestServiceDisabled      = errors.New("email service disabled")
	ErrSMTPTestServiceNotConfigured = errors.New("email service not configured")
)

// SMTPAdminService 是后台 SMTP 设置端口。
type SMTPAdminService interface {
	GetSMTPSetting() (settingsmodule.SMTPSetting, error)
	PatchSMTPSetting(patch settingsmodule.SMTPSettingPatch) (settingsmodule.SMTPSetting, error)
	ApplyRuntime(setting settingsmodule.SMTPSetting)
	SendTest(setting settingsmodule.SMTPSetting, toEmail, subject, body string) error
}

// SMTPHandler 处理后台 SMTP 设置请求。
type SMTPHandler struct {
	smtp SMTPAdminService
}

func NewSMTPHandler(smtp SMTPAdminService) *SMTPHandler {
	if smtp == nil {
		panic("settings smtp handler: smtp is nil")
	}
	return &SMTPHandler{smtp: smtp}
}

type smtpTestSendRequest struct {
	ToEmail string `json:"to_email" binding:"required"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

// GetSMTP 获取 SMTP 配置（脱敏）。
func (h *SMTPHandler) GetSMTP(c *gin.Context) {
	setting, err := h.smtp.GetSMTPSetting()
	if err != nil {
		shared.RespondError(c, response.CodeInternal, "error.settings_fetch_failed", err)
		return
	}
	response.Success(c, settingsmodule.MaskSMTPSettingForAdmin(setting))
}

// UpdateSMTP 更新 SMTP 配置。
func (h *SMTPHandler) UpdateSMTP(c *gin.Context) {
	var req settingsmodule.SMTPSettingPatch
	if err := c.ShouldBindJSON(&req); err != nil {
		shared.RespondBindError(c, err)
		return
	}

	setting, err := h.smtp.PatchSMTPSetting(req)
	if err != nil {
		switch {
		case errors.Is(err, settingsmodule.ErrSMTPConfigInvalid):
			shared.RespondErrorWithMsg(c, response.CodeBadRequest, err.Error(), nil)
		default:
			shared.RespondError(c, response.CodeInternal, "error.settings_save_failed", err)
		}
		return
	}

	h.smtp.ApplyRuntime(setting)
	response.Success(c, settingsmodule.MaskSMTPSettingForAdmin(setting))
}

// TestSMTP 测试 SMTP 配置发送。
func (h *SMTPHandler) TestSMTP(c *gin.Context) {
	var req smtpTestSendRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		shared.RespondBindError(c, err)
		return
	}

	toEmail := strings.TrimSpace(req.ToEmail)
	if toEmail == "" {
		shared.RespondError(c, response.CodeBadRequest, "error.email_invalid", nil)
		return
	}

	setting, err := h.smtp.GetSMTPSetting()
	if err != nil {
		shared.RespondError(c, response.CodeInternal, "error.settings_fetch_failed", err)
		return
	}

	if err := h.smtp.SendTest(setting, toEmail, req.Subject, req.Body); err != nil {
		switch {
		case errors.Is(err, ErrSMTPTestInvalidEmail):
			shared.RespondError(c, response.CodeBadRequest, "error.email_invalid", nil)
		case errors.Is(err, ErrSMTPTestRecipientRejected):
			shared.RespondError(c, response.CodeBadRequest, "error.email_recipient_not_found", nil)
		case errors.Is(err, ErrSMTPTestServiceDisabled),
			errors.Is(err, ErrSMTPTestServiceNotConfigured):
			shared.RespondError(c, response.CodeBadRequest, "error.email_service_not_configured", err)
		default:
			shared.RespondError(c, response.CodeInternal, "error.send_verify_code_failed", err)
		}
		return
	}

	response.Success(c, gin.H{"sent": true})
}
