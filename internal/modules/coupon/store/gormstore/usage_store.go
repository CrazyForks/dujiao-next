package gormstore

import (
	"github.com/dujiao-next/internal/models"
	"github.com/dujiao-next/internal/modules/coupon"

	"gorm.io/gorm"
)

// UsageStore GORM 优惠券使用记录存储。
type UsageStore struct {
	db *gorm.DB
}

func NewUsageStore(db *gorm.DB) *UsageStore {
	return &UsageStore{db: db}
}

// WithTx 绑定事务
func (r *UsageStore) WithTx(tx *gorm.DB) coupon.UsageRepository {
	if tx == nil {
		return r
	}
	return &UsageStore{db: tx}
}

// Create 创建使用记录
func (r *UsageStore) Create(usage *models.CouponUsage) error {
	return r.db.Create(usage).Error
}

// CountByUser 获取用户使用次数
func (r *UsageStore) CountByUser(couponID, userID uint) (int64, error) {
	var count int64
	if err := r.db.Model(&models.CouponUsage{}).
		Where("coupon_id = ? AND user_id = ?", couponID, userID).
		Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// ListByOrderID 获取订单使用记录
func (r *UsageStore) ListByOrderID(orderID uint) ([]models.CouponUsage, error) {
	var usages []models.CouponUsage
	if err := r.db.Where("order_id = ?", orderID).Find(&usages).Error; err != nil {
		return nil, err
	}
	return usages, nil
}

// ListByUser 获取用户使用记录
func (r *UsageStore) ListByUser(filter coupon.UsageListFilter) ([]models.CouponUsage, int64, error) {
	query := r.db.Model(&models.CouponUsage{}).Where("user_id = ?", filter.UserID)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	query = applyPagination(query, filter.Page, filter.PageSize)

	var usages []models.CouponUsage
	if err := query.Order("id desc").Find(&usages).Error; err != nil {
		return nil, 0, err
	}
	return usages, total, nil
}

// DeleteByOrderID 删除订单使用记录
func (r *UsageStore) DeleteByOrderID(orderID uint) error {
	return r.db.Where("order_id = ?", orderID).Delete(&models.CouponUsage{}).Error
}

var _ coupon.UsageRepository = (*UsageStore)(nil)
