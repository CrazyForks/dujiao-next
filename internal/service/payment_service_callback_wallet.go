package service

import (
	"strings"
	"time"

	"github.com/dujiao-next/internal/constants"
	"github.com/dujiao-next/internal/models"
	walletcontract "github.com/dujiao-next/internal/modules/wallet/contract"
	walletgormstore "github.com/dujiao-next/internal/modules/wallet/infrastructure/gormstore"

	"gorm.io/gorm"
)

func (s *PaymentService) handleWalletRechargeCallback(payment *models.Payment, status string, input PaymentCallbackInput) (*models.Payment, error) {
	log := paymentLogger(
		"payment_id", payment.ID,
		"recharge_no", strings.TrimSpace(input.OrderNo),
		"target_status", status,
		"callback_channel_id", input.ChannelID,
		"callback_provider_ref", strings.TrimSpace(input.ProviderRef),
		"callback_currency", strings.ToUpper(strings.TrimSpace(input.Currency)),
		"callback_amount", input.Amount.String(),
	)
	if s.walletRepo == nil {
		log.Errorw("wallet_recharge_callback_wallet_repo_nil")
		return nil, ErrPaymentUpdateFailed
	}
	recharge, err := s.walletRepo.GetRechargeOrderByPaymentID(payment.ID)
	if err != nil {
		log.Errorw("wallet_recharge_callback_recharge_fetch_failed", "error", err)
		return nil, ErrPaymentUpdateFailed
	}
	if recharge == nil {
		log.Warnw("wallet_recharge_callback_recharge_not_found")
		return nil, walletcontract.ErrRechargeNotFound
	}

	if input.ChannelID != 0 && input.ChannelID != payment.ChannelID {
		log.Warnw("wallet_recharge_callback_channel_mismatch",
			"stored_channel_id", payment.ChannelID,
			"callback_channel_id", input.ChannelID,
		)
		return nil, ErrPaymentInvalid
	}
	if !matchesBusinessOrderNo(input.OrderNo, recharge.RechargeNo, payment) {
		log.Warnw("wallet_recharge_callback_order_no_mismatch",
			"stored_recharge_no", recharge.RechargeNo,
			"stored_gateway_order_no", payment.GatewayOrderNo,
			"callback_order_no", input.OrderNo,
		)
		return nil, ErrPaymentInvalid
	}
	if input.Currency != "" && !strings.EqualFold(strings.TrimSpace(input.Currency), strings.TrimSpace(payment.Currency)) {
		log.Warnw("wallet_recharge_callback_currency_mismatch",
			"stored_currency", payment.Currency,
			"callback_currency", input.Currency,
		)
		return nil, ErrPaymentCurrencyMismatch
	}
	if !input.Amount.Decimal.IsZero() && input.Amount.Decimal.Cmp(payment.Amount.Decimal) != 0 {
		log.Warnw("wallet_recharge_callback_amount_mismatch",
			"stored_amount", payment.Amount.String(),
			"callback_amount", input.Amount.String(),
		)
		return nil, ErrPaymentAmountMismatch
	}

	// 幂等处理：已成功状态仅更新回调元信息。
	if payment.Status == constants.PaymentStatusSuccess {
		log.Infow("wallet_recharge_callback_idempotent_success",
			"current_status", payment.Status,
		)
		return s.updateCallbackMeta(payment, constants.PaymentStatusSuccess, input)
	}
	if payment.Status == status {
		log.Infow("wallet_recharge_callback_idempotent_same_status",
			"current_status", payment.Status,
		)
		return s.updateCallbackMeta(payment, status, input)
	}
	if !canApplyWalletRechargeCallback(payment.Status, recharge.Status, status) {
		log.Infow("wallet_recharge_callback_ignored_terminal_transition",
			"current_payment_status", payment.Status,
			"current_recharge_status", recharge.Status,
			"target_status", status,
		)
		return s.updateCallbackMeta(payment, payment.Status, input)
	}

	now := time.Now()
	updated, err := s.applyWalletRechargePaymentUpdate(payment, status, input, now)
	if err != nil {
		log.Errorw("wallet_recharge_callback_apply_failed", "error", err)
		return nil, err
	}
	log.Infow("wallet_recharge_callback_processed",
		"new_status", updated.Status,
	)
	if updated.Status == constants.PaymentStatusSuccess {
		s.enqueueWalletRechargeSuccessAsync(recharge, updated, log)
		s.enqueueWalletRechargeBotNotifyAsync(recharge, log)
	}
	return updated, nil
}

