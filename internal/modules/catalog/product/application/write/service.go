package productwrite

import (
	"encoding/json"
	"errors"

	"github.com/dujiao-next/internal/models"
	productdomain "github.com/dujiao-next/internal/modules/catalog/product/domain"

	"github.com/shopspring/decimal"
)

// ProductRepository 是商品创建和更新所需的最小持久化端口。
type ProductRepository interface {
	GetByID(id string) (*productdomain.Product, error)
	Create(item *productdomain.Product) error
	Update(item *productdomain.Product) error
	CountBySlug(slug string, excludeID *string) (int64, error)
	QuickUpdate(id string, fields map[string]interface{}) error
}

// SKURepository 是商品写入用例所需的 SKU 持久化端口。
type SKURepository interface {
	ListByProduct(productID uint, onlyActive bool) ([]productdomain.ProductSKU, error)
	Create(item *productdomain.ProductSKU) error
	Update(item *productdomain.ProductSKU) error
	Delete(id uint) error
	PurgeSoftDeletedByProductAndCode(productID uint, skuCode string) error
}

// CategoryRepository 是商品分类归属校验所需的最小端口。
type CategoryRepository interface {
	productdomain.CategoryAssignmentRepository
}

// PaymentChannelRepository 是商品允许支付渠道过滤所需的最小端口。
type PaymentChannelRepository interface {
	ListByIDs(ids []uint) ([]models.PaymentChannel, error)
}

// CardSecretStockRepository 是修改自动发货 SKU 前所需的库存保护端口。
type CardSecretStockRepository interface {
	CountByProduct(productID, skuID uint) (int64, int64, int64, error)
}

// TransactionRepositories 是一次商品写事务内绑定的仓储集合。
type TransactionRepositories struct {
	Products    ProductRepository
	SKUs        SKURepository
	CardSecrets CardSecretStockRepository
}

// UnitOfWork 隐藏具体数据库和事务对象，不向 Application 暴露 GORM。
type UnitOfWork interface {
	WithinTransaction(fn func(repositories TransactionRepositories) error) error
}

// ErrorSet 保留旧 Service 错误哨兵身份，保证 errors.Is 与 HTTP 映射兼容。
type ErrorSet struct {
	NotFound                     error
	SlugExists                   error
	ProductCategoryInvalid       error
	ProductPurchaseInvalid       error
	FulfillmentInvalid           error
	ProductPriceInvalid          error
	ManualStockInvalid           error
	ProductPurchaseLimitInvalid  error
	ProductStockDisplayInvalid   error
	ProductSKUInvalid            error
	ProductSKUHasCardSecretStock error
}

// Options 描述商品写入应用服务依赖。
type Options struct {
	Products        ProductRepository
	SKUs            SKURepository
	Categories      CategoryRepository
	PaymentChannels PaymentChannelRepository
	Transactions    UnitOfWork
	Errors          ErrorSet
}

// WriteService 编排商品创建、更新和 SKU 同步。
type WriteService struct {
	products        ProductRepository
	skus            SKURepository
	categories      CategoryRepository
	paymentChannels PaymentChannelRepository
	transactions    UnitOfWork
	errors          ErrorSet
}

// NewWriteService 创建商品写入应用服务。
func NewWriteService(options Options) *WriteService {
	return &WriteService{
		products:        options.Products,
		skus:            options.SKUs,
		categories:      options.Categories,
		paymentChannels: options.PaymentChannels,
		transactions:    options.Transactions,
		errors:          resolveErrorSet(options.Errors),
	}
}

