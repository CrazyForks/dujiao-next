package gormstore

import (
	"errors"

	"github.com/dujiao-next/internal/models"
	"github.com/dujiao-next/internal/modules/memberlevel"

	"gorm.io/gorm"
)

type PriceStore struct {
	db *gorm.DB
}

func NewPriceStore(db *gorm.DB) *PriceStore {
	return &PriceStore{db: db}
}

func (r *PriceStore) GetByID(id uint) (*models.MemberLevelPrice, error) {
	var price models.MemberLevelPrice
	if err := r.db.First(&price, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &price, nil
}

func (r *PriceStore) GetByLevelAndProductAndSKU(levelID, productID, skuID uint) (*models.MemberLevelPrice, error) {
	var price models.MemberLevelPrice
	if err := r.db.Where("member_level_id = ? AND product_id = ? AND sku_id = ?", levelID, productID, skuID).First(&price).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &price, nil
}

// ListByProduct 获取商品的所有等级价
func (r *PriceStore) ListByProduct(productID uint) ([]models.MemberLevelPrice, error) {
	var prices []models.MemberLevelPrice
	if err := r.db.Where("product_id = ?", productID).Order("member_level_id asc, sku_id asc").Find(&prices).Error; err != nil {
		return nil, err
	}
	return prices, nil
}

// ListByLevelAndProducts 获取指定等级和商品集合的等级价
func (r *PriceStore) ListByLevelAndProducts(levelID uint, productIDs []uint) ([]models.MemberLevelPrice, error) {
	if len(productIDs) == 0 {
		return nil, nil
	}
	var prices []models.MemberLevelPrice
	if err := r.db.Where("member_level_id = ? AND product_id IN ?", levelID, productIDs).Find(&prices).Error; err != nil {
		return nil, err
	}
	return prices, nil
}

// BatchUpsert 批量创建/更新等级价
func (r *PriceStore) BatchUpsert(prices []models.MemberLevelPrice) error {
	if len(prices) == 0 {
		return nil
	}
	for _, p := range prices {
		existing, err := r.GetByLevelAndProductAndSKU(p.MemberLevelID, p.ProductID, p.SKUID)
		if err != nil {
			return err
		}
		if existing != nil {
			existing.PriceAmount = p.PriceAmount
			if err := r.db.Save(existing).Error; err != nil {
				return err
			}
		} else {
			if err := r.db.Create(&p).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *PriceStore) Delete(id uint) error {
	return r.db.Delete(&models.MemberLevelPrice{}, id).Error
}

func (r *PriceStore) DeleteByProduct(productID uint) error {
	return r.db.Where("product_id = ?", productID).Delete(&models.MemberLevelPrice{}).Error
}

// DeleteByProductInTx is the adapter contract consumed by product deletion.
func (r *PriceStore) DeleteByProductInTx(tx *gorm.DB, productID uint) error {
	if tx == nil {
		return r.DeleteByProduct(productID)
	}
	return NewPriceStore(tx).DeleteByProduct(productID)
}

var _ memberlevel.PriceRepository = (*PriceStore)(nil)
