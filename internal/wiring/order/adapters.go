package orderwiring

import (
	"errors"
	"fmt"

	"github.com/dujiao-next/internal/constants"
	"github.com/dujiao-next/internal/http/handlers/shared"
	"github.com/dujiao-next/internal/models"
	"github.com/dujiao-next/internal/modules/captcha"
	"github.com/dujiao-next/internal/modules/coupon"
	"github.com/dujiao-next/internal/modules/orderrisk"
	"github.com/dujiao-next/internal/modules/promotion"
	"github.com/dujiao-next/internal/modules/reseller"
	"github.com/dujiao-next/internal/queue"
	"github.com/dujiao-next/internal/repository"
	"github.com/dujiao-next/internal/service"
	ordertransport "github.com/dujiao-next/internal/transport/http/order"
)

type orderAdminQueryAdapter struct {
	orders *service.OrderService
}

func (a orderAdminQueryAdapter) ListOrdersForAdmin(filter ordertransport.OrderListFilter) ([]models.Order, int64, error) {
	return a.orders.ListOrdersForAdmin(repository.OrderListFilter{
		Page:           filter.Page,
		PageSize:       filter.PageSize,
		UserID:         filter.UserID,
		UserKeyword:    filter.UserKeyword,
		Status:         filter.Status,
		OrderNo:        filter.OrderNo,
		GuestEmail:     filter.GuestEmail,
		ProductKeyword: filter.ProductKeyword,
		CreatedFrom:    filter.CreatedFrom,
		CreatedTo:      filter.CreatedTo,
		SortBy:         filter.SortBy,
		SortOrder:      filter.SortOrder,
	})
}

func (a orderAdminQueryAdapter) GetOrderForAdmin(orderID uint) (*models.Order, error) {
	order, err := a.orders.GetOrderForAdmin(orderID)
	return order, mapOrderTransportError(err)
}

func (a orderAdminQueryAdapter) UpdateOrderStatus(orderID uint, status string) (*models.Order, error) {
	order, err := a.orders.UpdateOrderStatus(orderID, status)
	return order, mapOrderTransportError(err)
}

type orderAdminUserAdapter struct {
	users repository.UserRepository
}

func (a orderAdminUserAdapter) ListByIDs(ids []uint) ([]models.User, error) {
	return a.users.ListByIDs(ids)
}

func (a orderAdminUserAdapter) GetByID(id uint) (*models.User, error) {
	return a.users.GetByID(id)
}

type orderAdminCouponAdapter struct {
	coupons coupon.Repository
}

func (a orderAdminCouponAdapter) GetByID(id uint) (*models.Coupon, error) {
	return a.coupons.GetByID(id)
}

type orderAdminPromotionAdapter struct {
	promotions promotion.Repository
}

func (a orderAdminPromotionAdapter) GetByID(id uint) (*models.Promotion, error) {
	return a.promotions.GetByID(id)
}

type orderAdminPaymentAdapter struct {
	payments repository.PaymentRepository
}

func (a orderAdminPaymentAdapter) ListByOrderID(orderID uint) ([]models.Payment, error) {
	return a.payments.ListByOrderID(orderID)
}

type orderAdminPaymentChannelAdapter struct {
	channels repository.PaymentChannelRepository
}

func (a orderAdminPaymentChannelAdapter) ListByIDs(ids []uint) ([]models.PaymentChannel, error) {
	return a.channels.ListByIDs(ids)
}

type orderUserQueryAdapter struct {
	orders *service.OrderService
}

func (a orderUserQueryAdapter) ListOrdersByUserForTenant(tenant reseller.TenantContext, filter ordertransport.UserOrderListFilter) ([]models.Order, int64, error) {
	return a.orders.ListOrdersByUserForTenant(tenant, repository.OrderListFilter{
		Page:     filter.Page,
		PageSize: filter.PageSize,
		UserID:   filter.UserID,
		Status:   filter.Status,
		OrderNo:  filter.OrderNo,
	})
}

func (a orderUserQueryAdapter) StatsOrdersByUserForTenant(tenant reseller.TenantContext, filter ordertransport.UserOrderListFilter) (map[string]int64, error) {
	return a.orders.StatsOrdersByUserForTenant(tenant, repository.OrderListFilter{
		UserID:  filter.UserID,
		OrderNo: filter.OrderNo,
	})
}

