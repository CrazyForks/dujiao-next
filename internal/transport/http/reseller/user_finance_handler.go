package resellerhttp

import (
	"strings"

	"github.com/dujiao-next/internal/dto"
	"github.com/dujiao-next/internal/http/handlers/shared"
	"github.com/dujiao-next/internal/http/response"
	"github.com/dujiao-next/internal/models"
	resellermodule "github.com/dujiao-next/internal/modules/reseller"
	"github.com/shopspring/decimal"

	"github.com/gin-gonic/gin"
)

// UserFinanceService 是用户中心分销财务端点所需的最小用例接口。
type UserFinanceService interface {
	GetUserFinanceDashboard(userID uint) (resellermodule.UserFinanceDashboard, error)
	ListUserBalanceAccounts(userID uint, filter resellermodule.UserBalanceAccountListFilter) ([]models.ResellerBalanceAccount, int64, error)
	ListUserLedgerEntries(userID uint, filter resellermodule.UserLedgerListFilter) ([]models.ResellerLedgerEntry, int64, error)
	ListUserWithdrawRequests(userID uint, filter resellermodule.UserWithdrawListFilter) ([]models.ResellerWithdrawRequest, int64, error)
	ApplyUserWithdraw(userID uint, input resellermodule.WithdrawApplyInput) (*models.ResellerWithdrawRequest, error)
}

// UserFinanceHandler 处理用户中心分销财务请求。
type UserFinanceHandler struct {
	finance UserFinanceService
}

func NewUserFinanceHandler(finance UserFinanceService) *UserFinanceHandler {
	if finance == nil {
		panic("reseller user finance handler: finance is nil")
	}
	return &UserFinanceHandler{finance: finance}
}

type withdrawApplyRequest struct {
	Amount   string `json:"amount" binding:"required"`
	Currency string `json:"currency" binding:"required"`
	Channel  string `json:"channel" binding:"required"`
	Account  string `json:"account" binding:"required"`
}

// GetDashboard 获取当前用户的分销商财务看板。
func (h *UserFinanceHandler) GetDashboard(c *gin.Context) {
	uid, ok := shared.GetUserID(c)
	if !ok {
		return
	}
	data, err := h.finance.GetUserFinanceDashboard(uid)
	if err != nil {
		shared.RespondError(c, response.CodeInternal, "error.user_fetch_failed", err)
		return
	}
	response.Success(c, dto.NewResellerDashboardResp(data.Opened, data.Profile, data.Balances, data.WithdrawEnabled, data.WithdrawDisabledReason))
}

// ListBalanceAccounts 查询当前用户的分销余额账户。
func (h *UserFinanceHandler) ListBalanceAccounts(c *gin.Context) {
	uid, ok := shared.GetUserID(c)
	if !ok {
		return
	}
	page, pageSize := shared.ParsePagination(c)
	rows, total, err := h.finance.ListUserBalanceAccounts(uid, resellermodule.UserBalanceAccountListFilter{
		Page:     page,
		PageSize: pageSize,
		Status:   strings.TrimSpace(c.Query("status")),
	})
	if err != nil {
		respondUserFinanceError(c, err, "error.user_fetch_failed")
		return
	}
	response.SuccessWithPage(c, dto.NewResellerBalanceRespList(rows), response.BuildPagination(page, pageSize, total))
}

// ListLedgerEntries 查询当前用户的分销账务流水。
func (h *UserFinanceHandler) ListLedgerEntries(c *gin.Context) {
	uid, ok := shared.GetUserID(c)
	if !ok {
		return
	}
	page, pageSize := shared.ParsePagination(c)
	orderID, err := shared.ParseQueryUint(c.Query("order_id"), false)
	if err != nil {
		shared.RespondError(c, response.CodeBadRequest, "error.bad_request", nil)
		return
	}
	rows, total, err := h.finance.ListUserLedgerEntries(uid, resellermodule.UserLedgerListFilter{
		Page:     page,
		PageSize: pageSize,
		Type:     strings.TrimSpace(c.Query("type")),
		Status:   strings.TrimSpace(c.Query("status")),
		OrderID:  orderID,
	})
	if err != nil {
		respondUserFinanceError(c, err, "error.user_fetch_failed")
		return
	}
	response.SuccessWithPage(c, dto.NewResellerLedgerRespList(rows), response.BuildPagination(page, pageSize, total))
}

// ListWithdraws 查询当前用户的分销提现申请。
func (h *UserFinanceHandler) ListWithdraws(c *gin.Context) {
	uid, ok := shared.GetUserID(c)
	if !ok {
		return
	}
	page, pageSize := shared.ParsePagination(c)
	rows, total, err := h.finance.ListUserWithdrawRequests(uid, resellermodule.UserWithdrawListFilter{
		Page:     page,
		PageSize: pageSize,
		Status:   strings.TrimSpace(c.Query("status")),
	})
	if err != nil {
		respondUserFinanceError(c, err, "error.user_fetch_failed")
		return
	}
	response.SuccessWithPage(c, dto.NewResellerWithdrawRespList(rows), response.BuildPagination(page, pageSize, total))
}

// ApplyWithdraw 提交当前用户的分销提现申请。
func (h *UserFinanceHandler) ApplyWithdraw(c *gin.Context) {
	uid, ok := shared.GetUserID(c)
	if !ok {
		return
	}
	var req withdrawApplyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		shared.RespondBindError(c, err)
		return
	}
	amount, err := decimal.NewFromString(strings.TrimSpace(req.Amount))
	if err != nil {
		shared.RespondError(c, response.CodeBadRequest, "error.bad_request", nil)
		return
	}
	row, err := h.finance.ApplyUserWithdraw(uid, resellermodule.WithdrawApplyInput{
		Amount:   amount,
		Currency: strings.TrimSpace(req.Currency),
		Channel:  strings.TrimSpace(req.Channel),
		Account:  strings.TrimSpace(req.Account),
	})
	if err != nil {
		respondUserFinanceError(c, err, "error.save_failed")
		return
	}
	response.Success(c, dto.NewResellerWithdrawResp(row))
}