func (s *PaymentService) applyWalletRechargePaymentUpdate(payment *models.Payment, status string, input PaymentCallbackInput, now time.Time) (*models.Payment, error) {
	paymentVal := payment

	switch status {
	case constants.PaymentStatusSuccess:
		paidAt := now
		if input.PaidAt != nil {
			paidAt = *input.PaidAt
		}
		payment.PaidAt = &paidAt
	case constants.PaymentStatusExpired:
		payment.ExpiredAt = &now
	}

	payment.Status = status
	payment.CallbackAt = &now
	payment.UpdatedAt = now
	if input.ProviderRef != "" {
		payment.ProviderRef = input.ProviderRef
	}
	if input.Payload != nil {
		payment.ProviderPayload = mergeProviderPayload(payment.ProviderPayload, input.Payload)
	}

	err := s.paymentRepo.Transaction(func(tx *gorm.DB) error {
		paymentRepo := s.paymentRepo.WithTx(tx)
		walletTx := walletgormstore.UseTransaction(tx)
		rechargeRepo := walletTx.Wallets()

		if err := paymentRepo.Update(payment); err != nil {
			return ErrPaymentUpdateFailed
		}
		recharge, err := rechargeRepo.GetRechargeOrderByPaymentIDForUpdate(payment.ID)
		if err != nil {
			return ErrPaymentUpdateFailed
		}
		if recharge == nil {
			return walletcontract.ErrRechargeNotFound
		}
		if recharge.Status == constants.WalletRechargeStatusSuccess {
			return nil
		}

		switch status {
		case constants.PaymentStatusSuccess:
			if s.walletSvc == nil {
				return walletcontract.ErrAccountNotFound
			}
			if _, err := s.walletSvc.ApplyRechargePayment(walletTx, recharge); err != nil {
				return err
			}
			recharge.Status = constants.WalletRechargeStatusSuccess
			paidAt := now
			if payment.PaidAt != nil {
				paidAt = *payment.PaidAt
			}
			recharge.PaidAt = &paidAt
		case constants.PaymentStatusFailed:
			recharge.Status = constants.WalletRechargeStatusFailed
		case constants.PaymentStatusExpired:
			recharge.Status = constants.WalletRechargeStatusExpired
		default:
			recharge.Status = constants.WalletRechargeStatusPending
		}
		recharge.UpdatedAt = now
		if err := rechargeRepo.UpdateRechargeOrder(recharge); err != nil {
			return ErrPaymentUpdateFailed
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	// 充值成功后触发会员等级升级检查（事务已提交）
	if status == constants.PaymentStatusSuccess && s.memberLevelSvc != nil {
		if recharge, _ := s.walletRepo.GetRechargeOrderByPaymentID(payment.ID); recharge != nil &&
			recharge.Status == constants.WalletRechargeStatusSuccess && recharge.UserID > 0 {
			if err := s.memberLevelSvc.OnRechargeCompleted(recharge.UserID, recharge.Amount.Decimal); err != nil {
				paymentLogger().Warnw("member_level_recharge_completed_failed",
					"payment_id", payment.ID,
					"user_id", recharge.UserID,
					"amount", recharge.Amount.Decimal.String(),
					"error", err,
				)
			}
		}
	}

	return paymentVal, nil
}

func canApplyWalletRechargeCallback(paymentStatus string, rechargeStatus string, targetStatus string) bool {
	// 成功回调允许覆盖终态（支付网关存在延迟成功通知场景）。
	if targetStatus == constants.PaymentStatusSuccess {
		return true
	}
	// 非成功回调不允许改变任何终态，避免 expired/failed/success 被回调串扰重开。
	if paymentStatus == constants.PaymentStatusSuccess || rechargeStatus == constants.WalletRechargeStatusSuccess {
		return false
	}
	if paymentStatus == constants.PaymentStatusFailed || rechargeStatus == constants.WalletRechargeStatusFailed {
		return false
	}
	if paymentStatus == constants.PaymentStatusExpired || rechargeStatus == constants.WalletRechargeStatusExpired {
		return false
	}
	return true
}
