package resellerhttp

import (
	"errors"
	"strings"
	"time"

	"github.com/dujiao-next/internal/models"
	catalogproduct "github.com/dujiao-next/internal/modules/catalog/product"
	resellermodule "github.com/dujiao-next/internal/modules/reseller"
	"github.com/dujiao-next/internal/platform/http/ginutil"
	"github.com/dujiao-next/internal/platform/http/response"

	"github.com/gin-gonic/gin"
)

// AdminFinanceService 是后台分销财务端点所需的最小用例接口。
type AdminFinanceService interface {
	ListAdminLedgerEntries(filter resellermodule.AdminLedgerListFilter) ([]models.ResellerLedgerEntry, int64, error)
	ListAdminBalanceAccounts(filter resellermodule.AdminBalanceAccountListFilter) ([]models.ResellerBalanceAccount, int64, error)
	ListAdminWithdrawRequests(filter resellermodule.AdminWithdrawListFilter) ([]models.ResellerWithdrawRequest, int64, error)
	ReviewWithdraw(adminID, withdrawID uint, action, reason string) (*models.ResellerWithdrawRequest, error)
}

// AdminFinanceHandler 处理后台分销财务请求。
type AdminFinanceHandler struct {
	finance AdminFinanceService
}

func NewAdminFinanceHandler(finance AdminFinanceService) *AdminFinanceHandler {
	if finance == nil {
		panic("reseller admin finance handler: finance is nil")
	}
	return &AdminFinanceHandler{finance: finance}
}

type reviewWithdrawRequest struct {
	Reason string `json:"reason"`
}

// ListLedgerEntries 管理端分销账务流水列表。
func (h *AdminFinanceHandler) ListLedgerEntries(c *gin.Context) {
	page, pageSize := ginutil.ParsePagination(c)
	resellerID, _ := ginutil.ParseQueryUint(c.Query("reseller_id"), false)
	userID, _ := ginutil.ParseQueryUint(c.Query("user_id"), false)
	orderID, _ := ginutil.ParseQueryUint(c.Query("order_id"), false)
	rows, total, err := h.finance.ListAdminLedgerEntries(resellermodule.AdminLedgerListFilter{
		Page:        page,
		PageSize:    pageSize,
		ResellerID:  resellerID,
		UserID:      userID,
		Keyword:     strings.TrimSpace(c.Query("keyword")),
		Type:        strings.TrimSpace(c.Query("type")),
		Status:      strings.TrimSpace(c.Query("status")),
		OrderID:     orderID,
		OrderNo:     strings.TrimSpace(c.Query("order_no")),
		CreatedFrom: parseFinanceTimePointer(c.Query("created_from")),
		CreatedTo:   parseFinanceTimePointer(c.Query("created_to")),
	})
	if err != nil {
		ginutil.RespondError(c, response.CodeInternal, "error.user_fetch_failed", err)
		return
	}
	response.SuccessWithPage(c, rows, response.BuildPagination(page, pageSize, total))
}

// ListBalanceAccounts 管理端分销余额账户列表。
func (h *AdminFinanceHandler) ListBalanceAccounts(c *gin.Context) {
	page, pageSize := ginutil.ParsePagination(c)
	resellerID, _ := ginutil.ParseQueryUint(c.Query("reseller_id"), false)
	userID, _ := ginutil.ParseQueryUint(c.Query("user_id"), false)
	rows, total, err := h.finance.ListAdminBalanceAccounts(resellermodule.AdminBalanceAccountListFilter{
		Page:       page,
		PageSize:   pageSize,
		ResellerID: resellerID,
		UserID:     userID,
		Keyword:    strings.TrimSpace(c.Query("keyword")),
		Status:     strings.TrimSpace(c.Query("status")),
	})
	if err != nil {
		ginutil.RespondError(c, response.CodeInternal, "error.user_fetch_failed", err)
		return
	}
	response.SuccessWithPage(c, rows, response.BuildPagination(page, pageSize, total))
}

// ListWithdraws 管理端分销提现申请列表。
func (h *AdminFinanceHandler) ListWithdraws(c *gin.Context) {
	page, pageSize := ginutil.ParsePagination(c)
	resellerID, _ := ginutil.ParseQueryUint(c.Query("reseller_id"), false)
	userID, _ := ginutil.ParseQueryUint(c.Query("user_id"), false)
	rows, total, err := h.finance.ListAdminWithdrawRequests(resellermodule.AdminWithdrawListFilter{
		Page:        page,
		PageSize:    pageSize,
		ResellerID:  resellerID,
		UserID:      userID,
		Keyword:     strings.TrimSpace(c.Query("keyword")),
		Status:      strings.TrimSpace(c.Query("status")),
		CreatedFrom: parseFinanceTimePointer(c.Query("created_from")),
		CreatedTo:   parseFinanceTimePointer(c.Query("created_to")),
	})
	if err != nil {
		ginutil.RespondError(c, response.CodeInternal, "error.user_fetch_failed", err)
		return
	}
	response.SuccessWithPage(c, rows, response.BuildPagination(page, pageSize, total))
}

// RejectWithdraw 拒绝分销提现申请。
func (h *AdminFinanceHandler) RejectWithdraw(c *gin.Context) {
	adminID, ok := ginutil.GetAdminID(c)
	if !ok {
		return
	}
	id, err := ginutil.ParseParamUint(c, "id")
	if err != nil {
		ginutil.RespondError(c, response.CodeBadRequest, "error.bad_request", nil)
		return
	}
	var req reviewWithdrawRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ginutil.RespondBindError(c, err)
		return
	}
	row, err := h.finance.ReviewWithdraw(adminID, id, "reject", req.Reason)
	if err != nil {
		respondAdminWithdrawReviewError(c, err)
		return
	}
	response.Success(c, row)
}

// PayWithdraw 标记分销提现已打款。
func (h *AdminFinanceHandler) PayWithdraw(c *gin.Context) {
	adminID, ok := ginutil.GetAdminID(c)
	if !ok {
		return
	}
	id, err := ginutil.ParseParamUint(c, "id")
	if err != nil {
		ginutil.RespondError(c, response.CodeBadRequest, "error.bad_request", nil)
		return
	}
	row, err := h.finance.ReviewWithdraw(adminID, id, "pay", "")
	if err != nil {
		respondAdminWithdrawReviewError(c, err)
		return
	}
	response.Success(c, row)
}

func respondAdminWithdrawReviewError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, catalogproduct.ErrNotFound):
		ginutil.RespondError(c, response.CodeNotFound, "error.bad_request", nil)
	case errors.Is(err, resellermodule.ErrWithdrawStatusInvalid):
		ginutil.RespondError(c, response.CodeBadRequest, "error.bad_request", nil)
	default:
		ginutil.RespondError(c, response.CodeInternal, "error.save_failed", err)
	}
}

func parseFinanceTimePointer(raw string) *time.Time {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	if t, err := time.Parse(time.RFC3339, raw); err == nil {
		return &t
	}
	return nil
}
