package coupon

import "github.com/dujiao-next/internal/models"

type ListFilter struct {
	ID         uint
	Code       string
	ScopeRefID uint
	IsActive   *bool
	Page       int
	PageSize   int
}

type UsageListFilter struct {
	UserID   uint
	Page     int
	PageSize int
}

type Repository interface {
	GetByID(id uint) (*models.Coupon, error)
	GetByCode(code string) (*models.Coupon, error)
	ListByIDs(ids []uint) ([]models.Coupon, error)
	Create(coupon *models.Coupon) error
	Update(coupon *models.Coupon) error
	Delete(id uint) error
	List(filter ListFilter) ([]models.Coupon, int64, error)
	IncrementUsedCount(id uint, delta int) error
	DecrementUsedCount(id uint, delta int) error
}

type UsageRepository interface {
	Create(usage *models.CouponUsage) error
	CountByUser(couponID, userID uint) (int64, error)
	ListByOrderID(orderID uint) ([]models.CouponUsage, error)
	ListByUser(filter UsageListFilter) ([]models.CouponUsage, int64, error)
	DeleteByOrderID(orderID uint) error
}
