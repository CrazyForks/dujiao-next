package affiliatehttp

import (
	"errors"
	"strings"

	"github.com/dujiao-next/internal/constants"
	"github.com/dujiao-next/internal/http/handlers/shared"
	"github.com/dujiao-next/internal/http/response"
	"github.com/dujiao-next/internal/models"
	"github.com/dujiao-next/internal/modules/affiliate"
	catalogproduct "github.com/dujiao-next/internal/modules/catalog/product"

	"github.com/gin-gonic/gin"
)

// AdminService 是后台推广返利管理端口。
type AdminService interface {
	ListAdminUsers(filter affiliate.AdminProfileListFilter) ([]affiliate.AdminUserItem, int64, error)
	ListAdminCommissions(filter affiliate.AdminCommissionListFilter) ([]models.AffiliateCommission, int64, error)
	ListAdminWithdraws(filter affiliate.AdminWithdrawListFilter) ([]models.AffiliateWithdrawRequest, int64, error)
	UpdateAffiliateProfileStatus(profileID uint, status string) (*models.AffiliateProfile, error)
	BatchUpdateAffiliateProfileStatus(profileIDs []uint, status string) (int64, error)
	ReviewWithdraw(adminID, withdrawID uint, action, reason string) (*models.AffiliateWithdrawRequest, error)
}

type profileStatusRequest struct {
	Status string `json:"status" binding:"required"`
}

type batchProfileStatusRequest struct {
	ProfileIDs []uint `json:"profile_ids" binding:"required"`
	Status     string `json:"status" binding:"required"`
}

type reviewWithdrawRequest struct {
	Reason string `json:"reason"`
}

// AdminHandler 处理后台推广返利管理请求。
type AdminHandler struct {
	svc AdminService
}

func NewAdminHandler(svc AdminService) *AdminHandler {
	if svc == nil {
		panic("affiliate admin handler: service is nil")
	}
	return &AdminHandler{svc: svc}
}

// ListAffiliateUsers 管理端推广用户列表
func (h *AdminHandler) ListAffiliateUsers(c *gin.Context) {
	page, pageSize := shared.ParsePagination(c)
	userID, _ := shared.ParseQueryUint(c.Query("user_id"), false)

	rows, total, err := h.svc.ListAdminUsers(affiliate.AdminProfileListFilter{
		Page:     page,
		PageSize: pageSize,
		UserID:   userID,
		Status:   strings.TrimSpace(c.Query("status")),
		Code:     strings.TrimSpace(c.Query("code")),
		Keyword:  strings.TrimSpace(c.Query("keyword")),
	})
	if err != nil {
		shared.RespondError(c, response.CodeInternal, "error.user_fetch_failed", err)
		return
	}
	response.SuccessWithPage(c, rows, response.BuildPagination(page, pageSize, total))
}

// ListAffiliateCommissions 管理端佣金列表
func (h *AdminHandler) ListAffiliateCommissions(c *gin.Context) {
	page, pageSize := shared.ParsePagination(c)
	profileID, _ := shared.ParseQueryUint(c.Query("affiliate_profile_id"), false)

	rows, total, err := h.svc.ListAdminCommissions(affiliate.AdminCommissionListFilter{
		Page:               page,
		PageSize:           pageSize,
		AffiliateProfileID: profileID,
		OrderNo:            strings.TrimSpace(c.Query("order_no")),
		Status:             strings.TrimSpace(c.Query("status")),
		Keyword:            strings.TrimSpace(c.Query("keyword")),
	})
	if err != nil {
		shared.RespondError(c, response.CodeInternal, "error.user_fetch_failed", err)
		return
	}
	response.SuccessWithPage(c, rows, response.BuildPagination(page, pageSize, total))
}

// ListAffiliateWithdraws 管理端提现审核列表
func (h *AdminHandler) ListAffiliateWithdraws(c *gin.Context) {
	page, pageSize := shared.ParsePagination(c)
	profileID, _ := shared.ParseQueryUint(c.Query("affiliate_profile_id"), false)

	rows, total, err := h.svc.ListAdminWithdraws(affiliate.AdminWithdrawListFilter{
		Page:               page,
		PageSize:           pageSize,
		AffiliateProfileID: profileID,
		Status:             strings.TrimSpace(c.Query("status")),
		Keyword:            strings.TrimSpace(c.Query("keyword")),
	})
	if err != nil {
		shared.RespondError(c, response.CodeInternal, "error.user_fetch_failed", err)
		return
	}
	response.SuccessWithPage(c, rows, response.BuildPagination(page, pageSize, total))
}

