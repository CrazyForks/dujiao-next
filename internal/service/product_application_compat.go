package service

import (
	"errors"

	catalogmapping "github.com/dujiao-next/internal/modules/catalog/mapping"
	mappinggormstore "github.com/dujiao-next/internal/modules/catalog/mapping/store/gormstore"
	productadmin "github.com/dujiao-next/internal/modules/catalog/product/application/admin"
	productwrite "github.com/dujiao-next/internal/modules/catalog/product/application/write"
	"github.com/dujiao-next/internal/repository"

	"gorm.io/gorm"
)

type MemberLevelPriceCleaner interface {
	DeleteByProductInTx(tx *gorm.DB, productID uint) error
}

// productWriteUnitOfWork 是旧 Repository 事务 API 到 Product Application 的临时适配器。
// Application 只看到绑定后的窄端口，不接触 *gorm.DB。
type productWriteUnitOfWork struct {
	products    repository.ProductRepository
	skus        repository.ProductSKURepository
	cardSecrets repository.CardSecretRepository
}

func newProductWriteUnitOfWork(
	products repository.ProductRepository,
	skus repository.ProductSKURepository,
	cardSecrets repository.CardSecretRepository,
) productwrite.UnitOfWork {
	return &productWriteUnitOfWork{
		products:    products,
		skus:        skus,
		cardSecrets: cardSecrets,
	}
}

func (unit *productWriteUnitOfWork) WithinTransaction(fn func(repositories productwrite.TransactionRepositories) error) error {
	if fn == nil {
		return nil
	}
	if unit == nil || unit.products == nil {
		return errors.New("product transaction repository is nil")
	}
	return unit.products.Transaction(func(tx *gorm.DB) error {
		var skus productwrite.SKURepository
		if unit.skus != nil {
			skus = unit.skus.WithTx(tx)
		}
		var cardSecrets productwrite.CardSecretStockRepository
		if unit.cardSecrets != nil {
			cardSecrets = unit.cardSecrets.WithTx(tx)
		}
		return fn(productwrite.TransactionRepositories{
			Products:    unit.products.WithTx(tx),
			SKUs:        skus,
			CardSecrets: cardSecrets,
		})
	})
}

// productAdminUnitOfWork 把旧仓储的 GORM 事务绑定收口在兼容边界内。
// Product Admin Application 只依赖各关联资源的删除端口。
type productAdminUnitOfWork struct {
	products          repository.ProductRepository
	productSKUs       repository.ProductSKURepository
	cardSecrets       repository.CardSecretRepository
	cardSecretBatches repository.CardSecretBatchRepository
	memberLevelPrices MemberLevelPriceCleaner
	carts             repository.CartRepository
	productMappings   catalogmapping.MappingRepository
}

func newProductAdminUnitOfWork(
	products repository.ProductRepository,
	productSKUs repository.ProductSKURepository,
	cardSecrets repository.CardSecretRepository,
	cardSecretBatches repository.CardSecretBatchRepository,
	memberLevelPrices MemberLevelPriceCleaner,
	carts repository.CartRepository,
	productMappings catalogmapping.MappingRepository,
) productadmin.UnitOfWork {
	return &productAdminUnitOfWork{
		products:          products,
		productSKUs:       productSKUs,
		cardSecrets:       cardSecrets,
		cardSecretBatches: cardSecretBatches,
		memberLevelPrices: memberLevelPrices,
		carts:             carts,
		productMappings:   productMappings,
	}
}

func (unit *productAdminUnitOfWork) WithinTransaction(fn func(productadmin.DeleteRepositories) error) error {
	if fn == nil {
		return nil
	}
	if unit == nil || unit.products == nil {
		return errors.New("product transaction repository is nil")
	}
	return unit.products.Transaction(func(tx *gorm.DB) error {
		return fn(productadmin.DeleteRepositories{
			Products:          unit.products.WithTx(tx),
			CardSecrets:       unit.cardSecrets.WithTx(tx),
			CardSecretBatches: unit.cardSecretBatches.WithTx(tx),
			SKUs:              unit.productSKUs.WithTx(tx),
			MemberLevelPrices: memberLevelPriceDeleteAdapter{tx: tx, cleaner: unit.memberLevelPrices},
			Carts:             unit.carts.WithTx(tx),
			ProductMappings:   bindMappingDeleteTx(unit.productMappings, tx),
		})
	})
}

func bindMappingDeleteTx(repo catalogmapping.MappingRepository, tx *gorm.DB) productadmin.ProductMappingDeleteRepository {
	switch binder := repo.(type) {
	case interface {
		WithTx(tx *gorm.DB) *mappinggormstore.MappingStore
	}:
		return binder.WithTx(tx)
	case interface {
		WithTx(tx *gorm.DB) repository.ProductMappingRepository
	}:
		return binder.WithTx(tx)
	default:
		return repo
	}
}

type memberLevelPriceDeleteAdapter struct {
	tx      *gorm.DB
	cleaner MemberLevelPriceCleaner
}

func (adapter memberLevelPriceDeleteAdapter) DeleteByProduct(productID uint) error {
	return adapter.cleaner.DeleteByProductInTx(adapter.tx, productID)
}
