package gormstore

import (
	"errors"

	"github.com/dujiao-next/internal/models"
	catalogmapping "github.com/dujiao-next/internal/modules/catalog/mapping"
	"github.com/dujiao-next/internal/persistence/gormutil"

	"gorm.io/gorm"
)

// MappingStore 是 Catalog 上游映射端口的 GORM 实现。
type MappingStore struct {
	db *gorm.DB
}

var _ catalogmapping.MappingRepository = (*MappingStore)(nil)

func NewMappingStore(db *gorm.DB) *MappingStore {
	return &MappingStore{db: db}
}

// WithTx 绑定事务
func (r *MappingStore) WithTx(tx *gorm.DB) *MappingStore {
	if tx == nil {
		return r
	}
	return NewMappingStore(tx)
}

func (r *MappingStore) GetByID(id uint) (*models.ProductMapping, error) {
	var m models.ProductMapping
	if err := r.db.Preload("Connection").Preload("Product").First(&m, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &m, nil
}

func (r *MappingStore) GetByLocalProductID(productID uint) (*models.ProductMapping, error) {
	var m models.ProductMapping
	if err := r.db.Where("local_product_id = ?", productID).First(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &m, nil
}

func (r *MappingStore) GetByConnectionAndUpstreamID(connectionID, upstreamProductID uint) (*models.ProductMapping, error) {
	var m models.ProductMapping
	if err := r.db.Where("connection_id = ? AND upstream_product_id = ?", connectionID, upstreamProductID).First(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &m, nil
}

func (r *MappingStore) Create(mapping *models.ProductMapping) error {
	return r.db.Create(mapping).Error
}

func (r *MappingStore) Update(mapping *models.ProductMapping) error {
	return r.db.Save(mapping).Error
}

func (r *MappingStore) Delete(id uint) error {
	return r.db.Delete(&models.ProductMapping{}, id).Error
}

// DeleteByLocalProduct 删除指定本地商品的所有映射及其 SKU 映射
func (r *MappingStore) DeleteByLocalProduct(productID uint) error {
	// 先删除关联的 SKU 映射
	if err := r.db.Where("product_mapping_id IN (?)",
		r.db.Model(&models.ProductMapping{}).Select("id").Where("local_product_id = ?", productID),
	).Delete(&models.SKUMapping{}).Error; err != nil {
		return err
	}
	return r.db.Where("local_product_id = ?", productID).Delete(&models.ProductMapping{}).Error
}

func (r *MappingStore) List(filter catalogmapping.ListFilter) ([]models.ProductMapping, int64, error) {
	var mappings []models.ProductMapping
	var total int64

	q := r.db.Model(&models.ProductMapping{})
	if filter.ConnectionID > 0 {
		q = q.Where("connection_id = ?", filter.ConnectionID)
	}
	if filter.UpstreamStatus != "" {
		q = q.Where("upstream_status = ?", filter.UpstreamStatus)
	}
	if filter.ProductStatus == "active" {
		q = q.Where("product_mappings.is_active = ?", true)
	} else if filter.ProductStatus == "inactive" {
		q = q.Where("product_mappings.is_active = ?", false)
	}
	if filter.Search != "" {
		like := "%" + filter.Search + "%"
		condition, argCount := gormutil.BuildLocalizedLikeCondition(r.db, nil, []string{"title_json"})
		q = q.Where(
			"product_mappings.local_product_id IN (SELECT id FROM products WHERE deleted_at IS NULL AND ("+condition+"))",
			gormutil.RepeatLikeArgs(like, argCount)...,
		)
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	q = q.Preload("Connection").Preload("Product").Preload("Product.SKUs").Order("created_at DESC")
	if filter.Page > 0 && filter.PageSize > 0 {
		q = q.Offset((filter.Page - 1) * filter.PageSize).Limit(filter.PageSize)
	}

	if err := q.Find(&mappings).Error; err != nil {
		return nil, 0, err
	}

	return mappings, total, nil
}

func (r *MappingStore) ListActiveByConnection(connectionID uint) ([]models.ProductMapping, error) {
	var mappings []models.ProductMapping
	if err := r.db.Where("connection_id = ? AND is_active = ?", connectionID, true).Find(&mappings).Error; err != nil {
		return nil, err
	}
	return mappings, nil
}

func (r *MappingStore) ListAllActive() ([]models.ProductMapping, error) {
	var mappings []models.ProductMapping
	if err := r.db.Where("is_active = ?", true).Preload("Connection").Find(&mappings).Error; err != nil {
		return nil, err
	}
	return mappings, nil
}

func (r *MappingStore) ListByLocalProductIDs(productIDs []uint) ([]models.ProductMapping, error) {
	if len(productIDs) == 0 {
		return nil, nil
	}
	var mappings []models.ProductMapping
	if err := r.db.Where("local_product_id IN ?", productIDs).Find(&mappings).Error; err != nil {
		return nil, err
	}
	return mappings, nil
}

func (r *MappingStore) ListUpstreamIDsByConnection(connectionID uint) ([]uint, error) {
	var ids []uint
	if err := r.db.Model(&models.ProductMapping{}).
		Where("connection_id = ?", connectionID).
		Pluck("upstream_product_id", &ids).Error; err != nil {
		return nil, err
	}
	return ids, nil
}
