package channelhttp

import (
	"context"
	"errors"
	"time"

	"github.com/dujiao-next/internal/models"
	externalidentitydomain "github.com/dujiao-next/internal/modules/identity/externalidentity/domain"
	"github.com/dujiao-next/internal/shared/jsonmap"

	"github.com/shopspring/decimal"
)

var (
	ErrRiskIPBlacklisted             = errors.New("risk ip blacklisted")
	ErrRiskEmailBlacklisted          = errors.New("risk email blacklisted")
	ErrRiskTooManyPendingOrders      = errors.New("too many pending orders")
	ErrRiskOrderRateLimited          = errors.New("order rate limited")
	ErrProductSKURequired            = errors.New("product sku required")
	ErrProductSKUInvalid             = errors.New("product sku invalid")
	ErrInvalidOrderItem              = errors.New("invalid order item")
	ErrInvalidOrderAmount            = errors.New("invalid order amount")
	ErrProductPurchaseNotAllowed     = errors.New("product purchase not allowed")
	ErrProductMaxPurchaseExceeded    = errors.New("product max purchase exceeded")
	ErrProductMinPurchaseNotMet      = errors.New("product min purchase not met")
	ErrProductNotAvailable           = errors.New("product not available")
	ErrManualStockInsufficient       = errors.New("manual stock insufficient")
	ErrCardSecretInsufficient        = errors.New("card secret insufficient")
	ErrOrderCurrencyMismatch         = errors.New("order currency mismatch")
	ErrProductPriceInvalid           = errors.New("product price invalid")
	ErrManualFormSchemaInvalid       = errors.New("manual form schema invalid")
	ErrManualFormRequiredMissing     = errors.New("manual form required missing")
	ErrManualFormFieldInvalid        = errors.New("manual form field invalid")
	ErrManualFormTypeInvalid         = errors.New("manual form type invalid")
	ErrManualFormOptionInvalid       = errors.New("manual form option invalid")
	ErrPaymentInvalid                = errors.New("payment invalid")
	ErrOrderNotFound                 = errors.New("order not found")
	ErrOrderStatusInvalid            = errors.New("order status invalid")
	ErrPaymentChannelNotFound        = errors.New("payment channel not found")
	ErrPaymentChannelInactive        = errors.New("payment channel inactive")
	ErrPaymentProviderUnsupported    = errors.New("payment provider unsupported")
	ErrPaymentChannelConfigInvalid   = errors.New("payment channel config invalid")
	ErrPaymentGatewayRequestFailed   = errors.New("payment gateway request failed")
	ErrPaymentGatewayResponseInvalid = errors.New("payment gateway response invalid")
	ErrPaymentCurrencyMismatch       = errors.New("payment currency mismatch")
	ErrWalletOnlyPaymentRequired     = errors.New("wallet only payment required")
)

type TelegramIdentityInput struct {
	ChannelUserID string
	Username      string
	FirstName     string
	LastName      string
	AvatarURL     string
}

type BindTelegramIdentityInput struct {
	Identity TelegramIdentityInput
	Email    string
	Code     string
}

type CreateOrderItem struct {
	ProductID       uint
	SKUID           uint
	Quantity        int
	FulfillmentType string
}

type CreateOrderInput struct {
	UserID              uint
	Items               []CreateOrderItem
	CouponCode          string
	AffiliateCode       string
	AffiliateVisitorKey string
	ClientIP            string
	ManualFormData      map[string]jsonmap.JSON
	SkipIPRiskControl   bool
}

type OrderPreview struct {
	Currency                string
	OriginalAmount          models.Money
	DiscountAmount          models.Money
	PromotionDiscountAmount models.Money
	WholesaleDiscountAmount models.Money
	TotalAmount             models.Money
	Items                   []OrderPreviewItem
}

type OrderPreviewItem struct {
	ProductID, SKUID   uint
	TitleJSON          jsonmap.JSON
	SKUSnapshotJSON    jsonmap.JSON
	OriginalUnitPrice  models.Money
	UnitPrice          models.Money
	Quantity           int
	OriginalTotalPrice models.Money
	TotalPrice         models.Money
	CouponDiscount     models.Money
	PromotionDiscount  models.Money
	WholesaleDiscount  models.Money
	FulfillmentType    string
}

type OrderListFilter struct {
	Page, PageSize int
	UserID         uint
	Status         string
}

type PaymentChannelListFilter struct {
	Page, PageSize int
	ActiveOnly     bool
}

type CreatePaymentInput struct {
	OrderID, ChannelID uint
	UseBalance         bool
	ClientIP           string
	Context            context.Context
}

