package repository

import (
	catalogproduct "github.com/dujiao-next/internal/modules/catalog/product"
	productgormstore "github.com/dujiao-next/internal/modules/catalog/product/store/gormstore"

	"gorm.io/gorm"
)

// Product 持久化已经迁入 Catalog 子域。以下接口与构造器仅承接尚未迁移的
// Order、Payment、Cart 和旧 ProductService 的事务绑定需求；只读消费方应直接依赖
// catalogproduct.Repository。
type ProductListFilter = catalogproduct.ListFilter

type ProductRepository interface {
	catalogproduct.Repository
	Transaction(fn func(tx *gorm.DB) error) error
	WithTx(tx *gorm.DB) ProductRepository
}

type ProductSKURepository interface {
	catalogproduct.SKURepository
	WithTx(tx *gorm.DB) ProductSKURepository
}

type GormProductRepository struct {
	*productgormstore.ProductStore
}

type GormProductSKURepository struct {
	*productgormstore.SKUStore
}

func NewProductRepository(db *gorm.DB) *GormProductRepository {
	return AdaptProductStore(productgormstore.NewProductStore(db))
}

func NewProductSKURepository(db *gorm.DB) *GormProductSKURepository {
	return AdaptProductSKUStore(productgormstore.NewSKUStore(db))
}

func AdaptProductStore(store *productgormstore.ProductStore) *GormProductRepository {
	return &GormProductRepository{ProductStore: store}
}

func AdaptProductSKUStore(store *productgormstore.SKUStore) *GormProductSKURepository {
	return &GormProductSKURepository{SKUStore: store}
}

func (r *GormProductRepository) WithTx(tx *gorm.DB) ProductRepository {
	if tx == nil {
		return r
	}
	return AdaptProductStore(r.ProductStore.WithTx(tx))
}

func (r *GormProductSKURepository) WithTx(tx *gorm.DB) ProductSKURepository {
	if tx == nil {
		return r
	}
	return AdaptProductSKUStore(r.SKUStore.WithTx(tx))
}
