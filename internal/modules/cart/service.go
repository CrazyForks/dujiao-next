package cart

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/dujiao-next/internal/constants"
	"github.com/dujiao-next/internal/models"
	productdomain "github.com/dujiao-next/internal/modules/catalog/product/domain"
	promotionapp "github.com/dujiao-next/internal/modules/promotion/application"
	promotioncontract "github.com/dujiao-next/internal/modules/promotion/contract"
	"github.com/dujiao-next/internal/shared/money"
)

var (
	ErrInvalidItem             = errors.New("invalid cart item")
	ErrProductUnavailable      = errors.New("product not available")
	ErrFulfillmentInvalid      = errors.New("fulfillment invalid")
	ErrSKURequired             = errors.New("product sku required")
	ErrSKUInvalid              = errors.New("product sku invalid")
	ErrManualStockInsufficient = errors.New("manual stock insufficient")
)

// ItemDetail 购物车项详情（用于响应）。
type ItemDetail struct {
	ProductID       uint                      `json:"product_id"`
	SKUID           uint                      `json:"sku_id"`
	Quantity        int                       `json:"quantity"`
	FulfillmentType string                    `json:"fulfillment_type"`
	UnitPrice       money.Amount              `json:"unit_price"`
	OriginalPrice   money.Amount              `json:"original_price"`
	Currency        string                    `json:"currency"`
	Product         *productdomain.Product    `json:"product"`
	SKU             *productdomain.ProductSKU `json:"sku"`
}

// UpsertItemInput 购物车更新输入。
type UpsertItemInput struct {
	UserID          uint
	ProductID       uint
	SKUID           uint
	Quantity        int
	FulfillmentType string
}

// Repository 是购物车服务所需的最小持久化端口。
type Repository interface {
	ListByUser(userID uint) ([]models.CartItem, error)
	Upsert(item *models.CartItem) error
	DeleteByUserProductSKU(userID, productID, skuID uint) error
}

// ProductReader 是购物车服务所需的商品读取端口。
type ProductReader interface {
	GetByID(id string) (*productdomain.Product, error)
}

// SKUReader 是购物车服务所需的 SKU 读取端口。
type SKUReader interface {
	GetByID(id uint) (*productdomain.ProductSKU, error)
	ListByProduct(productID uint, onlyActive bool) ([]productdomain.ProductSKU, error)
}

// CurrencyReader 是购物车服务所需的站点币种端口。
type CurrencyReader interface {
	GetSiteCurrency(defaultCurrency string) (string, error)
}

// Service 购物车服务。
type Service struct {
	cartRepo       Repository
	productRepo    ProductReader
	productSKURepo SKUReader
	promotionRepo  promotioncontract.Repository
	currencyReader CurrencyReader
}

// NewService 创建购物车服务。
func NewService(cartRepo Repository, productRepo ProductReader, productSKURepo SKUReader, promotionRepo promotioncontract.Repository, currencyReader CurrencyReader) *Service {
	return &Service{
		cartRepo:       cartRepo,
		productRepo:    productRepo,
		productSKURepo: productSKURepo,
		promotionRepo:  promotionRepo,
		currencyReader: currencyReader,
	}
}

// ListByUser 获取用户购物车
func (s *Service) ListByUser(userID uint) ([]ItemDetail, error) {
	if userID == 0 {
		return nil, ErrInvalidItem
	}
	items, err := s.cartRepo.ListByUser(userID)
	if err != nil {
		return nil, err
	}
	currency := s.siteCurrency()
	details := make([]ItemDetail, 0, len(items))
	promotionService := promotionapp.NewService(s.promotionRepo)
	for _, item := range items {
		product := item.Product
		if product == nil || product.ID == 0 {
			p, err := s.productRepo.GetByID(strconv.FormatUint(uint64(item.ProductID), 10))
			if err != nil {
				return nil, err
			}
			product = p
		}
		if product == nil || !product.IsActive {
			_ = s.cartRepo.DeleteByUserProductSKU(userID, item.ProductID, item.SKUID)
			continue
		}

		sku := item.SKU
		if sku == nil || sku.ID == 0 {
			resolvedSKU, resolveErr := resolveProductSKU(s.productSKURepo, product, item.SKUID)
			if resolveErr != nil {
				_ = s.cartRepo.DeleteByUserProductSKU(userID, item.ProductID, item.SKUID)
				continue
			}
			sku = resolvedSKU
		}

		if sku == nil || !sku.IsActive {
			_ = s.cartRepo.DeleteByUserProductSKU(userID, item.ProductID, item.SKUID)
			continue
		}
		if strings.TrimSpace(product.FulfillmentType) == constants.FulfillmentTypeManual &&
			productdomain.ShouldEnforceManualSKUStock(product, sku) &&
			productdomain.ManualSKUAvailable(sku) <= 0 {
			_ = s.cartRepo.DeleteByUserProductSKU(userID, item.ProductID, item.SKUID)
			continue
		}

		priceCarrier := *product
		priceCarrier.PriceAmount = sku.PriceAmount
		unitPrice := sku.PriceAmount
		if promotionService != nil {
			_, discounted, err := promotionService.ApplyPromotion(&priceCarrier, item.Quantity)
			if err != nil {
				return nil, err
			}
			unitPrice = discounted
		}

		fulfillmentType := strings.TrimSpace(product.FulfillmentType)
		if fulfillmentType == "" {
			fulfillmentType = constants.FulfillmentTypeManual
		}

		details = append(details, ItemDetail{
			ProductID:       item.ProductID,
			SKUID:           sku.ID,
			Quantity:        item.Quantity,
			FulfillmentType: fulfillmentType,
			UnitPrice:       unitPrice,
			OriginalPrice:   sku.PriceAmount,
			Currency:        currency,
			Product:         product,
			SKU:             sku,
		})
	}
	return details, nil
}

