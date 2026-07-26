package service

import (
	"context"
	"strings"
	"time"

	"github.com/dujiao-next/internal/constants"
	"github.com/dujiao-next/internal/logger"
	"github.com/dujiao-next/internal/models"
	walletcontract "github.com/dujiao-next/internal/modules/wallet/contract"
	walletdomain "github.com/dujiao-next/internal/modules/wallet/domain"
	walletgormstore "github.com/dujiao-next/internal/modules/wallet/infrastructure/gormstore"
	"github.com/dujiao-next/internal/queue"
	"github.com/dujiao-next/internal/shared/money"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// CreateWalletRechargePaymentInput 创建钱包充值支付请求
type CreateWalletRechargePaymentInput struct {
	UserID        uint
	ChannelID     uint
	Amount        money.Amount
	Currency      string
	Remark        string
	ClientIP      string
	Context       context.Context
	RequestScheme string
}

// CreateWalletRechargePaymentResult 创建钱包充值支付结果
type CreateWalletRechargePaymentResult struct {
	Recharge *walletdomain.RechargeOrder
	Payment  *models.Payment
}

// CreateWalletRechargePayment 创建钱包充值支付单
func (s *PaymentService) CreateWalletRechargePayment(input CreateWalletRechargePaymentInput) (*CreateWalletRechargePaymentResult, error) {
	if input.UserID == 0 || input.ChannelID == 0 {
		return nil, ErrPaymentInvalid
	}
	amount := input.Amount.Decimal.Round(2)
	if amount.LessThanOrEqual(decimal.Zero) {
		return nil, walletcontract.ErrInvalidAmount
	}
	if s.walletRepo == nil {
		return nil, ErrPaymentCreateFailed
	}

	channel, err := s.channelRepo.GetByID(input.ChannelID)
	if err != nil {
		return nil, err
	}
	if channel == nil {
		return nil, ErrPaymentChannelNotFound
	}
	if !channel.IsActive {
		return nil, ErrPaymentChannelInactive
	}

	// 校验钱包充值是否允许该支付渠道
	if err := s.validateWalletRechargeChannel(channel.ID); err != nil {
		return nil, err
	}

	feeRate := channel.FeeRate.Decimal.Round(2)
	if feeRate.LessThan(decimal.Zero) || feeRate.GreaterThan(decimal.NewFromInt(100)) {
		return nil, ErrPaymentChannelConfigInvalid
	}
	fixedFee := channel.FixedFee.Decimal.Round(2)
	if fixedFee.LessThan(decimal.Zero) || fixedFee.GreaterThanOrEqual(decimal.NewFromInt(10000)) {
		return nil, ErrPaymentChannelConfigInvalid
	}
	if err := validatePaymentAmountForChannel(amount, channel); err != nil {
		return nil, err
	}

	feeAmount := fixedFee
	if feeRate.GreaterThan(decimal.Zero) {
		feeAmount = feeAmount.Add(amount.Mul(feeRate).Div(decimal.NewFromInt(100))).Round(2)
	}
	payableAmount := amount.Add(feeAmount).Round(2)
	currency := strings.ToUpper(strings.TrimSpace(input.Currency))
	if currency == "" {
		currency = "CNY"
	}
	if err := validatePaymentCurrencyForChannel(currency, channel); err != nil {
		return nil, err
	}
	if shouldUseCNYPaymentCurrency(channel) {
		currency = "CNY"
	}
	now := time.Now()

	var payment *models.Payment
	var recharge *walletdomain.RechargeOrder
	err = s.paymentRepo.Transaction(func(tx *gorm.DB) error {
		rechargeNo := generateWalletRechargeNo()
		paymentRepo := s.paymentRepo.WithTx(tx)
		payment = &models.Payment{
			OrderID:         0,
			ChannelID:       channel.ID,
			ProviderType:    channel.ProviderType,
			ChannelType:     channel.ChannelType,
			InteractionMode: channel.InteractionMode,
			Amount:          money.FromDecimal(payableAmount),
			FeeRate:         money.FromDecimal(feeRate),
			FixedFee:        money.FromDecimal(fixedFee),
			FeeAmount:       money.FromDecimal(feeAmount),
			Currency:        currency,
			Status:          constants.PaymentStatusInitiated,
			CreatedAt:       now,
			UpdatedAt:       now,
		}
		if err := paymentRepo.Create(payment); err != nil {
			return ErrPaymentCreateFailed
		}

		rechargeRepo := walletgormstore.UseTransaction(tx).Wallets()
		remark := strings.TrimSpace(input.Remark)
		if remark == "" {
			remark = "余额充值"
		}
		recharge = &walletdomain.RechargeOrder{
			RechargeNo:      rechargeNo,
			UserID:          input.UserID,
			PaymentID:       payment.ID,
			ChannelID:       channel.ID,
			ProviderType:    channel.ProviderType,
			ChannelType:     channel.ChannelType,
			InteractionMode: channel.InteractionMode,
			Amount:          money.FromDecimal(amount),
			PayableAmount:   money.FromDecimal(payableAmount),
			FeeRate:         money.FromDecimal(feeRate),
			FeeAmount:       money.FromDecimal(feeAmount),
			Currency:        currency,
			Status:          constants.WalletRechargeStatusPending,
			Remark:          remark,
			CreatedAt:       now,
			UpdatedAt:       now,
		}
		if err := rechargeRepo.CreateRechargeOrder(recharge); err != nil {
			return ErrPaymentCreateFailed
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	if payment == nil || recharge == nil {
		return nil, ErrPaymentCreateFailed
	}

	// 复用支付网关下单逻辑，使用充值单号作为业务单号。
	virtualOrder := &models.Order{
		OrderNo: recharge.RechargeNo,
		UserID:  recharge.UserID,
	}
	if err := s.applyProviderPayment(CreatePaymentInput{
		ChannelID:        input.ChannelID,
		ClientIP:         input.ClientIP,
		Context:          input.Context,
		ReturnBizType:    "recharge",
		ReturnBusinessNo: recharge.RechargeNo,
		RequestScheme:    input.RequestScheme,
	}, virtualOrder, channel, payment); err != nil {
		_ = s.paymentRepo.Transaction(func(tx *gorm.DB) error {
			rechargeRepo := walletgormstore.UseTransaction(tx).Wallets()
			paymentRepo := s.paymentRepo.WithTx(tx)
			failedAt := time.Now()
			payment.Status = constants.PaymentStatusFailed
			payment.UpdatedAt = failedAt
			if updateErr := paymentRepo.Update(payment); updateErr != nil {
				return updateErr
			}
			lockedRecharge, getErr := rechargeRepo.GetRechargeOrderByPaymentIDForUpdate(payment.ID)
			if getErr != nil || lockedRecharge == nil {
				return getErr
			}
			lockedRecharge.Status = constants.WalletRechargeStatusFailed
			lockedRecharge.UpdatedAt = failedAt
			return rechargeRepo.UpdateRechargeOrder(lockedRecharge)
		})
		return nil, err
	}
	if s.queueClient != nil {
		delay := time.Duration(s.resolveExpireMinutes()) * time.Minute
		if err := s.queueClient.EnqueueWalletRechargeExpire(queue.WalletRechargeExpirePayload{
			PaymentID: payment.ID,
		}, delay); err != nil {
			logger.Errorw("wallet_recharge_enqueue_timeout_expire_failed",
				"payment_id", payment.ID,
				"recharge_no", recharge.RechargeNo,
				"delay_minutes", int(delay/time.Minute),
				"error", err,
			)
			_ = s.paymentRepo.Transaction(func(tx *gorm.DB) error {
				rechargeRepo := walletgormstore.UseTransaction(tx).Wallets()
				paymentRepo := s.paymentRepo.WithTx(tx)
				failedAt := time.Now()
				payment.Status = constants.PaymentStatusFailed
				payment.UpdatedAt = failedAt
				if updateErr := paymentRepo.Update(payment); updateErr != nil {
					return updateErr
				}
				lockedRecharge, getErr := rechargeRepo.GetRechargeOrderByPaymentIDForUpdate(payment.ID)
				if getErr != nil || lockedRecharge == nil {
					return getErr
				}
				if lockedRecharge.Status == constants.WalletRechargeStatusSuccess {
					return nil
				}
				lockedRecharge.Status = constants.WalletRechargeStatusFailed
				lockedRecharge.UpdatedAt = failedAt
				return rechargeRepo.UpdateRechargeOrder(lockedRecharge)
			})
			return nil, ErrQueueUnavailable
		}
	}

	reloadedRecharge, err := s.walletRepo.GetRechargeOrderByPaymentID(payment.ID)
	if err != nil {
		return nil, ErrPaymentUpdateFailed
	}
	if reloadedRecharge != nil {
		recharge = reloadedRecharge
	}
	return &CreateWalletRechargePaymentResult{
		Recharge: recharge,
		Payment:  payment,
	}, nil
}

func generateWalletRechargeNo() string {
	return generateSerialNo("WR")
}