func (a orderUserQueryAdapter) GetOrderByUserOrderNoForTenant(tenant reseller.TenantContext, orderNo string, userID uint) (*models.Order, error) {
	order, err := a.orders.GetOrderByUserOrderNoForTenant(tenant, orderNo, userID)
	return order, mapOrderTransportError(err)
}

func (a orderUserQueryAdapter) GetAnyOrderByUserOrderNoForTenant(tenant reseller.TenantContext, orderNo string, userID uint) (*models.Order, error) {
	order, err := a.orders.GetAnyOrderByUserOrderNoForTenant(tenant, orderNo, userID)
	return order, mapOrderTransportError(err)
}

func (a orderUserQueryAdapter) CancelOrder(orderID uint, userID uint) (*models.Order, error) {
	order, err := a.orders.CancelOrder(orderID, userID)
	return order, mapOrderTransportError(err)
}

type orderUserPaymentChannelAdapter struct {
	payments *service.PaymentService
}

func (a orderUserPaymentChannelAdapter) GetAllowedChannelIDsForOrder(items []models.OrderItem) []uint {
	if a.payments == nil {
		return nil
	}
	return a.payments.GetAllowedChannelIDsForOrder(items)
}

func (a orderUserPaymentChannelAdapter) GetAvailableChannels(filter ordertransport.AvailablePaymentChannelFilter) ([]map[string]interface{}, error) {
	if a.payments == nil {
		return nil, nil
	}
	return a.payments.GetAvailableChannels(service.AvailablePaymentChannelFilter{
		TargetAmount: filter.TargetAmount,
		User:         filter.User,
		PaymentType:  filter.PaymentType,
	})
}

type orderUserLookupAdapter struct {
	users repository.UserRepository
}

func (a orderUserLookupAdapter) GetByID(id uint) (*models.User, error) {
	return a.users.GetByID(id)
}

type orderUserRefundRecordAdapter struct {
	records repository.OrderRefundRecordRepository
}

func (a orderUserRefundRecordAdapter) ListByOrderIDs(orderIDs []uint) ([]models.OrderRefundRecord, error) {
	return a.records.ListByOrderIDs(orderIDs)
}

type orderGuestQueryAdapter struct {
	orders *service.OrderService
}

func (a orderGuestQueryAdapter) ListOrdersByGuestForTenant(tenant reseller.TenantContext, email, password string, page, pageSize int) ([]models.Order, int64, error) {
	return a.orders.ListOrdersByGuestForTenant(tenant, email, password, page, pageSize)
}

func (a orderGuestQueryAdapter) GetOrderByGuestOrderNoForTenant(tenant reseller.TenantContext, orderNo, email, password string) (*models.Order, error) {
	order, err := a.orders.GetOrderByGuestOrderNoForTenant(tenant, orderNo, email, password)
	return order, mapOrderTransportError(err)
}

func (a orderGuestQueryAdapter) GetAnyOrderByGuestOrderNoForTenant(tenant reseller.TenantContext, orderNo, email, password string) (*models.Order, error) {
	order, err := a.orders.GetAnyOrderByGuestOrderNoForTenant(tenant, orderNo, email, password)
	return order, mapOrderTransportError(err)
}

type orderAdminRefundAdapter struct {
	refunds *service.OrderRefundService
}

func (a orderAdminRefundAdapter) ListAdminRefundItems(query ordertransport.AdminRefundListQuery) ([]ordertransport.AdminRefundItem, int64, error) {
	items, total, err := a.refunds.ListAdminRefundItems(service.AdminOrderRefundListQuery{
		Page:           query.Page,
		PageSize:       query.PageSize,
		UserID:         query.UserID,
		UserKeyword:    query.UserKeyword,
		OrderNo:        query.OrderNo,
		GuestEmail:     query.GuestEmail,
		ProductKeyword: query.ProductKeyword,
		ProductName:    query.ProductName,
		CreatedFrom:    query.CreatedFrom,
		CreatedTo:      query.CreatedTo,
	})
	if err != nil {
		return nil, 0, mapOrderTransportError(err)
	}
	return mapAdminRefundItems(items), total, nil
}

