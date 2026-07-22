package memberlevel

import (
	"github.com/dujiao-next/internal/models"
	"github.com/shopspring/decimal"
)

type ListFilter struct {
	IsActive *bool
	Page     int
	PageSize int
}

type LevelRepository interface {
	GetByID(id uint) (*models.MemberLevel, error)
	GetBySlug(slug string) (*models.MemberLevel, error)
	GetDefault() (*models.MemberLevel, error)
	GetActiveBySortOrder(sortOrder int, excludeID uint) (*models.MemberLevel, error)
	ListAllActive() ([]models.MemberLevel, error)
	Create(level *models.MemberLevel) error
	Update(level *models.MemberLevel) error
	Delete(id uint) error
	List(filter ListFilter) ([]models.MemberLevel, int64, error)
	ClearDefault(excludeID uint) error
}

type PriceRepository interface {
	GetByID(id uint) (*models.MemberLevelPrice, error)
	GetByLevelAndProductAndSKU(levelID, productID, skuID uint) (*models.MemberLevelPrice, error)
	ListByProduct(productID uint) ([]models.MemberLevelPrice, error)
	ListByLevelAndProducts(levelID uint, productIDs []uint) ([]models.MemberLevelPrice, error)
	BatchUpsert(prices []models.MemberLevelPrice) error
	Delete(id uint) error
	DeleteByProduct(productID uint) error
}

// UserRepository is the member-level consumer's minimal user persistence port.
type UserRepository interface {
	GetByID(id uint) (*models.User, error)
	Update(user *models.User) error
	IncrementTotalRecharged(userID uint, amount decimal.Decimal) error
	IncrementTotalSpent(userID uint, amount decimal.Decimal) error
	UpdateMemberLevelIfCurrent(userID, currentLevelID, nextLevelID uint) (int64, error)
	AssignDefaultMemberLevel(defaultLevelID uint) (int64, error)
}
