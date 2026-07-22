package settingshttp

import (
	"errors"

	"github.com/dujiao-next/internal/http/handlers/shared"
	"github.com/dujiao-next/internal/http/response"
	settingsmodule "github.com/dujiao-next/internal/modules/settings"

	"github.com/gin-gonic/gin"
)

// AffiliateAdminService 是后台推广返利设置端口。
type AffiliateAdminService interface {
	GetAffiliateSetting() (settingsmodule.AffiliateSetting, error)
	UpdateAffiliateSetting(setting settingsmodule.AffiliateSetting) (settingsmodule.AffiliateSetting, error)
}

// AffiliateHandler 处理后台推广返利设置请求。
type AffiliateHandler struct {
	affiliate AffiliateAdminService
}

func NewAffiliateHandler(affiliate AffiliateAdminService) *AffiliateHandler {
	if affiliate == nil {
		panic("settings affiliate handler: affiliate is nil")
	}
	return &AffiliateHandler{affiliate: affiliate}
}

// GetAffiliate 获取推广返利设置。
func (h *AffiliateHandler) GetAffiliate(c *gin.Context) {
	setting, err := h.affiliate.GetAffiliateSetting()
	if err != nil {
		shared.RespondError(c, response.CodeInternal, "error.settings_fetch_failed", err)
		return
	}
	response.Success(c, setting)
}

// UpdateAffiliate 更新推广返利设置。
func (h *AffiliateHandler) UpdateAffiliate(c *gin.Context) {
	var req settingsmodule.AffiliateSetting
	if err := c.ShouldBindJSON(&req); err != nil {
		shared.RespondBindError(c, err)
		return
	}

	setting, err := h.affiliate.UpdateAffiliateSetting(req)
	if err != nil {
		if errors.Is(err, settingsmodule.ErrAffiliateConfigInvalid) {
			shared.RespondError(c, response.CodeBadRequest, "error.bad_request", nil)
			return
		}
		shared.RespondError(c, response.CodeInternal, "error.settings_save_failed", err)
		return
	}
	response.Success(c, setting)
}
