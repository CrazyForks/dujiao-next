package resellerhttp

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	productcontract "github.com/dujiao-next/internal/modules/catalog/product/contract"

	"github.com/dujiao-next/internal/models"
	resellermodule "github.com/dujiao-next/internal/modules/reseller"
	"github.com/dujiao-next/internal/platform/http/response"

	"github.com/gin-gonic/gin"
)

type financeServiceStub struct {
	dashboard resellermodule.UserFinanceDashboard
	err       error
}

func (s financeServiceStub) GetUserFinanceDashboard(userID uint) (resellermodule.UserFinanceDashboard, error) {
	return s.dashboard, s.err
}

func (s financeServiceStub) ListUserBalanceAccounts(userID uint, filter resellermodule.UserBalanceAccountListFilter) ([]models.ResellerBalanceAccount, int64, error) {
	return nil, 0, s.err
}

func (s financeServiceStub) ListUserLedgerEntries(userID uint, filter resellermodule.UserLedgerListFilter) ([]models.ResellerLedgerEntry, int64, error) {
	return nil, 0, s.err
}

func (s financeServiceStub) ListUserWithdrawRequests(userID uint, filter resellermodule.UserWithdrawListFilter) ([]models.ResellerWithdrawRequest, int64, error) {
	return nil, 0, s.err
}

func (s financeServiceStub) ApplyUserWithdraw(userID uint, input resellermodule.WithdrawApplyInput) (*models.ResellerWithdrawRequest, error) {
	return nil, s.err
}

func (s financeServiceStub) ListAdminLedgerEntries(filter resellermodule.AdminLedgerListFilter) ([]models.ResellerLedgerEntry, int64, error) {
	return nil, 0, s.err
}

func (s financeServiceStub) ListAdminBalanceAccounts(filter resellermodule.AdminBalanceAccountListFilter) ([]models.ResellerBalanceAccount, int64, error) {
	return nil, 0, s.err
}

func (s financeServiceStub) ListAdminWithdrawRequests(filter resellermodule.AdminWithdrawListFilter) ([]models.ResellerWithdrawRequest, int64, error) {
	return nil, 0, s.err
}

func (s financeServiceStub) ReviewWithdraw(adminID, withdrawID uint, action, reason string) (*models.ResellerWithdrawRequest, error) {
	return nil, s.err
}

func TestUserFinanceHandlerMapsWithdrawError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewUserFinanceHandler(financeServiceStub{err: resellermodule.ErrWithdrawInsufficient})
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/reseller/withdraws", bytes.NewReader([]byte(`{"amount":"10","currency":"USD","channel":"usdt","account":"T"}`)))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("user_id", uint(9))

	h.ApplyWithdraw(c)

	var resp struct {
		StatusCode int `json:"status_code"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if resp.StatusCode != response.CodeBadRequest {
		t.Fatalf("expected bad request, body=%s", recorder.Body.String())
	}
}

func TestAdminFinanceHandlerMapsWithdrawStatusInvalid(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewAdminFinanceHandler(financeServiceStub{err: resellermodule.ErrWithdrawStatusInvalid})
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodPost, "/admin/resellers/withdraws/1/pay", nil)
	c.Set("admin_id", uint(1))
	c.Params = gin.Params{{Key: "id", Value: "1"}}

	h.PayWithdraw(c)

	var resp struct {
		StatusCode int `json:"status_code"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if resp.StatusCode != response.CodeBadRequest {
		t.Fatalf("expected bad request, body=%s", recorder.Body.String())
	}
}

func TestAdminFinanceHandlerMapsNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewAdminFinanceHandler(financeServiceStub{err: productcontract.ErrNotFound})
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodPost, "/admin/resellers/withdraws/1/reject", bytes.NewReader([]byte(`{"reason":"x"}`)))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("admin_id", uint(1))
	c.Params = gin.Params{{Key: "id", Value: "1"}}

	h.RejectWithdraw(c)

	var resp struct {
		StatusCode int `json:"status_code"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if resp.StatusCode != response.CodeNotFound {
		t.Fatalf("expected not found, body=%s", recorder.Body.String())
	}
}