func (a orderAdminRefundAdapter) GetAdminRefundItem(id uint) (*ordertransport.AdminRefundItem, error) {
	item, err := a.refunds.GetAdminRefundItem(id)
	if err != nil {
		return nil, mapOrderTransportError(err)
	}
	mapped := mapAdminRefundItem(*item)
	return &mapped, nil
}

func (a orderAdminRefundAdapter) ParseRefundAmount(raw string) (models.Money, error) {
	amount, err := a.refunds.ParseRefundAmount(raw)
	return amount, mapOrderTransportError(err)
}

func (a orderAdminRefundAdapter) AdminManualRefund(input ordertransport.AdminManualRefundInput) (*models.Order, *models.OrderRefundRecord, error) {
	order, record, err := a.refunds.AdminManualRefund(service.AdminManualRefundInput{
		OrderID: input.OrderID,
		Amount:  input.Amount,
		Remark:  input.Remark,
	})
	return order, record, mapOrderTransportError(err)
}

type orderAdminWalletRefundAdapter struct {
	wallets *service.WalletService
}

func (a orderAdminWalletRefundAdapter) AdminRefundToWallet(input ordertransport.AdminRefundToWalletInput) (*models.Order, *models.WalletTransaction, *models.OrderRefundRecord, error) {
	order, txn, record, err := a.wallets.AdminRefundToWallet(service.AdminRefundToWalletInput{
		OrderID: input.OrderID,
		Amount:  input.Amount,
		Remark:  input.Remark,
	})
	return order, txn, record, mapOrderTransportError(err)
}

type orderAdminOrderLookupAdapter struct {
	orders repository.OrderRepository
}

func (a orderAdminOrderLookupAdapter) GetByID(id uint) (*models.Order, error) {
	return a.orders.GetByID(id)
}

type orderAdminStatusEmailAdapter struct {
	queue *queue.Client
}

func (a orderAdminStatusEmailAdapter) EnqueueOrderStatusEmail(orderID uint, status string, refundRecordID uint) error {
	if a.queue == nil {
		return nil
	}
	return a.queue.EnqueueOrderStatusEmail(queue.OrderStatusEmailPayload{
		OrderID:        orderID,
		Status:         status,
		RefundRecordID: refundRecordID,
	})
}

func mapAdminRefundItems(items []service.AdminOrderRefundItem) []ordertransport.AdminRefundItem {
	out := make([]ordertransport.AdminRefundItem, 0, len(items))
	for _, item := range items {
		out = append(out, mapAdminRefundItem(item))
	}
	return out
}

func mapAdminRefundItem(item service.AdminOrderRefundItem) ordertransport.AdminRefundItem {
	return ordertransport.AdminRefundItem{
		OrderRefundRecord: item.OrderRefundRecord,
		OrderNo:           item.OrderNo,
		GuestLocale:       item.GuestLocale,
		Items:             item.Items,
		UserEmail:         item.UserEmail,
		UserDisplayName:   item.UserDisplayName,
		RefundTypeLabel:   item.RefundTypeLabel,
	}
}

type orderPreviewAdapter struct {
	orders *service.OrderService
}

type orderCreateAdapter struct {
	orders *service.OrderService
}

func (a orderCreateAdapter) CreateOrder(input ordertransport.CreateOrderInput) (*models.Order, error) {
	order, err := a.orders.CreateOrder(service.CreateOrderInput{
		UserID:              input.UserID,
		Tenant:              input.Tenant,
		Items:               mapServiceOrderItems(input.Items),
		CouponCode:          input.CouponCode,
		AffiliateCode:       input.AffiliateCode,
		AffiliateVisitorKey: input.AffiliateVisitorKey,
		ClientIP:            input.ClientIP,
		ManualFormData:      input.ManualFormData,
	})
	return order, mapOrderTransportError(err)
}

func (a orderCreateAdapter) CreateGuestOrder(input ordertransport.CreateGuestOrderInput) (*models.Order, error) {
	order, err := a.orders.CreateGuestOrder(service.CreateGuestOrderInput{
		Email:               input.Email,
		OrderPassword:       input.OrderPassword,
		Locale:              input.Locale,
		Tenant:              input.Tenant,
		Items:               mapServiceOrderItems(input.Items),
		CouponCode:          input.CouponCode,
		AffiliateCode:       input.AffiliateCode,
		AffiliateVisitorKey: input.AffiliateVisitorKey,
		ClientIP:            input.ClientIP,
		ManualFormData:      input.ManualFormData,
	})
	return order, mapOrderTransportError(err)
}