type CreatePaymentResult struct {
	Payment          *models.Payment
	Channel          *models.PaymentChannel
	OrderPaid        bool
	WalletPaidAmount models.Money
	OnlinePayAmount  models.Money
}

type CategoryService interface {
	ListActive() ([]models.Category, error)
}

type CategoryRepository interface {
	CountActiveProducts(categoryID string) (int64, error)
}

type ProductService interface {
	ListPublicExact(categoryID string, page, pageSize int) ([]models.Product, int64, error)
	ListPublic(categoryID, search string, page, pageSize int) ([]models.Product, int64, error)
	ApplyAutoStockCounts(products []models.Product) error
}

type ProductRepository interface {
	GetByID(id string) (*models.Product, error)
}

type ProductMappingRepository interface {
	ListByLocalProductIDs(productIDs []uint) ([]models.ProductMapping, error)
}

type SKUMappingRepository interface {
	ListByProductMappingIDs(mappingIDs []uint) ([]models.SKUMapping, error)
}

type IdentityService interface {
	ResolveTelegramChannelIdentity(input TelegramIdentityInput) (*models.User, *externalidentitydomain.Identity, error)
	ProvisionTelegramChannelIdentity(input TelegramIdentityInput) (*models.User, *externalidentitydomain.Identity, bool, error)
	ProvisionTelegramChannelUserID(input TelegramIdentityInput) (uint, error)
	BindTelegramChannelByEmailCode(input BindTelegramIdentityInput) (*models.User, *externalidentitydomain.Identity, uint, error)
}

type MemberLevelService interface {
	ResolveMemberPrice(levelID, productID, skuID uint, basePrice decimal.Decimal) (decimal.Decimal, decimal.Decimal)
}

type Settings interface {
	GetSiteCurrency(defaultValue string) (string, error)
	GetWalletOnlyPayment() bool
}

type Orders interface {
	PreviewOrder(input CreateOrderInput) (*OrderPreview, error)
	CreateOrder(input CreateOrderInput) (*models.Order, error)
	GetOrderByUser(orderID, userID uint) (*models.Order, error)
	GetOrderByUserOrderNo(orderNo string, userID uint) (*models.Order, error)
	CancelOrder(orderID, userID uint) (*models.Order, error)
	ListOrdersByUser(filter OrderListFilter) ([]models.Order, int64, error)
}

type Payments interface {
	GetWalletRechargeChannels() ([]models.PaymentChannel, error)
	ListChannels(filter PaymentChannelListFilter) ([]models.PaymentChannel, int64, error)
	GetAllowedChannelsForProducts(productIDs []uint) ([]models.PaymentChannel, error)
	CreatePayment(input CreatePaymentInput) (*CreatePaymentResult, error)
}

type PaymentRepository interface {
	GetLatestPendingByOrder(orderID uint, now time.Time) (*models.Payment, error)
	GetByID(id uint) (*models.Payment, error)
}

type Dependencies struct {
	CategoryService    CategoryService
	CategoryRepo       CategoryRepository
	ProductService     ProductService
	ProductRepo        ProductRepository
	ProductMappingRepo ProductMappingRepository
	SKUMappingRepo     SKUMappingRepository
	UserAuthService    IdentityService
	MemberLevelService MemberLevelService
	SettingService     Settings
	OrderService       Orders
	PaymentService     Payments
	PaymentRepo        PaymentRepository
}

type Handler struct {
	Dependencies
}

func New(dependencies Dependencies) *Handler {
	if dependencies.CategoryService == nil || dependencies.CategoryRepo == nil || dependencies.ProductService == nil ||
		dependencies.ProductRepo == nil || dependencies.ProductMappingRepo == nil || dependencies.SKUMappingRepo == nil ||
		dependencies.UserAuthService == nil || dependencies.MemberLevelService == nil || dependencies.SettingService == nil ||
		dependencies.OrderService == nil || dependencies.PaymentService == nil || dependencies.PaymentRepo == nil {
		panic("channel handler: required dependency is nil")
	}
	return &Handler{Dependencies: dependencies}
}

type retryAfterError struct {
	cause   error
	seconds int64
}

func (e retryAfterError) Error() string { return e.cause.Error() }
func (e retryAfterError) Unwrap() error { return e.cause }

func WithRetryAfter(err error, seconds int64) error {
	if err == nil || seconds <= 0 {
		return err
	}
	return retryAfterError{cause: err, seconds: seconds}
}

func retryAfter(err error) int64 {
	var wrapped retryAfterError
	if errors.As(err, &wrapped) {
		return wrapped.seconds
	}
	return 0
}
