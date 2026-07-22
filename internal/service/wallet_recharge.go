package service

import (
	"fmt"
	"time"

	"github.com/dujiao-next/internal/constants"
	"github.com/dujiao-next/internal/models"
	"github.com/dujiao-next/internal/shared/money"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// ApplyRechargePayment 在事务内确认充值到账并写入钱包流水
func (s *WalletService) ApplyRechargePayment(tx *gorm.DB, recharge *models.WalletRechargeOrder) (*models.WalletTransaction, error) {
	if tx == nil {
		return nil, ErrWalletRechargeStatusInvalid
	}
	if recharge == nil || recharge.ID == 0 || recharge.UserID == 0 {
		return nil, ErrWalletRechargeNotFound
	}
	amount := recharge.Amount.Decimal.Round(2)
	if amount.LessThanOrEqual(decimal.Zero) {
		return nil, ErrWalletInvalidAmount
	}

	reference := fmt.Sprintf("recharge:%d:success", recharge.ID)
	repo := s.walletRepo.WithTx(tx)
	exists, err := repo.GetTransactionByReference(reference)
	if err != nil {
		return nil, err
	}
	if exists != nil {
		return exists, nil
	}

	now := time.Now()
	account, err := s.ensureAccountForUpdate(repo, recharge.UserID, now)
	if err != nil {
		return nil, err
	}
	before := account.Balance.Decimal.Round(2)
	after := before.Add(amount).Round(2)
	account.Balance = money.FromDecimal(after)
	account.UpdatedAt = now
	if err := repo.UpdateAccount(account); err != nil {
		return nil, ErrWalletAccountUpdateFailed
	}

	txn := &models.WalletTransaction{
		UserID:        recharge.UserID,
		OrderID:       nil,
		Type:          constants.WalletTxnTypeRecharge,
		Direction:     constants.WalletTxnDirectionIn,
		Amount:        money.FromDecimal(amount),
		BalanceBefore: money.FromDecimal(before),
		BalanceAfter:  money.FromDecimal(after),
		Currency:      normalizeWalletCurrency(recharge.Currency),
		Reference:     reference,
		Remark:        cleanWalletRemark(recharge.Remark, "在线充值到账"),
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	if err := repo.CreateTransaction(txn); err != nil {
		return nil, ErrWalletTransactionCreateFailed
	}
	return txn, nil
}
