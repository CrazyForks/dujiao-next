package admin

import (
	"errors"
	"strings"

	"github.com/dujiao-next/internal/http/response"
	"github.com/dujiao-next/internal/service"

	"github.com/gin-gonic/gin"
)

// GetNotificationCenterSettings 获取通知中心配置
func (h *Handler) GetNotificationCenterSettings(c *gin.Context) {
	setting, err := h.SettingService.GetNotificationCenterSetting()
	if err != nil {
		respondError(c, response.CodeInternal, "error.settings_fetch_failed", err)
		return
	}
	response.Success(c, service.MaskNotificationCenterSettingForAdmin(setting))
}

// UpdateNotificationCenterSettings 更新通知中心配置
func (h *Handler) UpdateNotificationCenterSettings(c *gin.Context) {
	var req service.NotificationCenterSettingPatch
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, response.CodeBadRequest, "error.bad_request", err)
		return
	}

	setting, err := h.SettingService.PatchNotificationCenterSetting(req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrNotificationConfigInvalid):
			respondErrorWithMsg(c, response.CodeBadRequest, err.Error(), nil)
		default:
			respondError(c, response.CodeInternal, "error.settings_save_failed", err)
		}
		return
	}
	response.Success(c, service.MaskNotificationCenterSettingForAdmin(setting))
}

// NotificationCenterTestSendRequest 通知中心测试发送请求
type NotificationCenterTestSendRequest struct {
	Channel   string                 `json:"channel" binding:"required"`
	Target    string                 `json:"target" binding:"required"`
	Scene     string                 `json:"scene"`
	Locale    string                 `json:"locale"`
	Variables map[string]interface{} `json:"variables"`
}

// TestNotificationCenterSettings 通知中心测试发送
func (h *Handler) TestNotificationCenterSettings(c *gin.Context) {
	var req NotificationCenterTestSendRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, response.CodeBadRequest, "error.bad_request", err)
		return
	}
	channel := strings.ToLower(strings.TrimSpace(req.Channel))
	if channel != "email" && channel != "telegram" {
		respondError(c, response.CodeBadRequest, "error.bad_request", nil)
		return
	}

	if h.NotificationService == nil {
		respondError(c, response.CodeInternal, "error.notification_send_failed", nil)
		return
	}

	err := h.NotificationService.SendTest(c.Request.Context(), service.NotificationTestSendInput{
		Channel:   channel,
		Target:    strings.TrimSpace(req.Target),
		Scene:     strings.TrimSpace(req.Scene),
		Locale:    strings.TrimSpace(req.Locale),
		Variables: req.Variables,
	})
	if err != nil {
		switch {
		case errors.Is(err, service.ErrNotificationConfigInvalid):
			respondErrorWithMsg(c, response.CodeBadRequest, err.Error(), nil)
		default:
			respondError(c, response.CodeInternal, "error.notification_send_failed", err)
		}
		return
	}
	response.Success(c, gin.H{"sent": true})
}
