package reseller

import (
	"testing"

	"github.com/dujiao-next/internal/models"
)

type accountingQueryStoreStub struct {
	profile  *models.ResellerProfile
	balances []models.ResellerBalanceAccount
	err      error
}

func (s accountingQueryStoreStub) GetProfileByUserID(userID uint) (*models.ResellerProfile, error) {
	return s.profile, s.err
}

func (s accountingQueryStoreStub) ListBalanceAccounts(filter BalanceAccountListFilter) ([]models.ResellerBalanceAccount, int64, error) {
	return s.balances, int64(len(s.balances)), s.err
}

func (s accountingQueryStoreStub) ListLedgerEntries(filter LedgerListFilter) ([]models.ResellerLedgerEntry, int64, error) {
	return nil, 0, s.err
}

func (s accountingQueryStoreStub) ListWithdrawRequests(filter WithdrawListFilter) ([]models.ResellerWithdrawRequest, int64, error) {
	return nil, 0, s.err
}

func (s accountingQueryStoreStub) ListAdminResellerLedgerEntries(filter AdminLedgerListFilter) ([]models.ResellerLedgerEntry, int64, error) {
	return nil, 0, s.err
}

func (s accountingQueryStoreStub) ListAdminResellerBalanceAccounts(filter AdminBalanceAccountListFilter) ([]models.ResellerBalanceAccount, int64, error) {
	return nil, 0, s.err
}

func (s accountingQueryStoreStub) ListAdminResellerWithdrawRequests(filter AdminWithdrawListFilter) ([]models.ResellerWithdrawRequest, int64, error) {
	return nil, 0, s.err
}

func TestAccountingQueryServiceDashboardNotOpened(t *testing.T) {
	svc := NewAccountingQueryService(accountingQueryStoreStub{})
	got, err := svc.GetUserFinanceDashboard(1)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if got.Opened {
		t.Fatalf("expected not opened")
	}
}

func TestRequireActiveProfileSettlementFrozen(t *testing.T) {
	err := RequireActiveProfile(&models.ResellerProfile{
		Status:           models.ResellerProfileStatusActive,
		SettlementStatus: models.ResellerSettlementStatusFrozen,
	})
	if err != ErrSettlementUnavailable {
		t.Fatalf("expected settlement unavailable, got %v", err)
	}
}

func TestWithdrawAvailabilityProfileInactive(t *testing.T) {
	ok, reason := WithdrawAvailability(&models.ResellerProfile{Status: models.ResellerProfileStatusDisabled})
	if ok || reason != WithdrawDisabledReasonProfileInactive {
		t.Fatalf("unexpected availability: ok=%v reason=%s", ok, reason)
	}
}
