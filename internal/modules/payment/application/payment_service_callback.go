package application

import (
	"strings"
	"time"

	productcontract "github.com/dujiao-next/internal/modules/catalog/product/contract"
	orderapp "github.com/dujiao-next/internal/modules/order/application"
	orderdomain "github.com/dujiao-next/internal/modules/order/domain"
	paymentcontract "github.com/dujiao-next/internal/modules/payment/contract"
	paymentdomain "github.com/dujiao-next/internal/modules/payment/domain"

	"github.com/dujiao-next/internal/constants"
	"github.com/dujiao-next/internal/shared/jsonmap"
	"github.com/dujiao-next/internal/shared/money"
)

// PaymentCallbackInput 支付回调输入
type PaymentCallbackInput struct {
	PaymentID   uint
	OrderNo     string
	ChannelID   uint
	Status      string
	ProviderRef string
	Amount      money.Amount
	Currency    string
	PaidAt      *time.Time
	Payload     jsonmap.JSON
}

func (s *PaymentService) HandleCallback(input PaymentCallbackInput) (*paymentdomain.Payment, error) {
	if input.PaymentID == 0 {
		return nil, ErrPaymentInvalid
	}
	status := normalizePaymentStatus(input.Status)
	if !isPaymentStatusValid(status) {
		return nil, ErrPaymentStatusInvalid
	}

	log := paymentLogger(
		"payment_id", input.PaymentID,
		"target_status", status,
		"callback_channel_id", input.ChannelID,
		"callback_order_no", strings.TrimSpace(input.OrderNo),
		"callback_provider_ref", strings.TrimSpace(input.ProviderRef),
		"callback_currency", strings.ToUpper(strings.TrimSpace(input.Currency)),
		"callback_amount", input.Amount.String(),
	)
	log.Infow("payment_callback_received")

	payment, err := s.paymentRepo.GetByID(input.PaymentID)
	if err != nil {
		log.Errorw("payment_callback_payment_fetch_failed", "error", err)
		return nil, ErrPaymentUpdateFailed
	}
	if payment == nil {
		log.Warnw("payment_callback_payment_not_found")
		return nil, ErrPaymentNotFound
	}
	if payment.OrderID == 0 {
		log.Infow("payment_callback_wallet_recharge_flow")
		return s.handleWalletRechargeCallback(payment, status, input)
	}

	order, err := s.orderRepo.GetByID(payment.OrderID)
	if err != nil {
		log.Errorw("payment_callback_order_fetch_failed", "order_id", payment.OrderID, "error", err)
		return nil, orderapp.ErrOrderFetchFailed
	}
	if order == nil {
		log.Warnw("payment_callback_order_not_found", "order_id", payment.OrderID)
		return nil, orderapp.ErrOrderNotFound
	}

	if input.ChannelID != 0 && input.ChannelID != payment.ChannelID {
		log.Warnw("payment_callback_channel_mismatch",
			"stored_channel_id", payment.ChannelID,
			"callback_channel_id", input.ChannelID,
		)
		return nil, ErrPaymentInvalid
	}
	if !matchesBusinessOrderNo(input.OrderNo, order.OrderNo, payment) {
		log.Warnw("payment_callback_order_no_mismatch",
			"stored_order_no", order.OrderNo,
			"stored_gateway_order_no", payment.GatewayOrderNo,
			"callback_order_no", input.OrderNo,
		)
		return nil, ErrPaymentInvalid
	}
	if input.Currency != "" && !strings.EqualFold(strings.TrimSpace(input.Currency), strings.TrimSpace(payment.Currency)) {
		log.Warnw("payment_callback_currency_mismatch",
			"stored_currency", payment.Currency,
			"callback_currency", input.Currency,
		)
		return nil, ErrPaymentCurrencyMismatch
	}
	if !input.Amount.Decimal.IsZero() && input.Amount.Decimal.Cmp(payment.Amount.Decimal) != 0 {
		log.Warnw("payment_callback_amount_mismatch",
			"stored_amount", payment.Amount.String(),
			"callback_amount", input.Amount.String(),
		)
		return nil, ErrPaymentAmountMismatch
	}

	// 幂等处理：已成功的不再回退状态
	if payment.Status == constants.PaymentStatusSuccess {
		log.Infow("payment_callback_idempotent_success",
			"current_status", payment.Status,
		)
		return s.updateCallbackMeta(payment, constants.PaymentStatusSuccess, input)
	}
	if payment.Status == status {
		log.Infow("payment_callback_idempotent_same_status",
			"current_status", payment.Status,
		)
		return s.updateCallbackMeta(payment, status, input)
	}

	previousStatus := payment.Status
	now := time.Now()
	updated, orderPaid, err := s.applyPaymentUpdate(payment, order, status, input, now)
	if err != nil {
		log.Errorw("payment_callback_apply_failed",
			"order_id", order.ID,
			"order_no", order.OrderNo,
			"current_status", payment.Status,
			"error", err,
		)
		return nil, err
	}
	if orderPaid {
		s.enqueueOrderPaidAsync(order, updated, log)
	}
	log.Infow("payment_callback_processed",
		"order_id", order.ID,
		"order_no", order.OrderNo,
		"previous_status", previousStatus,
		"new_status", updated.Status,
		"order_paid", orderPaid,
	)
	return updated, nil
}

