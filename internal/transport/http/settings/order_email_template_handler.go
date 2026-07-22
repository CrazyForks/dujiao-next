package settingshttp

import (
	"errors"

	"github.com/dujiao-next/internal/http/handlers/shared"
	"github.com/dujiao-next/internal/http/response"
	settingsmodule "github.com/dujiao-next/internal/modules/settings"

	"github.com/gin-gonic/gin"
)

// OrderEmailTemplateAdminService 是后台订单邮件模板设置端口。
type OrderEmailTemplateAdminService interface {
	GetOrderEmailTemplateSetting() (settingsmodule.OrderEmailTemplateSetting, error)
	PatchOrderEmailTemplateSetting(patch settingsmodule.OrderEmailTemplateSettingPatch) (settingsmodule.OrderEmailTemplateSetting, error)
	ResetOrderEmailTemplateSetting() (settingsmodule.OrderEmailTemplateSetting, error)
}

// OrderEmailTemplateHandler 处理后台订单邮件模板设置请求。
type OrderEmailTemplateHandler struct {
	templates OrderEmailTemplateAdminService
}

func NewOrderEmailTemplateHandler(templates OrderEmailTemplateAdminService) *OrderEmailTemplateHandler {
	if templates == nil {
		panic("settings order-email-template handler: templates is nil")
	}
	return &OrderEmailTemplateHandler{templates: templates}
}

// GetOrderEmailTemplate 获取订单邮件模板配置。
func (h *OrderEmailTemplateHandler) GetOrderEmailTemplate(c *gin.Context) {
	setting, err := h.templates.GetOrderEmailTemplateSetting()
	if err != nil {
		shared.RespondError(c, response.CodeInternal, "error.settings_fetch_failed", err)
		return
	}
	response.Success(c, settingsmodule.MaskOrderEmailTemplateSettingForAdmin(setting))
}

// UpdateOrderEmailTemplate 更新订单邮件模板配置。
func (h *OrderEmailTemplateHandler) UpdateOrderEmailTemplate(c *gin.Context) {
	var req settingsmodule.OrderEmailTemplateSettingPatch
	if err := c.ShouldBindJSON(&req); err != nil {
		shared.RespondBindError(c, err)
		return
	}

	setting, err := h.templates.PatchOrderEmailTemplateSetting(req)
	if err != nil {
		switch {
		case errors.Is(err, settingsmodule.ErrOrderEmailTemplateConfigInvalid):
			shared.RespondErrorWithMsg(c, response.CodeBadRequest, err.Error(), nil)
		default:
			shared.RespondError(c, response.CodeInternal, "error.settings_save_failed", err)
		}
		return
	}

	response.Success(c, settingsmodule.MaskOrderEmailTemplateSettingForAdmin(setting))
}

// ResetOrderEmailTemplate 重置订单邮件模板为默认。
func (h *OrderEmailTemplateHandler) ResetOrderEmailTemplate(c *gin.Context) {
	setting, err := h.templates.ResetOrderEmailTemplateSetting()
	if err != nil {
		shared.RespondError(c, response.CodeInternal, "error.settings_save_failed", err)
		return
	}
	response.Success(c, settingsmodule.MaskOrderEmailTemplateSettingForAdmin(setting))
}