// UpdateAffiliateUserStatus 管理端更新返利用户状态
func (h *AdminHandler) UpdateAffiliateUserStatus(c *gin.Context) {
	id, err := shared.ParseParamUint(c, "id")
	if err != nil {
		shared.RespondError(c, response.CodeBadRequest, "error.bad_request", nil)
		return
	}

	var req profileStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		shared.RespondBindError(c, err)
		return
	}

	row, err := h.svc.UpdateAffiliateProfileStatus(id, strings.TrimSpace(req.Status))
	if err != nil {
		switch {
		case errors.Is(err, catalogproduct.ErrNotFound):
			shared.RespondError(c, response.CodeNotFound, "error.bad_request", nil)
		case errors.Is(err, affiliate.ErrProfileStatusInvalid):
			shared.RespondError(c, response.CodeBadRequest, "error.bad_request", nil)
		default:
			shared.RespondError(c, response.CodeInternal, "error.save_failed", err)
		}
		return
	}
	response.Success(c, row)
}

// BatchUpdateAffiliateUserStatus 管理端批量更新返利用户状态
func (h *AdminHandler) BatchUpdateAffiliateUserStatus(c *gin.Context) {
	var req batchProfileStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		shared.RespondBindError(c, err)
		return
	}
	if len(req.ProfileIDs) == 0 {
		shared.RespondError(c, response.CodeBadRequest, "error.bad_request", nil)
		return
	}
	updated, err := h.svc.BatchUpdateAffiliateProfileStatus(req.ProfileIDs, strings.TrimSpace(req.Status))
	if err != nil {
		switch {
		case errors.Is(err, affiliate.ErrProfileStatusInvalid):
			shared.RespondError(c, response.CodeBadRequest, "error.bad_request", nil)
		default:
			shared.RespondError(c, response.CodeInternal, "error.save_failed", err)
		}
		return
	}
	response.Success(c, gin.H{"updated": updated})
}

// RejectAffiliateWithdraw 拒绝提现申请
func (h *AdminHandler) RejectAffiliateWithdraw(c *gin.Context) {
	adminID, ok := shared.GetAdminID(c)
	if !ok {
		return
	}
	id, err := shared.ParseParamUint(c, "id")
	if err != nil {
		shared.RespondError(c, response.CodeBadRequest, "error.bad_request", nil)
		return
	}

	var req reviewWithdrawRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		shared.RespondBindError(c, err)
		return
	}
	row, err := h.svc.ReviewWithdraw(adminID, id, constants.AffiliateWithdrawActionReject, req.Reason)
	if err != nil {
		switch {
		case errors.Is(err, catalogproduct.ErrNotFound):
			shared.RespondError(c, response.CodeNotFound, "error.bad_request", nil)
		case errors.Is(err, affiliate.ErrWithdrawStatusInvalid):
			shared.RespondError(c, response.CodeBadRequest, "error.bad_request", nil)
		default:
			shared.RespondError(c, response.CodeInternal, "error.save_failed", err)
		}
		return
	}
	response.Success(c, row)
}

// PayAffiliateWithdraw 标记提现已支付
func (h *AdminHandler) PayAffiliateWithdraw(c *gin.Context) {
	adminID, ok := shared.GetAdminID(c)
	if !ok {
		return
	}
	id, err := shared.ParseParamUint(c, "id")
	if err != nil {
		shared.RespondError(c, response.CodeBadRequest, "error.bad_request", nil)
		return
	}
	row, err := h.svc.ReviewWithdraw(adminID, id, constants.AffiliateWithdrawActionPay, "")
	if err != nil {
		switch {
		case errors.Is(err, catalogproduct.ErrNotFound):
			shared.RespondError(c, response.CodeNotFound, "error.bad_request", nil)
		case errors.Is(err, affiliate.ErrWithdrawStatusInvalid):
			shared.RespondError(c, response.CodeBadRequest, "error.bad_request", nil)
		default:
			shared.RespondError(c, response.CodeInternal, "error.save_failed", err)
		}
		return
	}
	response.Success(c, row)
}
