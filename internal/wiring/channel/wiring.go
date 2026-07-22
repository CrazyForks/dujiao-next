package channelwiring

import (
	"errors"
	"fmt"

	"github.com/dujiao-next/internal/models"
	externalidentitydomain "github.com/dujiao-next/internal/modules/identity/externalidentity/domain"
	"github.com/dujiao-next/internal/modules/orderrisk"
	"github.com/dujiao-next/internal/provider"
	"github.com/dujiao-next/internal/repository"
	"github.com/dujiao-next/internal/service"
	channeltransport "github.com/dujiao-next/internal/transport/http/channel"
)

// NewHandler connects the legacy application services to the channel HTTP
// transport. Conversion stays at the composition boundary so transport can
// depend on narrow contracts only.
func NewHandler(c *provider.Container) *channeltransport.Handler {
	return channeltransport.New(channeltransport.Dependencies{
		CategoryService: c.CategoryService, CategoryRepo: c.CategoryRepo,
		ProductService: c.ProductService, ProductRepo: c.ProductRepo,
		ProductMappingRepo: c.ProductMappingRepo, SKUMappingRepo: c.SKUMappingRepo,
		UserAuthService: identityAdapter{auth: c.UserAuthService}, MemberLevelService: c.MemberLevelService,
		SettingService: c.SettingService, OrderService: orderAdapter{orders: c.OrderService},
		PaymentService: paymentAdapter{payments: c.PaymentService}, PaymentRepo: c.PaymentRepo,
	})
}

type identityAdapter struct {
	auth *service.UserAuthService
}

func identityServiceInput(input channeltransport.TelegramIdentityInput) service.TelegramChannelIdentityInput {
	return service.TelegramChannelIdentityInput{
		ChannelUserID: input.ChannelUserID, Username: input.Username,
		FirstName: input.FirstName, LastName: input.LastName, AvatarURL: input.AvatarURL,
	}
}

func (a identityAdapter) ResolveTelegramChannelIdentity(input channeltransport.TelegramIdentityInput) (*models.User, *externalidentitydomain.Identity, error) {
	return a.auth.ResolveTelegramChannelIdentity(identityServiceInput(input))
}

func (a identityAdapter) ProvisionTelegramChannelIdentity(input channeltransport.TelegramIdentityInput) (*models.User, *externalidentitydomain.Identity, bool, error) {
	return a.auth.ProvisionTelegramChannelIdentity(identityServiceInput(input))
}

func (a identityAdapter) ProvisionTelegramChannelUserID(input channeltransport.TelegramIdentityInput) (uint, error) {
	user, _, _, err := a.auth.ProvisionTelegramChannelIdentity(identityServiceInput(input))
	if err != nil {
		return 0, err
	}
	if user == nil {
		return 0, service.ErrNotFound
	}
	return user.ID, nil
}

func (a identityAdapter) BindTelegramChannelByEmailCode(input channeltransport.BindTelegramIdentityInput) (*models.User, *externalidentitydomain.Identity, uint, error) {
	return a.auth.BindTelegramChannelByEmailCode(service.BindTelegramChannelByEmailCodeInput{
		Identity: identityServiceInput(input.Identity), Email: input.Email, Code: input.Code,
	})
}

type orderAdapter struct {
	orders *service.OrderService
}

func createOrderServiceInput(input channeltransport.CreateOrderInput) service.CreateOrderInput {
	items := make([]service.CreateOrderItem, 0, len(input.Items))
	for _, item := range input.Items {
		items = append(items, service.CreateOrderItem{
			ProductID: item.ProductID, SKUID: item.SKUID, Quantity: item.Quantity, FulfillmentType: item.FulfillmentType,
		})
	}
	return service.CreateOrderInput{
		UserID: input.UserID, Items: items, CouponCode: input.CouponCode,
		AffiliateCode: input.AffiliateCode, AffiliateVisitorKey: input.AffiliateVisitorKey,
		ClientIP: input.ClientIP, ManualFormData: input.ManualFormData, SkipIPRiskControl: input.SkipIPRiskControl,
	}
}

