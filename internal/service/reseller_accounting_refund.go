package service

import (
	"github.com/dujiao-next/internal/models"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

func (s *ResellerAccountingService) HandleRefundDeductTx(tx *gorm.DB, order *models.Order, refundRecord *models.OrderRefundRecord, refundedBefore decimal.Decimal) error {
	if s == nil || s.repo == nil || s.ledger == nil || tx == nil || order == nil || refundRecord == nil || refundRecord.ID == 0 {
		return nil
	}
	return s.ledger.HandleRefundDeduct(newResellerAccountingLedgerStoreAdapter(s.repo.WithTx(tx)), order, refundRecord, refundedBefore)
}
