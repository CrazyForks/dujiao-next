package reseller

import (
	"testing"

	"github.com/dujiao-next/internal/models"
	"github.com/shopspring/decimal"
)

func TestAccountingWithdrawServiceRejectsInvalidAmount(t *testing.T) {
	svc := NewAccountingWithdrawService(nil)
	_, err := svc.ApplyWithdraw(1, WithdrawApplyInput{
		Amount:   decimal.Zero,
		Currency: "USD",
		Channel:  "usdt",
		Account:  "T",
	})
	if err != ErrAccountingUnavailable {
		t.Fatalf("expected unavailable with nil store, got %v", err)
	}

	svc = NewAccountingWithdrawService(accountingWithdrawStoreStub{})
	_, err = svc.ApplyWithdraw(1, WithdrawApplyInput{
		Amount:   decimal.Zero,
		Currency: "USD",
		Channel:  "usdt",
		Account:  "T",
	})
	if err != ErrWithdrawAmountInvalid {
		t.Fatalf("expected amount invalid, got %v", err)
	}
}

type accountingWithdrawStoreStub struct{}

func (accountingWithdrawStoreStub) WithinTransaction(fn func(store AccountingWithdrawStore) error) error {
	return fn(accountingWithdrawStoreStub{})
}
func (accountingWithdrawStoreStub) GetProfileByUserID(userID uint) (*models.ResellerProfile, error) {
	return nil, nil
}
func (accountingWithdrawStoreStub) GetOrCreateBalanceAccountForUpdate(resellerID uint, currency string) (*models.ResellerBalanceAccount, error) {
	return nil, nil
}
func (accountingWithdrawStoreStub) SumLedgerAmountGroupedByStatus(resellerID uint, currency string, statuses []string) (map[string]decimal.Decimal, error) {
	return map[string]decimal.Decimal{}, nil
}
func (accountingWithdrawStoreStub) UpdateBalanceAccount(account *models.ResellerBalanceAccount) error {
	return nil
}
func (accountingWithdrawStoreStub) ListAvailableLedgerEntriesForUpdate(resellerID uint, currency string) ([]models.ResellerLedgerEntry, error) {
	return nil, nil
}
func (accountingWithdrawStoreStub) UpdateLedgerEntry(entry *models.ResellerLedgerEntry) error {
	return nil
}
func (accountingWithdrawStoreStub) CreateLedgerEntryIfNotExists(entry *models.ResellerLedgerEntry) (bool, error) {
	return false, nil
}
func (accountingWithdrawStoreStub) BatchUpdateLedgerEntries(ids []uint, updates map[string]interface{}) error {
	return nil
}
func (accountingWithdrawStoreStub) BatchUpdateLedgerEntriesByWithdrawID(withdrawID uint, updates map[string]interface{}) error {
	return nil
}
func (accountingWithdrawStoreStub) CreateWithdrawRequest(req *models.ResellerWithdrawRequest) error {
	return nil
}
func (accountingWithdrawStoreStub) GetWithdrawRequestByID(id uint) (*models.ResellerWithdrawRequest, error) {
	return nil, nil
}
func (accountingWithdrawStoreStub) GetWithdrawRequestByIDForUpdate(id uint) (*models.ResellerWithdrawRequest, error) {
	return nil, nil
}
func (accountingWithdrawStoreStub) UpdateWithdrawRequest(req *models.ResellerWithdrawRequest) error {
	return nil
}
