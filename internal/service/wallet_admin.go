package service

import (
	"fmt"
	"strings"
	"time"

	settingsapp "github.com/dujiao-next/internal/modules/settings/application"
	"github.com/dujiao-next/internal/shared/money"

	"github.com/dujiao-next/internal/constants"
	"github.com/dujiao-next/internal/models"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// WalletRechargeInput 用户充值输入
type WalletRechargeInput struct {
	UserID   uint
	Amount   money.Amount
	Currency string
	Remark   string
}

// WalletAdjustInput 管理员余额调整输入
type WalletAdjustInput struct {
	UserID   uint
	Delta    money.Amount
	Currency string
	Remark   string
}

// AdminRefundToWalletInput 管理员退款到余额输入
type AdminRefundToWalletInput struct {
	OrderID uint
	Amount  money.Amount
	Remark  string
}

// Recharge 用户充值余额
func (s *WalletService) Recharge(input WalletRechargeInput) (*models.WalletAccount, *models.WalletTransaction, error) {
	if input.UserID == 0 {
		return nil, nil, ErrWalletAccountNotFound
	}
	amount := input.Amount.Decimal.Round(2)
	if amount.LessThanOrEqual(decimal.Zero) {
		return nil, nil, ErrWalletInvalidAmount
	}
	reference := buildWalletReference("recharge", input.UserID)
	remark := cleanWalletRemark(input.Remark, "用户充值")
	currency := normalizeWalletCurrency(input.Currency)
	return s.changeBalance(input.UserID, amount, constants.WalletTxnTypeRecharge, nil, reference, remark, currency)
}

// AdminAdjustBalance 管理员增减用户余额
func (s *WalletService) AdminAdjustBalance(input WalletAdjustInput) (*models.WalletAccount, *models.WalletTransaction, error) {
	if input.UserID == 0 {
		return nil, nil, ErrWalletAccountNotFound
	}
	delta := input.Delta.Decimal.Round(2)
	if delta.IsZero() {
		return nil, nil, ErrWalletInvalidAmount
	}
	reference := buildWalletReference("admin_adjust", input.UserID)
	remark := cleanWalletRemark(input.Remark, "管理员调整余额")
	currency := normalizeWalletCurrency(input.Currency)
	return s.changeBalance(input.UserID, delta, constants.WalletTxnTypeAdminAdjust, nil, reference, remark, currency)
}

// AdminRefundToWallet 管理端订单退款到余额
func (s *WalletService) AdminRefundToWallet(input AdminRefundToWalletInput) (*models.Order, *models.WalletTransaction, *models.OrderRefundRecord, error) {
	if input.OrderID == 0 {
		return nil, nil, nil, ErrOrderNotFound
	}
	amount := input.Amount.Decimal.Round(2)
	if amount.LessThanOrEqual(decimal.Zero) {
		return nil, nil, nil, ErrWalletInvalidAmount
	}
	reference := buildWalletReference(fmt.Sprintf("order:%d:admin_refund", input.OrderID), input.OrderID)
	remark := cleanWalletRemark(input.Remark, "管理员退款到余额")
	recordRemark := strings.TrimSpace(input.Remark)

	var txnResult *models.WalletTransaction
	var refundRecordResult *models.OrderRefundRecord

	cfg := settingsapp.DefaultOrderRefundConfig()
	if s.settingService != nil {
		cfgLoaded, cfgErr := s.settingService.GetOrderRefundConfig()
		if cfgErr != nil {
			return nil, nil, nil, cfgErr
		}
		cfg = cfgLoaded
	}

	if err := s.walletRepo.Transaction(func(tx *gorm.DB) error {
		locked, err := s.orderRepo.WithTx(tx).GetByIDForUpdate(input.OrderID)
		if err != nil {
			return err
		}
		if locked == nil {
			return ErrOrderNotFound
		}
		order := *locked
		if order.UserID == 0 {
			return ErrWalletNotSupportedForGuest
		}
		if order.PaidAt == nil {
			return ErrOrderStatusInvalid
		}
		if settingsapp.IsOrderRefundWindowExpired(order.CreatedAt, order.PaidAt, cfg.MaxRefundDays, time.Now()) {
			return ErrOrderRefundExpired
		}
		if order.TotalAmount.Decimal.LessThanOrEqual(decimal.Zero) {
			return ErrOrderStatusInvalid
		}
		refundedBefore := order.RefundedAmount.Decimal.Round(2)
		refundable := order.TotalAmount.Decimal.Sub(refundedBefore).Round(2)
		if amount.GreaterThan(refundable) {
			return ErrWalletRefundExceeded
		}

		repo := s.walletRepo.WithTx(tx)
		account, err := s.ensureAccountForUpdate(repo, order.UserID, time.Now())
		if err != nil {
			return err
		}
		before := account.Balance.Decimal.Round(2)
		after := before.Add(amount).Round(2)
		account.Balance = money.FromDecimal(after)
		account.UpdatedAt = time.Now()
		if err := repo.UpdateAccount(account); err != nil {
			return ErrWalletAccountUpdateFailed
		}

		newRefunded := refundedBefore.Add(amount).Round(2)
		now := time.Now()
		updates := map[string]interface{}{
			"refunded_amount": money.FromDecimal(newRefunded),
			"updated_at":      now,
		}
		markRefunded := newRefunded.GreaterThanOrEqual(order.TotalAmount.Decimal.Round(2))
		if markRefunded {
			updates["status"] = constants.OrderStatusRefunded
		} else {
			updates["status"] = constants.OrderStatusPartiallyRefunded
		}
		if err := s.orderRepo.WithTx(tx).UpdateFields(order.ID, updates); err != nil {
			return ErrOrderUpdateFailed
		}
		if order.ParentID == nil {
			targetStatus := constants.OrderStatusPartiallyRefunded
			if markRefunded {
				targetStatus = constants.OrderStatusRefunded
			}
			if err := applyParentRefundChildStatusUpdates(s.orderRepo.WithTx(tx), order.ID, targetStatus, now); err != nil {
				return ErrOrderUpdateFailed
			}
		}
		if order.ParentID != nil {
			if _, err := syncParentStatus(s.orderRepo.WithTx(tx), *order.ParentID, now); err != nil {
				return ErrOrderUpdateFailed
			}
		}
		if s.affiliateSvc != nil {
			if err := s.affiliateSvc.HandleOrderRefundedTx(
				tx,
				&order,
				amount,
				refundedBefore,
				"order_refunded_to_wallet",
			); err != nil {
				return err
			}
		}

		txn := &models.WalletTransaction{
			UserID:        order.UserID,
			OrderID:       &order.ID,
			Type:          constants.WalletTxnTypeAdminRefund,
			Direction:     constants.WalletTxnDirectionIn,
			Amount:        money.FromDecimal(amount),
			BalanceBefore: money.FromDecimal(before),
			BalanceAfter:  money.FromDecimal(after),
			Currency:      normalizeWalletCurrency(order.Currency),
			Reference:     reference,
			Remark:        remark,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}
		if err := repo.CreateTransaction(txn); err != nil {
			return ErrWalletTransactionCreateFailed
		}
		record := &models.OrderRefundRecord{
			UserID:     order.UserID,
			GuestEmail: order.GuestEmail,
			OrderID:    order.ID,
			Type:       constants.OrderRefundTypeWallet,
			Amount:     money.FromDecimal(amount),
			Currency:   normalizeWalletCurrency(order.Currency),
			Remark:     recordRemark,
			CreatedAt:  now,
			UpdatedAt:  now,
		}
		if err := s.refundRecordRepo.WithTx(tx).Create(record); err != nil {
			return ErrRefundRecordCreateFailed
		}
		if s.resellerAccountingSvc != nil {
			if err := s.resellerAccountingSvc.HandleRefundDeductTx(tx, &order, record, refundedBefore); err != nil {
				return err
			}
		}
		txnResult = txn
		refundRecordResult = record
		return nil
	}); err != nil {
		return nil, nil, nil, err
	}

	order, err := s.orderRepo.GetByID(input.OrderID)
	if err != nil {
		return nil, nil, nil, ErrOrderFetchFailed
	}
	if order == nil {
		return nil, nil, nil, ErrOrderNotFound
	}
	return order, txnResult, refundRecordResult, nil
}
