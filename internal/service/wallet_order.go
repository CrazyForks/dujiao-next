package service

import (
	"time"

	"github.com/dujiao-next/internal/constants"
	"github.com/dujiao-next/internal/models"
	"github.com/dujiao-next/internal/shared/money"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// ApplyOrderBalance 在事务内为订单扣减余额并记录流水，返回扣减金额
func (s *WalletService) ApplyOrderBalance(tx *gorm.DB, order *models.Order, useBalance bool) (decimal.Decimal, error) {
	if tx == nil {
		return decimal.Zero, ErrOrderUpdateFailed
	}
	if order == nil {
		return decimal.Zero, ErrOrderNotFound
	}
	if !useBalance {
		return order.WalletPaidAmount.Decimal.Round(2), nil
	}
	if order.UserID == 0 {
		return decimal.Zero, ErrWalletNotSupportedForGuest
	}
	existing := order.WalletPaidAmount.Decimal.Round(2)
	if existing.GreaterThan(decimal.Zero) {
		return existing, nil
	}
	if order.TotalAmount.Decimal.LessThanOrEqual(decimal.Zero) {
		return decimal.Zero, nil
	}

	now := time.Now()
	repo := s.walletRepo.WithTx(tx)
	account, err := s.ensureAccountForUpdate(repo, order.UserID, now)
	if err != nil {
		return decimal.Zero, err
	}

	available := account.Balance.Decimal.Round(2)
	if available.LessThanOrEqual(decimal.Zero) {
		return decimal.Zero, nil
	}
	deduct := available
	if deduct.GreaterThan(order.TotalAmount.Decimal) {
		deduct = order.TotalAmount.Decimal.Round(2)
	}
	if deduct.LessThanOrEqual(decimal.Zero) {
		return decimal.Zero, nil
	}

	reference := buildOrderWalletReference(order.ID, constants.WalletTxnTypeOrderPay)
	exists, err := repo.GetTransactionByReference(reference)
	if err != nil {
		return decimal.Zero, err
	}
	if exists != nil {
		return exists.Amount.Decimal.Round(2), nil
	}

	before := account.Balance.Decimal.Round(2)
	after := before.Sub(deduct).Round(2)
	if after.LessThan(decimal.Zero) {
		return decimal.Zero, ErrWalletInsufficientBalance
	}
	account.Balance = money.FromDecimal(after)
	account.UpdatedAt = now
	if err := repo.UpdateAccount(account); err != nil {
		return decimal.Zero, ErrWalletAccountUpdateFailed
	}

	txn := &models.WalletTransaction{
		UserID:        order.UserID,
		OrderID:       &order.ID,
		Type:          constants.WalletTxnTypeOrderPay,
		Direction:     constants.WalletTxnDirectionOut,
		Amount:        money.FromDecimal(deduct),
		BalanceBefore: money.FromDecimal(before),
		BalanceAfter:  money.FromDecimal(after),
		Currency:      normalizeWalletCurrency(order.Currency),
		Reference:     reference,
		Remark:        "订单余额支付",
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	if err := repo.CreateTransaction(txn); err != nil {
		return decimal.Zero, ErrWalletTransactionCreateFailed
	}

	onlineAmount := normalizeOrderAmount(order.TotalAmount.Decimal.Sub(deduct))
	if err := s.orderRepo.WithTx(tx).UpdateFields(order.ID, map[string]interface{}{
		"wallet_paid_amount": money.FromDecimal(deduct),
		"online_paid_amount": money.FromDecimal(onlineAmount),
		"updated_at":         now,
	}); err != nil {
		return decimal.Zero, ErrOrderUpdateFailed
	}
	order.WalletPaidAmount = money.FromDecimal(deduct)
	order.OnlinePaidAmount = money.FromDecimal(onlineAmount)
	order.UpdatedAt = now
	return deduct, nil
}

// ReleaseOrderBalance 在事务内将订单已扣余额退回钱包，返回退回金额
func (s *WalletService) ReleaseOrderBalance(tx *gorm.DB, order *models.Order, txnType string, remark string) (decimal.Decimal, error) {
	if tx == nil {
		return decimal.Zero, ErrOrderUpdateFailed
	}
	if order == nil || order.UserID == 0 {
		return decimal.Zero, nil
	}
	amount := order.WalletPaidAmount.Decimal.Round(2)
	if amount.LessThanOrEqual(decimal.Zero) {
		return decimal.Zero, nil
	}
	now := time.Now()
	reference := buildOrderWalletReference(order.ID, txnType)
	repo := s.walletRepo.WithTx(tx)

	exists, err := repo.GetTransactionByReference(reference)
	if err != nil {
		return decimal.Zero, err
	}
	if exists != nil {
		return exists.Amount.Decimal.Round(2), nil
	}

	affected, err := s.orderRepo.WithTx(tx).UpdateFieldsWhereWalletPaid(order.ID, map[string]interface{}{
		"wallet_paid_amount": money.FromDecimal(decimal.Zero),
		"online_paid_amount": money.FromDecimal(order.TotalAmount.Decimal.Round(2)),
		"updated_at":         now,
	})
	if err != nil {
		return decimal.Zero, ErrOrderUpdateFailed
	}
	if affected == 0 {
		return decimal.Zero, nil
	}

	account, err := s.ensureAccountForUpdate(repo, order.UserID, now)
	if err != nil {
		return decimal.Zero, err
	}
	before := account.Balance.Decimal.Round(2)
	after := before.Add(amount).Round(2)
	account.Balance = money.FromDecimal(after)
	account.UpdatedAt = now
	if err := repo.UpdateAccount(account); err != nil {
		return decimal.Zero, ErrWalletAccountUpdateFailed
	}

	txn := &models.WalletTransaction{
		UserID:        order.UserID,
		OrderID:       &order.ID,
		Type:          txnType,
		Direction:     constants.WalletTxnDirectionIn,
		Amount:        money.FromDecimal(amount),
		BalanceBefore: money.FromDecimal(before),
		BalanceAfter:  money.FromDecimal(after),
		Currency:      normalizeWalletCurrency(order.Currency),
		Reference:     reference,
		Remark:        cleanWalletRemark(remark, "订单余额退回"),
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	if err := repo.CreateTransaction(txn); err != nil {
		return decimal.Zero, ErrWalletTransactionCreateFailed
	}

	order.WalletPaidAmount = money.FromDecimal(decimal.Zero)
	order.OnlinePaidAmount = money.FromDecimal(order.TotalAmount.Decimal.Round(2))
	order.UpdatedAt = now
	return amount, nil
}
