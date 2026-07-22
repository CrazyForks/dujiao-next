package settingshttp

import (
	"errors"

	"github.com/dujiao-next/internal/cache"
	"github.com/dujiao-next/internal/http/handlers/shared"
	"github.com/dujiao-next/internal/http/response"
	"github.com/dujiao-next/internal/modules/captcha"
	settingsmodule "github.com/dujiao-next/internal/modules/settings"

	"github.com/gin-gonic/gin"
)

// CaptchaAdminService 是后台验证码设置端口。
type CaptchaAdminService interface {
	GetCaptchaSetting() (settingsmodule.CaptchaSetting, error)
	PatchCaptchaSetting(patch settingsmodule.CaptchaSettingPatch) (settingsmodule.CaptchaSetting, error)
	ApplyRuntime(setting settingsmodule.CaptchaSetting)
}

// CaptchaHandler 处理后台验证码设置请求。
type CaptchaHandler struct {
	captcha CaptchaAdminService
}

func NewCaptchaHandler(captcha CaptchaAdminService) *CaptchaHandler {
	if captcha == nil {
		panic("settings captcha handler: captcha is nil")
	}
	return &CaptchaHandler{captcha: captcha}
}

// GetCaptcha 获取验证码配置（脱敏）。
func (h *CaptchaHandler) GetCaptcha(c *gin.Context) {
	setting, err := h.captcha.GetCaptchaSetting()
	if err != nil {
		shared.RespondError(c, response.CodeInternal, "error.settings_fetch_failed", err)
		return
	}
	response.Success(c, settingsmodule.MaskCaptchaSettingForAdmin(setting))
}

// UpdateCaptcha 更新验证码配置。
func (h *CaptchaHandler) UpdateCaptcha(c *gin.Context) {
	var req settingsmodule.CaptchaSettingPatch
	if err := c.ShouldBindJSON(&req); err != nil {
		shared.RespondBindError(c, err)
		return
	}

	setting, err := h.captcha.PatchCaptchaSetting(req)
	if err != nil {
		switch {
		case errors.Is(err, captcha.ErrConfigInvalid):
			shared.RespondErrorWithMsg(c, response.CodeBadRequest, err.Error(), nil)
		default:
			shared.RespondError(c, response.CodeInternal, "error.settings_save_failed", err)
		}
		return
	}

	h.captcha.ApplyRuntime(setting)
	_ = cache.DelAllPublicConfig(c.Request.Context())
	response.Success(c, settingsmodule.MaskCaptchaSettingForAdmin(setting))
}