// UpsertItem 添加或更新购物车项
func (s *Service) UpsertItem(input UpsertItemInput) error {
	if input.UserID == 0 || input.ProductID == 0 || input.Quantity <= 0 {
		return ErrInvalidItem
	}
	product, err := s.productRepo.GetByID(strconv.FormatUint(uint64(input.ProductID), 10))
	if err != nil {
		return err
	}
	if product == nil || !product.IsActive {
		return ErrProductUnavailable
	}
	if err := productdomain.ValidatePurchaseQuantity(product, input.Quantity); err != nil {
		return err
	}
	sku, err := resolveProductSKU(s.productSKURepo, product, input.SKUID)
	if err != nil {
		return err
	}

	fulfillmentType := strings.TrimSpace(product.FulfillmentType)
	if fulfillmentType == "" {
		fulfillmentType = constants.FulfillmentTypeManual
	}
	if fulfillmentType != constants.FulfillmentTypeManual && fulfillmentType != constants.FulfillmentTypeAuto {
		return ErrFulfillmentInvalid
	}
	if fulfillmentType == constants.FulfillmentTypeManual &&
		productdomain.ShouldEnforceManualSKUStock(product, sku) &&
		productdomain.ManualSKUAvailable(sku) < input.Quantity {
		return ErrManualStockInsufficient
	}

	now := time.Now()
	item := &models.CartItem{
		UserID:          input.UserID,
		ProductID:       input.ProductID,
		SKUID:           sku.ID,
		Quantity:        input.Quantity,
		FulfillmentType: fulfillmentType,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	return s.cartRepo.Upsert(item)
}

// RemoveItem 删除购物车项
func (s *Service) RemoveItem(userID, productID, skuID uint) error {
	if userID == 0 || productID == 0 {
		return ErrInvalidItem
	}
	return s.cartRepo.DeleteByUserProductSKU(userID, productID, skuID)
}

func (s *Service) siteCurrency() string {
	if s == nil || s.currencyReader == nil {
		return constants.SiteCurrencyDefault
	}
	currency, err := s.currencyReader.GetSiteCurrency(constants.SiteCurrencyDefault)
	if err != nil || strings.TrimSpace(currency) == "" {
		return constants.SiteCurrencyDefault
	}
	return currency
}

func resolveProductSKU(repo SKUReader, product *productdomain.Product, rawSKUID uint) (*productdomain.ProductSKU, error) {
	if product == nil || product.ID == 0 {
		return nil, ErrProductUnavailable
	}
	if repo == nil {
		return nil, ErrSKUInvalid
	}
	if rawSKUID > 0 {
		sku, err := repo.GetByID(rawSKUID)
		if err != nil {
			return nil, err
		}
		if sku == nil || sku.ProductID != product.ID || !sku.IsActive {
			return nil, ErrSKUInvalid
		}
		return sku, nil
	}
	activeSKUs, err := repo.ListByProduct(product.ID, true)
	if err != nil {
		return nil, err
	}
	if len(activeSKUs) == 1 {
		return &activeSKUs[0], nil
	}
	if len(activeSKUs) == 0 {
		return nil, ErrSKUInvalid
	}
	return nil, ErrSKURequired
}
