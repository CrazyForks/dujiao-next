package service

import (
	"github.com/dujiao-next/internal/models"
	resellermodule "github.com/dujiao-next/internal/modules/reseller"
	"github.com/dujiao-next/internal/repository"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// resellerAccountingWithdrawStoreAdapter 把遗留 ResellerRepository 适配为 AccountingWithdrawStore。
type resellerAccountingWithdrawStoreAdapter struct {
	repo repository.ResellerRepository
}

func newResellerAccountingWithdrawStoreAdapter(repo repository.ResellerRepository) *resellerAccountingWithdrawStoreAdapter {
	return &resellerAccountingWithdrawStoreAdapter{repo: repo}
}

func (a *resellerAccountingWithdrawStoreAdapter) WithinTransaction(fn func(store resellermodule.AccountingWithdrawStore) error) error {
	if a == nil || a.repo == nil {
		return fn(a)
	}
	return a.repo.Transaction(func(tx *gorm.DB) error {
		return fn(&resellerAccountingWithdrawStoreAdapter{repo: a.repo.WithTx(tx)})
	})
}

func (a *resellerAccountingWithdrawStoreAdapter) GetProfileByUserID(userID uint) (*models.ResellerProfile, error) {
	return a.repo.GetProfileByUserID(userID)
}

func (a *resellerAccountingWithdrawStoreAdapter) GetOrCreateBalanceAccountForUpdate(resellerID uint, currency string) (*models.ResellerBalanceAccount, error) {
	return a.repo.GetOrCreateBalanceAccountForUpdate(resellerID, currency)
}

func (a *resellerAccountingWithdrawStoreAdapter) SumLedgerAmountGroupedByStatus(resellerID uint, currency string, statuses []string) (map[string]decimal.Decimal, error) {
	return a.repo.SumLedgerAmountGroupedByStatus(resellerID, currency, statuses)
}

func (a *resellerAccountingWithdrawStoreAdapter) UpdateBalanceAccount(account *models.ResellerBalanceAccount) error {
	return a.repo.UpdateBalanceAccount(account)
}

func (a *resellerAccountingWithdrawStoreAdapter) ListAvailableLedgerEntriesForUpdate(resellerID uint, currency string) ([]models.ResellerLedgerEntry, error) {
	return a.repo.ListAvailableLedgerEntriesForUpdate(resellerID, currency)
}

func (a *resellerAccountingWithdrawStoreAdapter) UpdateLedgerEntry(entry *models.ResellerLedgerEntry) error {
	return a.repo.UpdateLedgerEntry(entry)
}

func (a *resellerAccountingWithdrawStoreAdapter) CreateLedgerEntryIfNotExists(entry *models.ResellerLedgerEntry) (bool, error) {
	return a.repo.CreateLedgerEntryIfNotExists(entry)
}

func (a *resellerAccountingWithdrawStoreAdapter) BatchUpdateLedgerEntries(ids []uint, updates map[string]interface{}) error {
	return a.repo.BatchUpdateLedgerEntries(ids, updates)
}

func (a *resellerAccountingWithdrawStoreAdapter) BatchUpdateLedgerEntriesByWithdrawID(withdrawID uint, updates map[string]interface{}) error {
	return a.repo.BatchUpdateLedgerEntriesByWithdrawID(withdrawID, updates)
}

func (a *resellerAccountingWithdrawStoreAdapter) CreateWithdrawRequest(req *models.ResellerWithdrawRequest) error {
	return a.repo.CreateWithdrawRequest(req)
}

func (a *resellerAccountingWithdrawStoreAdapter) GetWithdrawRequestByID(id uint) (*models.ResellerWithdrawRequest, error) {
	return a.repo.GetWithdrawRequestByID(id)
}

func (a *resellerAccountingWithdrawStoreAdapter) GetWithdrawRequestByIDForUpdate(id uint) (*models.ResellerWithdrawRequest, error) {
	return a.repo.GetWithdrawRequestByIDForUpdate(id)
}

func (a *resellerAccountingWithdrawStoreAdapter) UpdateWithdrawRequest(req *models.ResellerWithdrawRequest) error {
	return a.repo.UpdateWithdrawRequest(req)
}
