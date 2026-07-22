package gormstore

import (
	"errors"

	"github.com/dujiao-next/internal/models"
	catalogmapping "github.com/dujiao-next/internal/modules/catalog/mapping"

	"gorm.io/gorm"
)

// SKUMappingStore 是 Catalog 上游 SKU 映射端口的 GORM 实现。
type SKUMappingStore struct {
	db *gorm.DB
}

var _ catalogmapping.SKUMappingRepository = (*SKUMappingStore)(nil)

func NewSKUMappingStore(db *gorm.DB) *SKUMappingStore {
	return &SKUMappingStore{db: db}
}

// WithTx 绑定事务
func (r *SKUMappingStore) WithTx(tx *gorm.DB) *SKUMappingStore {
	if tx == nil {
		return r
	}
	return NewSKUMappingStore(tx)
}

// BindTx 将事务内 store 暴露为导入用例所需的窄写入端口。
func (r *SKUMappingStore) BindTx(tx *gorm.DB) catalogmapping.SKUMappingRepository {
	return r.WithTx(tx)
}

func (r *SKUMappingStore) GetByLocalSKUID(skuID uint) (*models.SKUMapping, error) {
	var m models.SKUMapping
	if err := r.db.Where("local_sku_id = ?", skuID).First(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &m, nil
}

func (r *SKUMappingStore) ListByProductMapping(productMappingID uint) ([]models.SKUMapping, error) {
	var mappings []models.SKUMapping
	if err := r.db.Where("product_mapping_id = ?", productMappingID).Find(&mappings).Error; err != nil {
		return nil, err
	}
	return mappings, nil
}

func (r *SKUMappingStore) ListByProductMappingIDs(productMappingIDs []uint) ([]models.SKUMapping, error) {
	if len(productMappingIDs) == 0 {
		return nil, nil
	}
	var mappings []models.SKUMapping
	if err := r.db.Where("product_mapping_id IN ?", productMappingIDs).Find(&mappings).Error; err != nil {
		return nil, err
	}
	return mappings, nil
}

func (r *SKUMappingStore) Create(mapping *models.SKUMapping) error {
	return r.db.Create(mapping).Error
}

func (r *SKUMappingStore) Update(mapping *models.SKUMapping) error {
	return r.db.Save(mapping).Error
}

func (r *SKUMappingStore) DeleteByProductMapping(productMappingID uint) error {
	return r.db.Where("product_mapping_id = ?", productMappingID).Delete(&models.SKUMapping{}).Error
}
