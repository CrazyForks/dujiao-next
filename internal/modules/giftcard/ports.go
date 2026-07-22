package giftcard

import (
	"time"

	userdomain "github.com/dujiao-next/internal/modules/identity/user/domain"

	"github.com/dujiao-next/internal/models"
)

// ListFilter 礼品卡仓储列表筛选。
type ListFilter struct {
	Code           string
	Status         string
	BatchNo        string
	RedeemedUserID uint
	CreatedFrom    *time.Time
	CreatedTo      *time.Time
	RedeemedFrom   *time.Time
	RedeemedTo     *time.Time
	ExpiresFrom    *time.Time
	ExpiresTo      *time.Time
	Page           int
	PageSize       int
}

// Repository 是礼品卡管理用例的数据端口。
type Repository interface {
	CreateBatch(batch *models.GiftCardBatch, cards []models.GiftCard) error
	GetByID(id uint) (*models.GiftCard, error)
	List(filter ListFilter) ([]models.GiftCard, int64, error)
	ListByIDs(ids []uint) ([]models.GiftCard, error)
	Update(card *models.GiftCard) error
	Delete(id uint) error
	BatchUpdateStatus(ids []uint, status string, updatedAt time.Time) (int64, error)
	WithinTransaction(fn func(repo Repository) error) error
}

// UserDirectory 是兑换用户解析端口。
type UserDirectory interface {
	ListByIDs(ids []uint) ([]userdomain.User, error)
}

// CurrencyProvider 是站点币种读取端口。
type CurrencyProvider interface {
	SiteCurrency() string
}