func (a orderAdapter) PreviewOrder(input channeltransport.CreateOrderInput) (*channeltransport.OrderPreview, error) {
	preview, err := a.orders.PreviewOrder(createOrderServiceInput(input))
	if err != nil {
		return nil, mapError(err)
	}
	if preview == nil {
		return nil, nil
	}
	items := make([]channeltransport.OrderPreviewItem, 0, len(preview.Items))
	for _, item := range preview.Items {
		items = append(items, channeltransport.OrderPreviewItem{
			ProductID: item.ProductID, SKUID: item.SKUID, TitleJSON: item.TitleJSON, SKUSnapshotJSON: item.SKUSnapshotJSON,
			OriginalUnitPrice: item.OriginalUnitPrice, UnitPrice: item.UnitPrice, Quantity: item.Quantity,
			OriginalTotalPrice: item.OriginalTotalPrice, TotalPrice: item.TotalPrice,
			CouponDiscount: item.CouponDiscount, PromotionDiscount: item.PromotionDiscount,
			WholesaleDiscount: item.WholesaleDiscount, FulfillmentType: item.FulfillmentType,
		})
	}
	return &channeltransport.OrderPreview{
		Currency: preview.Currency, OriginalAmount: preview.OriginalAmount, DiscountAmount: preview.DiscountAmount,
		PromotionDiscountAmount: preview.PromotionDiscountAmount, WholesaleDiscountAmount: preview.WholesaleDiscountAmount,
		TotalAmount: preview.TotalAmount, Items: items,
	}, nil
}

func (a orderAdapter) CreateOrder(input channeltransport.CreateOrderInput) (*models.Order, error) {
	order, err := a.orders.CreateOrder(createOrderServiceInput(input))
	return order, mapError(err)
}

func (a orderAdapter) GetOrderByUser(orderID, userID uint) (*models.Order, error) {
	order, err := a.orders.GetOrderByUser(orderID, userID)
	return order, mapError(err)
}

func (a orderAdapter) GetOrderByUserOrderNo(orderNo string, userID uint) (*models.Order, error) {
	order, err := a.orders.GetOrderByUserOrderNo(orderNo, userID)
	return order, mapError(err)
}

func (a orderAdapter) CancelOrder(orderID, userID uint) (*models.Order, error) {
	order, err := a.orders.CancelOrder(orderID, userID)
	return order, mapError(err)
}

func (a orderAdapter) ListOrdersByUser(filter channeltransport.OrderListFilter) ([]models.Order, int64, error) {
	return a.orders.ListOrdersByUser(repository.OrderListFilter{
		Page: filter.Page, PageSize: filter.PageSize, UserID: filter.UserID, Status: filter.Status,
	})
}

type paymentAdapter struct {
	payments *service.PaymentService
}

func (a paymentAdapter) GetWalletRechargeChannels() ([]models.PaymentChannel, error) {
	return a.payments.GetWalletRechargeChannels()
}

func (a paymentAdapter) ListChannels(filter channeltransport.PaymentChannelListFilter) ([]models.PaymentChannel, int64, error) {
	return a.payments.ListChannels(repository.PaymentChannelListFilter{
		Page: filter.Page, PageSize: filter.PageSize, ActiveOnly: filter.ActiveOnly,
	})
}

func (a paymentAdapter) GetAllowedChannelsForProducts(productIDs []uint) ([]models.PaymentChannel, error) {
	return a.payments.GetAllowedChannelsForProducts(productIDs)
}

func (a paymentAdapter) CreatePayment(input channeltransport.CreatePaymentInput) (*channeltransport.CreatePaymentResult, error) {
	result, err := a.payments.CreatePayment(service.CreatePaymentInput{
		OrderID: input.OrderID, ChannelID: input.ChannelID, UseBalance: input.UseBalance,
		ClientIP: input.ClientIP, Context: input.Context,
	})
	if err != nil {
		return nil, mapError(err)
	}
	if result == nil {
		return nil, nil
	}
	return &channeltransport.CreatePaymentResult{
		Payment: result.Payment, Channel: result.Channel, OrderPaid: result.OrderPaid,
		WalletPaidAmount: result.WalletPaidAmount, OnlinePayAmount: result.OnlinePayAmount,
	}, nil
}

