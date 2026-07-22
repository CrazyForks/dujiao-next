package service

import (
	"github.com/dujiao-next/internal/models"
	resellermodule "github.com/dujiao-next/internal/modules/reseller"
)

type ResellerAdminLedgerListFilter = resellermodule.AdminLedgerListFilter
type ResellerAdminBalanceAccountListFilter = resellermodule.AdminBalanceAccountListFilter
type ResellerAdminWithdrawListFilter = resellermodule.AdminWithdrawListFilter
type ResellerUserFinanceDashboard = resellermodule.UserFinanceDashboard
type ResellerUserLedgerListFilter = resellermodule.UserLedgerListFilter
type ResellerUserBalanceAccountListFilter = resellermodule.UserBalanceAccountListFilter
type ResellerUserWithdrawListFilter = resellermodule.UserWithdrawListFilter

func (s *ResellerAccountingService) GetUserFinanceDashboard(userID uint) (ResellerUserFinanceDashboard, error) {
	if s == nil || s.query == nil {
		return ResellerUserFinanceDashboard{Opened: false}, nil
	}
	return s.query.GetUserFinanceDashboard(userID)
}

func (s *ResellerAccountingService) ListUserBalanceAccounts(userID uint, filter ResellerUserBalanceAccountListFilter) ([]models.ResellerBalanceAccount, int64, error) {
	if s == nil || s.query == nil {
		return nil, 0, ErrResellerNotOpened
	}
	return s.query.ListUserBalanceAccounts(userID, filter)
}

func (s *ResellerAccountingService) ListUserLedgerEntries(userID uint, filter ResellerUserLedgerListFilter) ([]models.ResellerLedgerEntry, int64, error) {
	if s == nil || s.query == nil {
		return nil, 0, ErrResellerNotOpened
	}
	return s.query.ListUserLedgerEntries(userID, filter)
}

func (s *ResellerAccountingService) ListUserWithdrawRequests(userID uint, filter ResellerUserWithdrawListFilter) ([]models.ResellerWithdrawRequest, int64, error) {
	if s == nil || s.query == nil {
		return nil, 0, ErrResellerNotOpened
	}
	return s.query.ListUserWithdrawRequests(userID, filter)
}

func (s *ResellerAccountingService) ListAdminLedgerEntries(filter ResellerAdminLedgerListFilter) ([]models.ResellerLedgerEntry, int64, error) {
	if s == nil || s.query == nil {
		return []models.ResellerLedgerEntry{}, 0, nil
	}
	return s.query.ListAdminLedgerEntries(filter)
}

func (s *ResellerAccountingService) ListAdminBalanceAccounts(filter ResellerAdminBalanceAccountListFilter) ([]models.ResellerBalanceAccount, int64, error) {
	if s == nil || s.query == nil {
		return []models.ResellerBalanceAccount{}, 0, nil
	}
	return s.query.ListAdminBalanceAccounts(filter)
}

func (s *ResellerAccountingService) ListAdminWithdrawRequests(filter ResellerAdminWithdrawListFilter) ([]models.ResellerWithdrawRequest, int64, error) {
	if s == nil || s.query == nil {
		return []models.ResellerWithdrawRequest{}, 0, nil
	}
	return s.query.ListAdminWithdrawRequests(filter)
}
