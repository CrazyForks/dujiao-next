package service

import (
	"time"

	"github.com/dujiao-next/internal/models"
	resellermodule "github.com/dujiao-next/internal/modules/reseller"
	"github.com/dujiao-next/internal/repository"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type resellerAccountingLedgerStoreAdapter struct {
	repo repository.ResellerRepository
}

func newResellerAccountingLedgerStoreAdapter(repo repository.ResellerRepository) *resellerAccountingLedgerStoreAdapter {
	return &resellerAccountingLedgerStoreAdapter{repo: repo}
}

func (a *resellerAccountingLedgerStoreAdapter) WithinTransaction(fn func(store resellermodule.AccountingLedgerStore) error) error {
	if a == nil || a.repo == nil {
		return fn(a)
	}
	return a.repo.Transaction(func(tx *gorm.DB) error {
		return fn(&resellerAccountingLedgerStoreAdapter{repo: a.repo.WithTx(tx)})
	})
}

func (a *resellerAccountingLedgerStoreAdapter) GetOrCreateBalanceAccountForUpdate(resellerID uint, currency string) (*models.ResellerBalanceAccount, error) {
	return a.repo.GetOrCreateBalanceAccountForUpdate(resellerID, currency)
}

func (a *resellerAccountingLedgerStoreAdapter) SumLedgerAmountGroupedByStatus(resellerID uint, currency string, statuses []string) (map[string]decimal.Decimal, error) {
	return a.repo.SumLedgerAmountGroupedByStatus(resellerID, currency, statuses)
}

func (a *resellerAccountingLedgerStoreAdapter) UpdateBalanceAccount(account *models.ResellerBalanceAccount) error {
	return a.repo.UpdateBalanceAccount(account)
}

func (a *resellerAccountingLedgerStoreAdapter) GetOrderSnapshotByOrderID(orderID uint) (*models.ResellerOrderSnapshot, error) {
	return a.repo.GetOrderSnapshotByOrderID(orderID)
}

func (a *resellerAccountingLedgerStoreAdapter) CreateLedgerEntryIfNotExists(entry *models.ResellerLedgerEntry) (bool, error) {
	return a.repo.CreateLedgerEntryIfNotExists(entry)
}

func (a *resellerAccountingLedgerStoreAdapter) ListDueLedgerScopes(now time.Time) ([]resellermodule.LedgerScope, error) {
	return a.repo.ListDueLedgerScopes(now)
}

func (a *resellerAccountingLedgerStoreAdapter) MarkDueLedgerEntriesAvailable(now time.Time) (int64, error) {
	return a.repo.MarkDueLedgerEntriesAvailable(now)
}

func (a *resellerAccountingLedgerStoreAdapter) SumLedgerAmountByOrderAndType(orderID uint, ledgerType string) (decimal.Decimal, error) {
	return a.repo.SumLedgerAmountByOrderAndType(orderID, ledgerType)
}

func (a *resellerAccountingLedgerStoreAdapter) GetLedgerEntryByIdempotencyKey(key string) (*models.ResellerLedgerEntry, error) {
	return a.repo.GetLedgerEntryByIdempotencyKey(key)
}