func mapError(err error) error {
	if err == nil {
		return nil
	}
	for _, mapping := range []struct {
		source error
		target error
	}{
		{orderrisk.ErrRiskIPBlacklisted, channeltransport.ErrRiskIPBlacklisted},
		{orderrisk.ErrRiskEmailBlacklisted, channeltransport.ErrRiskEmailBlacklisted},
		{orderrisk.ErrRiskTooManyPendingOrders, channeltransport.ErrRiskTooManyPendingOrders},
		{orderrisk.ErrRiskOrderRateLimited, channeltransport.ErrRiskOrderRateLimited},
		{service.ErrProductSKURequired, channeltransport.ErrProductSKURequired},
		{service.ErrProductSKUInvalid, channeltransport.ErrProductSKUInvalid},
		{service.ErrInvalidOrderItem, channeltransport.ErrInvalidOrderItem},
		{service.ErrInvalidOrderAmount, channeltransport.ErrInvalidOrderAmount},
		{service.ErrProductPurchaseNotAllowed, channeltransport.ErrProductPurchaseNotAllowed},
		{service.ErrProductMaxPurchaseExceeded, channeltransport.ErrProductMaxPurchaseExceeded},
		{service.ErrProductMinPurchaseNotMet, channeltransport.ErrProductMinPurchaseNotMet},
		{service.ErrProductNotAvailable, channeltransport.ErrProductNotAvailable},
		{service.ErrManualStockInsufficient, channeltransport.ErrManualStockInsufficient},
		{service.ErrCardSecretInsufficient, channeltransport.ErrCardSecretInsufficient},
		{service.ErrOrderCurrencyMismatch, channeltransport.ErrOrderCurrencyMismatch},
		{service.ErrProductPriceInvalid, channeltransport.ErrProductPriceInvalid},
		{service.ErrManualFormSchemaInvalid, channeltransport.ErrManualFormSchemaInvalid},
		{service.ErrManualFormRequiredMissing, channeltransport.ErrManualFormRequiredMissing},
		{service.ErrManualFormFieldInvalid, channeltransport.ErrManualFormFieldInvalid},
		{service.ErrManualFormTypeInvalid, channeltransport.ErrManualFormTypeInvalid},
		{service.ErrManualFormOptionInvalid, channeltransport.ErrManualFormOptionInvalid},
		{service.ErrPaymentInvalid, channeltransport.ErrPaymentInvalid},
		{service.ErrOrderNotFound, channeltransport.ErrOrderNotFound},
		{service.ErrOrderStatusInvalid, channeltransport.ErrOrderStatusInvalid},
		{service.ErrPaymentChannelNotFound, channeltransport.ErrPaymentChannelNotFound},
		{service.ErrPaymentChannelInactive, channeltransport.ErrPaymentChannelInactive},
		{service.ErrPaymentProviderNotSupported, channeltransport.ErrPaymentProviderUnsupported},
		{service.ErrPaymentChannelConfigInvalid, channeltransport.ErrPaymentChannelConfigInvalid},
		{service.ErrPaymentGatewayRequestFailed, channeltransport.ErrPaymentGatewayRequestFailed},
		{service.ErrPaymentGatewayResponseInvalid, channeltransport.ErrPaymentGatewayResponseInvalid},
		{service.ErrPaymentCurrencyMismatch, channeltransport.ErrPaymentCurrencyMismatch},
		{service.ErrWalletOnlyPaymentRequired, channeltransport.ErrWalletOnlyPaymentRequired},
	} {
		if errors.Is(err, mapping.source) {
			mapped := fmt.Errorf("%w: %v", mapping.target, err)
			return channeltransport.WithRetryAfter(mapped, orderrisk.GetRetryAfter(err))
		}
	}
	return channeltransport.WithRetryAfter(err, orderrisk.GetRetryAfter(err))
}
