package product

import (
	"time"

	"github.com/dujiao-next/internal/models"
)

// ListFilter 描述 Catalog 商品列表的持久化筛选条件。
type ListFilter struct {
	Page               int
	PageSize           int
	CategoryID         string
	CategoryIDs        []uint
	ExcludeProductIDs  []uint
	Search             string
	FulfillmentType    string
	StockStatus        string
	HasWholesalePrices *bool
	LowStockThreshold  int
	OnlyActive         bool
	WithCategory       bool
	UpdatedAfter       *time.Time
}

// Repository 定义商品领域需要的持久化能力；事务适配保留在 GORM Store。
type Repository interface {
	List(filter ListFilter) ([]models.Product, int64, error)
	GetBySlug(slug string, onlyActive bool) (*models.Product, error)
	GetByID(id string) (*models.Product, error)
	GetAdminByID(id string) (*models.Product, error)
	ListByIDs(ids []uint) ([]models.Product, error)
	Create(item *models.Product) error
	Update(item *models.Product) error
	Delete(id string) error
	CountBySlug(slug string, excludeID *string) (int64, error)
	ReserveManualStock(productID uint, quantity int) (int64, error)
	ReleaseManualStock(productID uint, quantity int) (int64, error)
	ConsumeManualStock(productID uint, quantity int) (int64, error)
	QuickUpdate(id string, fields map[string]interface{}) error
}

// SKURepository 定义商品 SKU 领域需要的持久化能力。
type SKURepository interface {
	ListByProduct(productID uint, onlyActive bool) ([]models.ProductSKU, error)
	GetByID(id uint) (*models.ProductSKU, error)
	GetByProductAndCode(productID uint, skuCode string) (*models.ProductSKU, error)
	ListByIDs(ids []uint) ([]models.ProductSKU, error)
	Create(item *models.ProductSKU) error
	CreateBatch(items []models.ProductSKU) error
	Update(item *models.ProductSKU) error
	Delete(id uint) error
	DeleteByProduct(productID uint) error
	PurgeSoftDeletedByProductAndCode(productID uint, skuCode string) error
	ReserveManualStock(skuID uint, quantity int) (int64, error)
	ReleaseManualStock(skuID uint, quantity int) (int64, error)
	ConsumeManualStock(skuID uint, quantity int) (int64, error)
}