// CreateProductInput 创建或完整更新商品的输入。
type CreateProductInput struct {
	CategoryID           uint
	Slug                 string
	SeoMetaJSON          map[string]interface{}
	TitleJSON            map[string]interface{}
	DescriptionJSON      map[string]interface{}
	ContentJSON          map[string]interface{}
	InstructionsJSON     map[string]interface{}
	ManualFormSchemaJSON map[string]interface{}
	PriceAmount          decimal.Decimal
	CostPriceAmount      decimal.Decimal
	// WholesalePrices 为可选字段：nil 表示更新时保留，非 nil 表示整体覆盖。
	WholesalePrices     *[]WholesalePriceInput
	Images              []string
	Tags                []string
	PurchaseType        string
	MinPurchaseQuantity *int
	MaxPurchaseQuantity *int
	StockDisplayMode    string
	FulfillmentType     string
	ManualStockTotal    *int
	SKUs                []ProductSKUInput
	PaymentChannelIDs   []uint
	IsAffiliateEnabled  *bool
	IsActive            *bool
	SortOrder           int
}

type WholesalePriceInput = productdomain.WholesalePriceInput

// ProductSKUInput 描述商品 SKU 的完整写入值。
type ProductSKUInput struct {
	ID               uint
	SKUCode          string
	SpecValuesJSON   map[string]interface{}
	PriceAmount      decimal.Decimal
	CostPriceAmount  decimal.Decimal
	ManualStockTotal int
	IsActive         *bool
	SortOrder        int
}

func resolveErrorSet(values ErrorSet) ErrorSet {
	return ErrorSet{
		NotFound:                    errorOrDefault(values.NotFound, "not found"),
		SlugExists:                  errorOrDefault(values.SlugExists, "slug exists"),
		ProductCategoryInvalid:      errorOrDefault(values.ProductCategoryInvalid, "product category invalid"),
		ProductPurchaseInvalid:      errorOrDefault(values.ProductPurchaseInvalid, "product purchase invalid"),
		FulfillmentInvalid:          errorOrDefault(values.FulfillmentInvalid, "fulfillment invalid"),
		ProductPriceInvalid:         errorOrDefault(values.ProductPriceInvalid, "product price invalid"),
		ManualStockInvalid:          errorOrDefault(values.ManualStockInvalid, "manual stock invalid"),
		ProductPurchaseLimitInvalid: errorOrDefault(values.ProductPurchaseLimitInvalid, "product purchase limit invalid"),
		ProductStockDisplayInvalid:  errorOrDefault(values.ProductStockDisplayInvalid, "product stock display invalid"),
		ProductSKUInvalid:           errorOrDefault(values.ProductSKUInvalid, "product sku invalid"),
		ProductSKUHasCardSecretStock: errorOrDefault(
			values.ProductSKUHasCardSecretStock,
			"product sku has card secret stock",
		),
	}
}

func errorOrDefault(value error, message string) error {
	if value != nil {
		return value
	}
	return errors.New(message)
}

func (s *WriteService) filterAvailablePaymentChannelIDs(ids []uint) ([]uint, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	uniqueIDs := make([]uint, 0, len(ids))
	seen := make(map[uint]struct{}, len(ids))
	for _, id := range ids {
		if id == 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		uniqueIDs = append(uniqueIDs, id)
	}
	if len(uniqueIDs) == 0 {
		return nil, nil
	}
	if s.paymentChannels == nil {
		return uniqueIDs, nil
	}

	channels, err := s.paymentChannels.ListByIDs(uniqueIDs)
	if err != nil {
		return nil, err
	}
	activeIDs := make(map[uint]struct{}, len(channels))
	for _, channel := range channels {
		if channel.IsActive {
			activeIDs[channel.ID] = struct{}{}
		}
	}

	filtered := make([]uint, 0, len(uniqueIDs))
	for _, id := range uniqueIDs {
		if _, ok := activeIDs[id]; ok {
			filtered = append(filtered, id)
		}
	}
	if len(filtered) == 0 {
		return nil, nil
	}
	return filtered, nil
}

func encodeChannelIDs(ids []uint) string {
	if len(ids) == 0 {
		return ""
	}
	payload, err := json.Marshal(ids)
	if err != nil {
		return ""
	}
	return string(payload)
}
