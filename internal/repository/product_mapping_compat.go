package repository

import (
	catalogmapping "github.com/dujiao-next/internal/modules/catalog/mapping"
	mappinggormstore "github.com/dujiao-next/internal/modules/catalog/mapping/store/gormstore"

	"gorm.io/gorm"
)

// ProductMapping 持久化已迁入 Catalog 上游映射子域。
// 以下接口只额外暴露事务绑定，供尚未迁移的 Order/Product 删除链路使用。
type ProductMappingRepository interface {
	catalogmapping.MappingRepository
	WithTx(tx *gorm.DB) ProductMappingRepository
}

type SKUMappingRepository interface {
	catalogmapping.SKUMappingRepository
	WithTx(tx *gorm.DB) SKUMappingRepository
}

type GormProductMappingRepository struct {
	*mappinggormstore.MappingStore
}

type GormSKUMappingRepository struct {
	*mappinggormstore.SKUMappingStore
}

func NewProductMappingRepository(db *gorm.DB) *GormProductMappingRepository {
	return AdaptProductMappingStore(mappinggormstore.NewMappingStore(db))
}

func NewSKUMappingRepository(db *gorm.DB) *GormSKUMappingRepository {
	return AdaptSKUMappingStore(mappinggormstore.NewSKUMappingStore(db))
}

func AdaptProductMappingStore(store *mappinggormstore.MappingStore) *GormProductMappingRepository {
	return &GormProductMappingRepository{MappingStore: store}
}

func AdaptSKUMappingStore(store *mappinggormstore.SKUMappingStore) *GormSKUMappingRepository {
	return &GormSKUMappingRepository{SKUMappingStore: store}
}

func (r *GormProductMappingRepository) WithTx(tx *gorm.DB) ProductMappingRepository {
	if tx == nil {
		return r
	}
	return AdaptProductMappingStore(r.MappingStore.WithTx(tx))
}

func (r *GormSKUMappingRepository) WithTx(tx *gorm.DB) SKUMappingRepository {
	if tx == nil {
		return r
	}
	return AdaptSKUMappingStore(r.SKUMappingStore.WithTx(tx))
}
