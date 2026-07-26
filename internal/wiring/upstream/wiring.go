package upstreamwiring

import (
	"errors"
	"fmt"
	"time"

	productdomain "github.com/dujiao-next/internal/modules/catalog/product/domain"

	"github.com/dujiao-next/internal/models"
	productapplication "github.com/dujiao-next/internal/modules/catalog/product/application"
	walletcontract "github.com/dujiao-next/internal/modules/wallet/contract"
	"github.com/dujiao-next/internal/provider"
	"github.com/dujiao-next/internal/service"
	"github.com/dujiao-next/internal/shared/jsonmap"
	upstreamtransport "github.com/dujiao-next/internal/transport/http/upstream"
)

// NewHandler connects the legacy application services to the upstream HTTP
// transport without leaking concrete implementations into transport.
func NewHandler(c *provider.Container) *upstreamtransport.Handler {
	return upstreamtransport.New(upstreamtransport.Dependencies{
		Categories:        c.CategoryRepo,
		Products:          productServiceAdapter{products: c.ProductReadService},
		Users:             c.UserStore,
		ProductRepository: c.ProductRepo,
		SKUs:              c.ProductSKURepo,
		ProductMappings:   c.ProductMappingRepo,
		SKUMappings:       c.SKUMappingRepo,
		MemberLevels:      c.MemberLevelService,
		Settings:          c.SettingService,
		Wallet:            c.WalletService,
		Orders:            orderServiceAdapter{orders: c.OrderService},
		Payments:          paymentServiceAdapter{payments: c.PaymentService},
		Procurements:      c.ProcurementOrderService,
		DownstreamRefs:    c.DownstreamOrderRefRepo,
		Connections:       c.SiteConnectionRepo,
		ConnectionSecrets: c.SiteConnectionService,
	})
}

type productServiceAdapter struct {
	products *productapplication.Service
}

func (a productServiceAdapter) ListForUpstreamSync(updatedAfter *time.Time, includeInactive bool, page, pageSize int) ([]productdomain.Product, int64, error) {
	return a.products.ListForUpstreamSync(updatedAfter, includeInactive, page, pageSize)
}

func (a productServiceAdapter) ApplyAutoStockCounts(products []productdomain.Product) error {
	return a.products.ApplyAutoStockCounts(products)
}

func (a productServiceAdapter) GetAdminByID(id string) (*productdomain.Product, error) {
	product, err := a.products.GetAdminByID(id)
	if errors.Is(err, service.ErrNotFound) {
		return nil, fmt.Errorf("%w: %v", upstreamtransport.ErrProductNotFound, err)
	}
	return product, err
}

type orderServiceAdapter struct {
	orders *service.OrderService
}

func (a orderServiceAdapter) CreateOrder(input upstreamtransport.CreateOrderInput) (*models.Order, error) {
	items := make([]service.CreateOrderItem, 0, len(input.Items))
	for _, item := range input.Items {
		items = append(items, service.CreateOrderItem{
			ProductID: item.ProductID, SKUID: item.SKUID, Quantity: item.Quantity, FulfillmentType: item.FulfillmentType,
		})
	}
	order, err := a.orders.CreateOrder(service.CreateOrderInput{
		UserID: input.UserID, Items: items, ClientIP: input.ClientIP,
		ManualFormData: input.ManualFormData, SkipRiskControl: input.SkipRiskControl,
	})
	return order, mapOrderError(err)
}

func (a orderServiceAdapter) GetOrderByUser(orderID, userID uint) (*models.Order, error) {
	order, err := a.orders.GetOrderByUser(orderID, userID)
	return order, mapOrderError(err)
}

func (a orderServiceAdapter) CancelOrder(orderID, userID uint) (*models.Order, error) {
	order, err := a.orders.CancelOrder(orderID, userID)
	return order, mapOrderError(err)
}

func (a orderServiceAdapter) BuildLocalRefundRecordsForOrder(order *models.Order) ([]jsonmap.JSON, error) {
	return a.orders.BuildLocalRefundRecordsForOrder(order)
}

type paymentServiceAdapter struct {
	payments *service.PaymentService
}

func (a paymentServiceAdapter) CreatePayment(input upstreamtransport.CreatePaymentInput) (*upstreamtransport.CreatePaymentResult, error) {
	result, err := a.payments.CreatePayment(service.CreatePaymentInput{
		OrderID: input.OrderID, UseBalance: input.UseBalance, ClientIP: input.ClientIP,
	})
	if err != nil || result == nil {
		return nil, err
	}
	return &upstreamtransport.CreatePaymentResult{OrderPaid: result.OrderPaid}, nil
}

func mapOrderError(err error) error {
	if err == nil {
		return nil
	}
	for _, mapping := range []struct {
		sources []error
		target  error
	}{
		{[]error{service.ErrOrderNotFound}, upstreamtransport.ErrOrderNotFound},
		{[]error{service.ErrOrderCancelNotAllowed}, upstreamtransport.ErrOrderCancelNotAllowed},
		{[]error{walletcontract.ErrInsufficientBalance}, upstreamtransport.ErrWalletInsufficient},
		{[]error{service.ErrCardSecretInsufficient, service.ErrManualStockInsufficient}, upstreamtransport.ErrStockInsufficient},
		{[]error{service.ErrProductNotAvailable, service.ErrProductNotFound}, upstreamtransport.ErrProductUnavailable},
		{[]error{service.ErrProductSKUInvalid, service.ErrProductSKURequired}, upstreamtransport.ErrSKUUnavailable},
		{[]error{service.ErrInvalidOrderItem}, upstreamtransport.ErrInvalidOrderItem},
		{[]error{service.ErrManualFormRequiredMissing, service.ErrManualFormFieldInvalid, service.ErrManualFormTypeInvalid, service.ErrManualFormOptionInvalid}, upstreamtransport.ErrManualFormInvalid},
	} {
		for _, source := range mapping.sources {
			if errors.Is(err, source) {
				return fmt.Errorf("%w: %v", mapping.target, err)
			}
		}
	}
	return err
}
