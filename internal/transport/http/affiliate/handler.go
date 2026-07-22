package affiliatehttp

import (
	"errors"
	"strings"

	ginutil "github.com/dujiao-next/internal/platform/http/ginutil"

	"github.com/dujiao-next/internal/dto"
	"github.com/dujiao-next/internal/models"
	"github.com/dujiao-next/internal/modules/affiliate"
	catalogproduct "github.com/dujiao-next/internal/modules/catalog/product"
	"github.com/dujiao-next/internal/platform/http/response"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

// Service 是前台推广返利端口。
type Service interface {
	TrackClick(input affiliate.TrackClickInput) error
	OpenAffiliate(userID uint) (*models.AffiliateProfile, error)
	GetUserDashboard(userID uint) (affiliate.Dashboard, error)
	ListUserCommissions(userID uint, page, pageSize int, status string) ([]models.AffiliateCommission, int64, error)
	ListUserWithdraws(userID uint, page, pageSize int, status string) ([]models.AffiliateWithdrawRequest, int64, error)
	ApplyWithdraw(userID uint, input affiliate.WithdrawApplyInput) (*models.AffiliateWithdrawRequest, error)
}

type trackClickRequest struct {
	AffiliateCode string `json:"affiliate_code" binding:"required"`
	VisitorKey    string `json:"visitor_key"`
	LandingPath   string `json:"landing_path"`
	Referrer      string `json:"referrer"`
}

type withdrawApplyRequest struct {
	Amount  string `json:"amount" binding:"required"`
	Channel string `json:"channel" binding:"required"`
	Account string `json:"account" binding:"required"`
}

// Handler 处理前台推广返利请求。
type Handler struct {
	svc Service
}

func NewHandler(svc Service) *Handler {
	if svc == nil {
		panic("affiliate handler: service is nil")
	}
	return &Handler{svc: svc}
}

// TrackAffiliateClick 记录推广点击
func (h *Handler) TrackAffiliateClick(c *gin.Context) {
	var req trackClickRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ginutil.RespondBindError(c, err)
		return
	}

	if err := h.svc.TrackClick(affiliate.TrackClickInput{
		AffiliateCode: req.AffiliateCode,
		VisitorKey:    req.VisitorKey,
		LandingPath:   req.LandingPath,
		Referrer:      req.Referrer,
		ClientIP:      c.ClientIP(),
		UserAgent:     c.GetHeader("User-Agent"),
	}); err != nil {
		ginutil.RespondError(c, response.CodeInternal, "error.save_failed", err)
		return
	}
	response.Success(c, gin.H{"ok": true})
}

// OpenAffiliate 开通推广返利
func (h *Handler) OpenAffiliate(c *gin.Context) {
	uid, ok := ginutil.GetUserID(c)
	if !ok {
		return
	}

	profile, err := h.svc.OpenAffiliate(uid)
	if err != nil {
		switch {
		case errors.Is(err, affiliate.ErrDisabled):
			ginutil.RespondError(c, response.CodeBadRequest, "error.forbidden", nil)
		case errors.Is(err, catalogproduct.ErrNotFound):
			ginutil.RespondError(c, response.CodeNotFound, "error.user_not_found", nil)
		default:
			ginutil.RespondError(c, response.CodeInternal, "error.save_failed", err)
		}
		return
	}
	response.Success(c, dto.NewAffiliateProfileResp(profile))
}

// GetAffiliateDashboard 获取推广返利看板
func (h *Handler) GetAffiliateDashboard(c *gin.Context) {
	uid, ok := ginutil.GetUserID(c)
	if !ok {
		return
	}
	data, err := h.svc.GetUserDashboard(uid)
	if err != nil {
		ginutil.RespondError(c, response.CodeInternal, "error.user_fetch_failed", err)
		return
	}
	response.Success(c, data)
}

// ListAffiliateCommissions 查询我的推广佣金记录
func (h *Handler) ListAffiliateCommissions(c *gin.Context) {
	uid, ok := ginutil.GetUserID(c)
	if !ok {
		return
	}
	page, pageSize := ginutil.ParsePagination(c)
	status := strings.TrimSpace(c.Query("status"))

	rows, total, err := h.svc.ListUserCommissions(uid, page, pageSize, status)
	if err != nil {
		ginutil.RespondError(c, response.CodeInternal, "error.user_fetch_failed", err)
		return
	}
	response.SuccessWithPage(c, dto.NewAffiliateCommissionRespList(rows), response.BuildPagination(page, pageSize, total))
}

// ListAffiliateWithdraws 查询我的提现申请记录
func (h *Handler) ListAffiliateWithdraws(c *gin.Context) {
	uid, ok := ginutil.GetUserID(c)
	if !ok {
		return
	}
	page, pageSize := ginutil.ParsePagination(c)
	status := strings.TrimSpace(c.Query("status"))

	rows, total, err := h.svc.ListUserWithdraws(uid, page, pageSize, status)
	if err != nil {
		ginutil.RespondError(c, response.CodeInternal, "error.user_fetch_failed", err)
		return
	}
	response.SuccessWithPage(c, dto.NewAffiliateWithdrawRespList(rows), response.BuildPagination(page, pageSize, total))
}

// ApplyAffiliateWithdraw 提交提现申请
func (h *Handler) ApplyAffiliateWithdraw(c *gin.Context) {
	uid, ok := ginutil.GetUserID(c)
	if !ok {
		return
	}

	var req withdrawApplyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ginutil.RespondBindError(c, err)
		return
	}
	amount, err := decimal.NewFromString(strings.TrimSpace(req.Amount))
	if err != nil {
		ginutil.RespondError(c, response.CodeBadRequest, "error.bad_request", nil)
		return
	}

	row, err := h.svc.ApplyWithdraw(uid, affiliate.WithdrawApplyInput{
		Amount:  amount,
		Channel: req.Channel,
		Account: req.Account,
	})
	if err != nil {
		switch {
		case errors.Is(err, affiliate.ErrDisabled):
			ginutil.RespondError(c, response.CodeBadRequest, "error.forbidden", nil)
		case errors.Is(err, affiliate.ErrNotOpened):
			ginutil.RespondError(c, response.CodeBadRequest, "error.bad_request", nil)
		case errors.Is(err, affiliate.ErrWithdrawAmountInvalid):
			ginutil.RespondError(c, response.CodeBadRequest, "error.bad_request", nil)
		case errors.Is(err, affiliate.ErrWithdrawChannelInvalid):
			ginutil.RespondError(c, response.CodeBadRequest, "error.bad_request", nil)
		case errors.Is(err, affiliate.ErrWithdrawInsufficient):
			ginutil.RespondError(c, response.CodeBadRequest, "error.bad_request", nil)
		default:
			ginutil.RespondError(c, response.CodeInternal, "error.save_failed", err)
		}
		return
	}
	response.Success(c, dto.NewAffiliateWithdrawResp(row))
}
