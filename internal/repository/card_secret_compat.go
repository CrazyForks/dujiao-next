package repository

import (
	"github.com/dujiao-next/internal/modules/cardsecret"
	cardsecretgormstore "github.com/dujiao-next/internal/modules/cardsecret/store/gormstore"

	"gorm.io/gorm"
)

// Card Secret 的领域与 GORM 实现已迁入 modules/cardsecret。
// 这些别名和构造器只服务于尚未迁移的 Order/Product/Fulfillment 消费方；
// 对应领域完成迁移后，本兼容门面可以整体删除。
type CardSecretListFilter = cardsecret.ListFilter
type CardSecretBatchStatusCount = cardsecret.BatchStatusCount
type SKUStockCount = cardsecret.SKUStockCount

type CardSecretRepository interface {
	cardsecret.Repository
	cardsecret.UnitOfWork
	WithTx(tx *gorm.DB) *cardsecretgormstore.Store
}

type CardSecretBatchRepository interface {
	cardsecret.BatchRepository
	WithTx(tx *gorm.DB) *cardsecretgormstore.BatchStore
}

func NewCardSecretRepository(db *gorm.DB) *cardsecretgormstore.Store {
	return cardsecretgormstore.New(db)
}

func NewCardSecretBatchRepository(db *gorm.DB) *cardsecretgormstore.BatchStore {
	return cardsecretgormstore.NewBatch(db)
}