type orderGuestCreateCaptchaAdapter struct {
	captcha *captcha.Service
}

func (a orderGuestCreateCaptchaAdapter) VerifyGuestCreateOrder(payload shared.CaptchaPayloadRequest, clientIP string) error {
	if a.captcha == nil {
		return nil
	}
	return mapOrderTransportError(a.captcha.Verify(constants.CaptchaSceneGuestCreateOrder, payload.ToCaptchaPayload(), clientIP))
}

type orderPaymentCreatorAdapter struct {
	payments *service.PaymentService
}

func (a orderPaymentCreatorAdapter) CreatePayment(input ordertransport.CreatePaymentInput) (*ordertransport.CreatePaymentResult, error) {
	if a.payments == nil {
		return nil, service.ErrPaymentInvalid
	}
	result, err := a.payments.CreatePayment(service.CreatePaymentInput{
		OrderID:       input.OrderID,
		ChannelID:     input.ChannelID,
		UseBalance:    input.UseBalance,
		ClientIP:      input.ClientIP,
		Context:       input.Context,
		RequestScheme: input.RequestScheme,
	})
	if err != nil {
		// create-and-pay returns payment_error as plain string; keep original error text.
		return nil, err
	}
	if result == nil {
		return nil, nil
	}
	return &ordertransport.CreatePaymentResult{
		Payment:          result.Payment,
		OrderPaid:        result.OrderPaid,
		WalletPaidAmount: result.WalletPaidAmount,
		OnlinePayAmount:  result.OnlinePayAmount,
	}, nil
}

func (a orderPreviewAdapter) PreviewOrder(input ordertransport.CreateOrderInput) (*ordertransport.OrderPreview, error) {
	preview, err := a.orders.PreviewOrder(service.CreateOrderInput{
		UserID:              input.UserID,
		Tenant:              input.Tenant,
		Items:               mapServiceOrderItems(input.Items),
		CouponCode:          input.CouponCode,
		AffiliateCode:       input.AffiliateCode,
		AffiliateVisitorKey: input.AffiliateVisitorKey,
		ClientIP:            input.ClientIP,
		ManualFormData:      input.ManualFormData,
	})
	if err != nil {
		return nil, mapOrderTransportError(err)
	}
	return mapOrderPreview(preview), nil
}

func (a orderPreviewAdapter) PreviewGuestOrder(input ordertransport.CreateGuestOrderInput) (*ordertransport.OrderPreview, error) {
	preview, err := a.orders.PreviewGuestOrder(service.CreateGuestOrderInput{
		Email:               input.Email,
		OrderPassword:       input.OrderPassword,
		Locale:              input.Locale,
		Tenant:              input.Tenant,
		Items:               mapServiceOrderItems(input.Items),
		CouponCode:          input.CouponCode,
		AffiliateCode:       input.AffiliateCode,
		AffiliateVisitorKey: input.AffiliateVisitorKey,
		ClientIP:            input.ClientIP,
		ManualFormData:      input.ManualFormData,
	})
	if err != nil {
		return nil, mapOrderTransportError(err)
	}
	return mapOrderPreview(preview), nil
}

func mapServiceOrderItems(items []ordertransport.CreateOrderItem) []service.CreateOrderItem {
	out := make([]service.CreateOrderItem, 0, len(items))
	for _, item := range items {
		out = append(out, service.CreateOrderItem{
			ProductID:       item.ProductID,
			SKUID:           item.SKUID,
			Quantity:        item.Quantity,
			FulfillmentType: item.FulfillmentType,
		})
	}
	return out
}

