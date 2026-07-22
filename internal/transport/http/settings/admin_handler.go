package settingshttp

import (
	"github.com/dujiao-next/internal/cache"
	"github.com/dujiao-next/internal/constants"
	"github.com/dujiao-next/internal/http/handlers/shared"
	"github.com/dujiao-next/internal/http/response"
	"github.com/dujiao-next/internal/models"
	settingsmodule "github.com/dujiao-next/internal/modules/settings"

	"github.com/gin-gonic/gin"
)

// AdminService 是后台通用设置端口。
type AdminService interface {
	GetByKey(key string) (models.JSON, error)
	UpdateWithEffects(key string, value map[string]interface{}) (settingsmodule.UpdateResult, error)
	InvalidateCallbackRoutesCache()
}

// AdminHandler 处理后台通用设置请求。
type AdminHandler struct {
	settings AdminService
}

func NewAdminHandler(settings AdminService) *AdminHandler {
	if settings == nil {
		panic("settings admin handler: settings is nil")
	}
	return &AdminHandler{settings: settings}
}

type updateRequest struct {
	Key   string                 `json:"key" binding:"required"`
	Value map[string]interface{} `json:"value" binding:"required"`
}

// Get 获取设置。
func (h *AdminHandler) Get(c *gin.Context) {
	key := c.DefaultQuery("key", constants.SettingKeySiteConfig)

	value, err := h.settings.GetByKey(key)
	if err != nil {
		shared.RespondError(c, response.CodeInternal, "error.settings_fetch_failed", err)
		return
	}
	if value == nil {
		response.Success(c, gin.H{})
		return
	}

	response.Success(c, value)
}

// Update 更新设置。
func (h *AdminHandler) Update(c *gin.Context) {
	var req updateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		shared.RespondBindError(c, err)
		return
	}

	result, err := h.settings.UpdateWithEffects(req.Key, req.Value)
	if err != nil {
		shared.RespondError(c, response.CodeInternal, "error.settings_save_failed", err)
		return
	}

	if result.HasEffect(settingsmodule.EffectInvalidatePublicConfigCache) {
		_ = cache.DelAllPublicConfig(c.Request.Context())
	}
	if result.HasEffect(settingsmodule.EffectInvalidateCallbackRoutesCache) {
		h.settings.InvalidateCallbackRoutesCache()
	}
	response.Success(c, result.Value)
}
