package gormstore

import (
	"time"

	"github.com/dujiao-next/internal/constants"
	"github.com/dujiao-next/internal/models"
	"github.com/dujiao-next/internal/modules/dashboard"

	"gorm.io/gorm"
)

// Store is the GORM adapter for dashboard persistence ports.
type Store struct {
	db *gorm.DB
}

func New(db *gorm.DB) *Store {
	return &Store{db: db}
}

func paidOrderStatuses() []string {
	return []string{
		constants.OrderStatusPaid,
		constants.OrderStatusFulfilling,
		constants.OrderStatusPartiallyDelivered,
		constants.OrderStatusPartiallyRefunded,
		constants.OrderStatusDelivered,
		constants.OrderStatusCompleted,
	}
}

func onlinePaymentBase(db *gorm.DB, startAt, endAt time.Time) *gorm.DB {
	return db.Model(&models.Payment{}).
		Where("created_at >= ? AND created_at < ? AND provider_type <> ?", startAt, endAt, constants.PaymentProviderWallet)
}

var _ dashboard.Repository = (*Store)(nil)
