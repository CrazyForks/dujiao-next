package gormstore

import (
	"errors"
	"strings"

	"github.com/dujiao-next/internal/models"
	catalogproduct "github.com/dujiao-next/internal/modules/catalog/product"
	"github.com/dujiao-next/internal/persistence/gormutil"

	"gorm.io/gorm"
)

// SKUStore 是 Catalog Product SKU 端口的 GORM 实现。
type SKUStore struct {
	db *gorm.DB
}

var _ catalogproduct.SKURepository = (*SKUStore)(nil)

func NewSKUStore(db *gorm.DB) *SKUStore {
	return &SKUStore{db: db}
}

// BindTx 将 Store 绑定到调用方事务，并仅暴露 SKU 端口。
func (r *SKUStore) BindTx(tx *gorm.DB) catalogproduct.SKURepository {
	if tx == nil {
		return r
	}
	return NewSKUStore(tx)
}

// ListByProduct 根据商品获取 SKU 列表
func (r *SKUStore) ListByProduct(productID uint, onlyActive bool) ([]models.ProductSKU, error) {
	if productID == 0 {
		return nil, errors.New("invalid product id")
	}
	query := r.db.Model(&models.ProductSKU{}).Where("product_id = ?", productID)
	if onlyActive {
		query = query.Where("is_active = ?", true)
	}
	var items []models.ProductSKU
	if err := query.Order("sort_order DESC, id ASC").Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

// GetByID 根据 ID 获取 SKU
func (r *SKUStore) GetByID(id uint) (*models.ProductSKU, error) {
	if id == 0 {
		return nil, errors.New("invalid sku id")
	}
	var item models.ProductSKU
	if err := r.db.First(&item, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &item, nil
}

// GetByProductAndCode 按商品和编码获取 SKU
func (r *SKUStore) GetByProductAndCode(productID uint, skuCode string) (*models.ProductSKU, error) {
	if productID == 0 {
		return nil, errors.New("invalid product id")
	}
	code := strings.TrimSpace(skuCode)
	if code == "" {
		return nil, errors.New("invalid sku code")
	}

	var item models.ProductSKU
	if err := r.db.Where("product_id = ? AND sku_code = ?", productID, code).First(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &item, nil
}

// ListByIDs 批量获取 SKU
func (r *SKUStore) ListByIDs(ids []uint) ([]models.ProductSKU, error) {
	if len(ids) == 0 {
		return []models.ProductSKU{}, nil
	}
	var items []models.ProductSKU
	if err := r.db.Where("id IN ?", ids).Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

// Create 创建 SKU
func (r *SKUStore) Create(item *models.ProductSKU) error {
	if item == nil {
		return errors.New("sku is nil")
	}
	return r.db.Create(item).Error
}

// CreateBatch 批量创建 SKU
func (r *SKUStore) CreateBatch(items []models.ProductSKU) error {
	if len(items) == 0 {
		return nil
	}
	return r.db.Create(&items).Error
}

// Update 更新 SKU
func (r *SKUStore) Update(item *models.ProductSKU) error {
	if item == nil {
		return errors.New("sku is nil")
	}
	return r.db.Save(item).Error
}

// Delete 硬删除单个 SKU（绕过软删除，避免唯一索引冲突）
func (r *SKUStore) Delete(id uint) error {
	if id == 0 {
		return errors.New("invalid sku id")
	}
	return r.db.Unscoped().Delete(&models.ProductSKU{}, id).Error
}

// PurgeSoftDeletedByProductAndCode 清理指定商品下同 sku_code 的软删除残留记录
func (r *SKUStore) PurgeSoftDeletedByProductAndCode(productID uint, skuCode string) error {
	return r.db.Unscoped().
		Where("product_id = ? AND sku_code = ? AND deleted_at IS NOT NULL", productID, skuCode).
		Delete(&models.ProductSKU{}).Error
}

// DeleteByProduct 删除指定商品下的 SKU
func (r *SKUStore) DeleteByProduct(productID uint) error {
	if productID == 0 {
		return errors.New("invalid product id")
	}
	return r.db.Where("product_id = ?", productID).Delete(&models.ProductSKU{}).Error
}

// ReserveManualStock 预占手动库存
func (r *SKUStore) ReserveManualStock(skuID uint, quantity int) (int64, error) {
	return gormutil.ReserveManualStock(r.db, &models.ProductSKU{}, skuID, quantity)
}

// ReleaseManualStock 释放手动库存占用
func (r *SKUStore) ReleaseManualStock(skuID uint, quantity int) (int64, error) {
	return gormutil.ReleaseManualStock(r.db, &models.ProductSKU{}, skuID, quantity)
}

// ConsumeManualStock 消耗手动库存（支付成功后占用转已售）
func (r *SKUStore) ConsumeManualStock(skuID uint, quantity int) (int64, error) {
	return gormutil.ConsumeManualStock(r.db, &models.ProductSKU{}, skuID, quantity)
}
