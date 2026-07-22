package promotion

import (
	"time"

	"github.com/dujiao-next/internal/models"
)

// ListFilter 定义管理端活动价列表的查询条件。
type ListFilter struct {
	ID         uint
	Name       string
	ScopeRefID uint
	IsActive   *bool
	Page       int
	PageSize   int
}

// Repository 定义 Promotion 领域所需的持久化能力。
type Repository interface {
	GetByID(id uint) (*models.Promotion, error)
	GetActiveByProduct(productID uint, now time.Time) (*models.Promotion, error)
	GetAllActiveByProduct(productID uint, now time.Time) ([]models.Promotion, error)
	Create(promotion *models.Promotion) error
	Update(promotion *models.Promotion) error
	Delete(id uint) error
	List(filter ListFilter) ([]models.Promotion, int64, error)
}