func (s *PaymentService) updateCallbackMeta(payment *paymentdomain.Payment, status string, input PaymentCallbackInput) (*paymentdomain.Payment, error) {
	updated := false
	if input.ProviderRef != "" && payment.ProviderRef == "" {
		payment.ProviderRef = input.ProviderRef
		updated = true
	}
	if input.Payload != nil {
		payment.ProviderPayload = mergeProviderPayload(payment.ProviderPayload, input.Payload)
		updated = true
	}
	if status != "" && payment.Status != status {
		payment.Status = status
		updated = true
	}
	if payment.Status == constants.PaymentStatusSuccess && payment.PaidAt == nil && input.PaidAt != nil {
		payment.PaidAt = input.PaidAt
		updated = true
	}
	if updated {
		now := time.Now()
		payment.CallbackAt = &now
		payment.UpdatedAt = now
		if err := s.paymentRepo.Update(payment); err != nil {
			return nil, ErrPaymentUpdateFailed
		}
	}
	return payment, nil
}

func (s *PaymentService) applyPaymentUpdate(payment *paymentdomain.Payment, order *orderdomain.Order, status string, input PaymentCallbackInput, now time.Time) (*paymentdomain.Payment, bool, error) {
	returnVal := payment
	orderPaid := false

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

	err := s.paymentRepo.WithinTransaction(func(tx paymentcontract.Transaction) error {
		paymentRepo := tx.Payments()

		if err := paymentRepo.Update(payment); err != nil {
			return ErrPaymentUpdateFailed
		}

		if status == constants.PaymentStatusSuccess && order.Status != constants.OrderStatusPaid {
			if err := s.markOrderPaid(tx, order, now); err != nil {
				return err
			}
			if s.resellerAccounting != nil {
				if err := s.resellerAccounting.PostOrderProfit(tx.ResellerAccounting(), order, payment); err != nil {
					return err
				}
			}
			orderPaid = true
		}
		if (status == constants.PaymentStatusFailed || status == constants.PaymentStatusExpired) && order.Status == constants.OrderStatusPendingPayment && s.walletSvc != nil {
			if _, err := orderapp.ReleaseWalletBalance(s.walletSvc, tx, order, constants.WalletTxnTypeOrderRefund, "在线支付失败，退回余额"); err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return nil, false, err
	}
	return returnVal, orderPaid, nil
}

// mergeProviderPayload 合并第三方回调原文，同时保留创建支付阶段写入的展示快照等元数据。
// 回调字段优先覆盖同名旧字段，未出现在回调中的 display_channel_type 等字段不会丢失。
func mergeProviderPayload(existing jsonmap.JSON, incoming jsonmap.JSON) jsonmap.JSON {
	if incoming == nil {
		return existing
	}
	merged := make(jsonmap.JSON, len(existing)+len(incoming))
	for key, value := range existing {
		merged[key] = value
	}
	for key, value := range incoming {
		merged[key] = value
	}
	return merged
}

// markOrderPaid 在事务内将订单更新为已支付并处理库存
func (s *PaymentService) markOrderPaid(tx paymentcontract.Transaction, order *orderdomain.Order, now time.Time) error {
	if order == nil {
		return orderapp.ErrOrderNotFound
	}
	if !orderapp.IsTransitionAllowed(order.Status, constants.OrderStatusPaid) {
		return orderapp.ErrOrderStatusInvalid
	}
	orderRepo := tx.Orders()
	productRepo := tx.Products()
	var productSKURepo productcontract.SKURepository
	if s.productSKURepo != nil {
		productSKURepo = tx.ProductSKUs()
	}

	onlineAmount := normalizeOrderAmount(order.TotalAmount.Decimal.Sub(order.WalletPaidAmount.Decimal))
	orderUpdates := map[string]interface{}{
		"paid_at":            now,
		"online_paid_amount": money.FromDecimal(onlineAmount),
		"updated_at":         now,
	}
	if err := orderRepo.UpdateStatus(order.ID, constants.OrderStatusPaid, orderUpdates); err != nil {
		return orderapp.ErrOrderUpdateFailed
	}
	order.Status = constants.OrderStatusPaid
	order.PaidAt = &now
	order.OnlinePaidAmount = money.FromDecimal(onlineAmount)
	order.UpdatedAt = now

	if len(order.Children) > 0 {
		for idx := range order.Children {
			child := &order.Children[idx]
			childStatus := constants.OrderStatusPaid
			if shouldMarkFulfilling(child) {
				childStatus = constants.OrderStatusFulfilling
			}
			if err := orderRepo.UpdateStatus(child.ID, childStatus, map[string]interface{}{
				"paid_at":    now,
				"updated_at": now,
			}); err != nil {
				return orderapp.ErrOrderUpdateFailed
			}
			if err := orderapp.ConsumeManualStockByItems(productRepo, productSKURepo, child.Items); err != nil {
				return err
			}
			child.Status = childStatus
			child.PaidAt = &now
			child.UpdatedAt = now
		}
		parentStatus := orderapp.CalcParentStatus(order.Children, constants.OrderStatusPaid)
		if parentStatus != "" && parentStatus != constants.OrderStatusPaid {
			if err := orderRepo.UpdateStatus(order.ID, parentStatus, map[string]interface{}{
				"online_paid_amount": money.FromDecimal(onlineAmount),
				"updated_at":         now,
			}); err != nil {
				return orderapp.ErrOrderUpdateFailed
			}
			order.Status = parentStatus
		}
		return nil
	}

	if err := orderapp.ConsumeManualStockByItems(productRepo, productSKURepo, order.Items); err != nil {
		return err
	}
	return nil
}
