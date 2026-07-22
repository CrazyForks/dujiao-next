package repository

import (
	"github.com/dujiao-next/internal/persistence/gormutil"

	"gorm.io/gorm"
)

func reserveManualStock(db *gorm.DB, model interface{}, id uint, quantity int) (int64, error) {
	return gormutil.ReserveManualStock(db, model, id, quantity)
}

func releaseManualStock(db *gorm.DB, model interface{}, id uint, quantity int) (int64, error) {
	return gormutil.ReleaseManualStock(db, model, id, quantity)
}

func consumeManualStock(db *gorm.DB, model interface{}, id uint, quantity int) (int64, error) {
	return gormutil.ConsumeManualStock(db, model, id, quantity)
}
