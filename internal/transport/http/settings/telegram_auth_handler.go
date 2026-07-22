package settingshttp

import (
	"errors"

	"github.com/dujiao-next/internal/cache"
	"github.com/dujiao-next/internal/http/handlers/shared"
	"github.com/dujiao-next/internal/http/response"
	settingsmodule "github.com/dujiao-next/internal/modules/settings"

	"github.com/gin-gonic/gin"
)

// TelegramAuthAdminService 是后台 Telegram 登录设置端口。
type TelegramAuthAdminService interface {
	GetTelegramAuthSetting() (settingsmodule.TelegramAuthSetting, error)
	PatchTelegramAuthSetting(patch settingsmodule.TelegramAuthSettingPatch) (settingsmodule.TelegramAuthSetting, error)
	ApplyRuntime(setting settingsmodule.TelegramAuthSetting)
}

// TelegramAuthHandler 处理后台 Telegram 登录设置请求。
type TelegramAuthHandler struct {
	telegramAuth TelegramAuthAdminService
}

func NewTelegramAuthHandler(telegramAuth TelegramAuthAdminService) *TelegramAuthHandler {
	if telegramAuth == nil {
		panic("settings telegram auth handler: telegramAuth is nil")
	}
	return &TelegramAuthHandler{telegramAuth: telegramAuth}
}

// GetTelegramAuth 获取 Telegram 登录配置（脱敏）。
func (h *TelegramAuthHandler) GetTelegramAuth(c *gin.Context) {
	setting, err := h.telegramAuth.GetTelegramAuthSetting()
	if err != nil {
		shared.RespondError(c, response.CodeInternal, "error.settings_fetch_failed", err)
		return
	}
	response.Success(c, settingsmodule.MaskTelegramAuthSettingForAdmin(setting))
}

// UpdateTelegramAuth 更新 Telegram 登录配置。
func (h *TelegramAuthHandler) UpdateTelegramAuth(c *gin.Context) {
	var req settingsmodule.TelegramAuthSettingPatch
	if err := c.ShouldBindJSON(&req); err != nil {
		shared.RespondBindError(c, err)
		return
	}

	setting, err := h.telegramAuth.PatchTelegramAuthSetting(req)
	if err != nil {
		switch {
		case errors.Is(err, settingsmodule.ErrTelegramAuthConfigInvalid):
			shared.RespondErrorWithMsg(c, response.CodeBadRequest, err.Error(), nil)
		default:
			shared.RespondError(c, response.CodeInternal, "error.settings_save_failed", err)
		}
		return
	}

	h.telegramAuth.ApplyRuntime(setting)
	_ = cache.DelAllPublicConfig(c.Request.Context())
	response.Success(c, settingsmodule.MaskTelegramAuthSettingForAdmin(setting))
}
