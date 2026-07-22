package repository

import (
	"github.com/dujiao-next/internal/persistence/gormutil"

	"gorm.io/gorm"
)

func applyPagination(query *gorm.DB, page, pageSize int) *gorm.DB {
	return gormutil.ApplyPagination(query, page, pageSize)
}