func mapOrderPreview(preview *service.OrderPreview) *ordertransport.OrderPreview {
	if preview == nil {
		return nil
	}
	items := make([]ordertransport.OrderPreviewItem, 0, len(preview.Items))
	for _, item := range preview.Items {
		items = append(items, ordertransport.OrderPreviewItem{
			ProductID:          item.ProductID,
			SKUID:              item.SKUID,
			TitleJSON:          item.TitleJSON,
			SKUSnapshotJSON:    item.SKUSnapshotJSON,
			Tags:               item.Tags,
			OriginalUnitPrice:  item.OriginalUnitPrice,
			UnitPrice:          item.UnitPrice,
			Quantity:           item.Quantity,
			OriginalTotalPrice: item.OriginalTotalPrice,
			TotalPrice:         item.TotalPrice,
			MemberDiscount:     item.MemberDiscount,
			CouponDiscount:     item.CouponDiscount,
			PromotionDiscount:  item.PromotionDiscount,
			WholesaleDiscount:  item.WholesaleDiscount,
			FulfillmentType:    item.FulfillmentType,
		})
	}
	return &ordertransport.OrderPreview{
		Currency:                preview.Currency,
		OriginalAmount:          preview.OriginalAmount,
		MemberDiscountAmount:    preview.MemberDiscountAmount,
		DiscountAmount:          preview.DiscountAmount,
		PromotionDiscountAmount: preview.PromotionDiscountAmount,
		WholesaleDiscountAmount: preview.WholesaleDiscountAmount,
		TotalAmount:             preview.TotalAmount,
		Items:                   items,
	}
}

func mapOrderTransportError(err error) error {
	if err == nil {
		return nil
	}
	if retryAfter := orderrisk.GetRetryAfter(err); retryAfter > 0 {
		return ordertransport.WrapRiskRateLimited(retryAfter, err)
	}
	for _, mapping := range []struct {
		source error
		target error
	}{
		{service.ErrOrderNotFound, ordertransport.ErrOrderNotFound},
		{service.ErrOrderStatusInvalid, ordertransport.ErrOrderStatusInvalid},
		{service.ErrOrderFetchFailed, ordertransport.ErrOrderFetchFailed},
		{service.ErrGuestOrderNotFound, ordertransport.ErrGuestOrderNotFound},
		{service.ErrOrderCancelNotAllowed, ordertransport.ErrOrderCancelNotAllowed},
		{service.ErrOrderRefundExpired, ordertransport.ErrOrderRefundExpired},
		{service.ErrWalletInvalidAmount, ordertransport.ErrWalletInvalidAmount},
		{service.ErrWalletRefundExceeded, ordertransport.ErrWalletRefundExceeded},
		{service.ErrWalletNotSupportedForGuest, ordertransport.ErrWalletNotSupportedForGuest},
		{service.ErrProductSKURequired, ordertransport.ErrProductSKURequired},
		{service.ErrInvalidOrderAmount, ordertransport.ErrInvalidOrderAmount},
		{service.ErrGuestEmailRequired, ordertransport.ErrGuestEmailRequired},
		{service.ErrGuestPasswordRequired, ordertransport.ErrGuestPasswordRequired},
		{service.ErrInvalidEmail, ordertransport.ErrInvalidEmail},
		{service.ErrProductPurchaseNotAllowed, ordertransport.ErrProductPurchaseNotAllowed},
		{service.ErrGuestCouponNotAllowed, ordertransport.ErrGuestCouponNotAllowed},
		{service.ErrManualStockInsufficient, ordertransport.ErrManualStockInsufficient},
		{service.ErrOrderCurrencyMismatch, ordertransport.ErrOrderCurrencyMismatch},
		{service.ErrProductNotAvailable, ordertransport.ErrProductNotAvailable},
		{service.ErrResellerCouponNotAllowed, ordertransport.ErrResellerCouponNotAllowed},
		{service.ErrQueueUnavailable, ordertransport.ErrQueueUnavailable},
		{orderrisk.ErrRiskIPBlacklisted, ordertransport.ErrRiskIPBlacklisted},
		{orderrisk.ErrRiskEmailBlacklisted, ordertransport.ErrRiskEmailBlacklisted},
		{orderrisk.ErrRiskTooManyPendingOrders, ordertransport.ErrRiskTooManyPendingOrders},
		{orderrisk.ErrRiskOrderRateLimited, ordertransport.ErrRiskOrderRateLimited},
	} {
		if errors.Is(err, mapping.source) {
			return fmt.Errorf("%w: %v", mapping.target, err)
		}
	}
	return err
}
