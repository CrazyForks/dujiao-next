package settingshttp

import (
	"github.com/dujiao-next/internal/http/handlers/shared"
	"github.com/dujiao-next/internal/http/response"
	settingsmodule "github.com/dujiao-next/internal/modules/settings"

	"github.com/gin-gonic/gin"
)

// TelegramBotAdminService 是后台 Telegram Bot 设置端口。
type TelegramBotAdminService interface {
	GetTelegramBotConfig() (settingsmodule.TelegramBotConfigSetting, error)
	UpdateTelegramBotConfig(cfg settingsmodule.TelegramBotConfigSetting) (settingsmodule.TelegramBotConfigSetting, error)
	GetTelegramBotRuntimeStatus() (settingsmodule.TelegramBotRuntimeStatusSetting, error)
}

// TelegramBotHandler 处理后台 Telegram Bot 设置请求。
type TelegramBotHandler struct {
	bot TelegramBotAdminService
}

func NewTelegramBotHandler(bot TelegramBotAdminService) *TelegramBotHandler {
	if bot == nil {
		panic("settings telegram-bot handler: bot is nil")
	}
	return &TelegramBotHandler{bot: bot}
}

// GetTelegramBotConfig 获取 Telegram Bot 配置。
func (h *TelegramBotHandler) GetTelegramBotConfig(c *gin.Context) {
	setting, err := h.bot.GetTelegramBotConfig()
	if err != nil {
		shared.RespondError(c, response.CodeInternal, "error.settings_fetch_failed", err)
		return
	}
	response.Success(c, settingsmodule.MaskTelegramBotConfigForAdmin(setting))
}

// UpdateTelegramBotConfig 更新 Telegram Bot 配置（整对象覆盖）。
func (h *TelegramBotHandler) UpdateTelegramBotConfig(c *gin.Context) {
	var req settingsmodule.TelegramBotConfigSetting
	if err := c.ShouldBindJSON(&req); err != nil {
		shared.RespondBindError(c, err)
		return
	}

	setting, err := h.bot.UpdateTelegramBotConfig(req)
	if err != nil {
		shared.RespondError(c, response.CodeInternal, "error.settings_save_failed", err)
		return
	}

	response.Success(c, settingsmodule.MaskTelegramBotConfigForAdmin(setting))
}

// GetTelegramBotRuntimeStatus 获取 Telegram Bot 运行时状态。
func (h *TelegramBotHandler) GetTelegramBotRuntimeStatus(c *gin.Context) {
	status, err := h.bot.GetTelegramBotRuntimeStatus()
	if err != nil {
		shared.RespondError(c, response.CodeInternal, "error.settings_fetch_failed", err)
		return
	}
	response.Success(c, status)
}
