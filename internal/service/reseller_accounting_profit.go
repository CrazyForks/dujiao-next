package service

import (
	"time"

	"github.com/dujiao-next/internal/models"
	"gorm.io/gorm"
)

func (s *ResellerAccountingService) PostOrderProfitTx(tx *gorm.DB, order *models.Order, payment *models.Payment) error {
	if s == nil || s.repo == nil || s.ledger == nil || tx == nil || order == nil || order.ID == 0 {
		return nil
	}
	return s.ledger.PostOrderProfit(newResellerAccountingLedgerStoreAdapter(s.repo.WithTx(tx)), order, payment)
}

func (s *ResellerAccountingService) ConfirmDueLedgerEntries(now time.Time) (int64, error) {
	if s == nil || s.ledger == nil {
		return 0, nil
	}
	return s.ledger.ConfirmDueLedgerEntries(now)
}
