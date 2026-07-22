package catalogproductbootstrap

import (
	"errors"

	categorycontract "github.com/dujiao-next/internal/modules/catalog/category/contract"

	catalogmapping "github.com/dujiao-next/internal/modules/catalog/mapping"
	catalogproduct "github.com/dujiao-next/internal/modules/catalog/product"
	productapplication "github.com/dujiao-next/internal/modules/catalog/product/application"
	productadmin "github.com/dujiao-next/internal/modules/catalog/product/application/admin"
	productwrite "github.com/dujiao-next/internal/modules/catalog/product/application/write"
	"github.com/dujiao-next/internal/repository"

	"gorm.io/gorm"
)

type memberLevelPriceCleaner interface {
	DeleteByProductInTx(tx *gorm.DB, productID uint) error
}

// ProductStore 是装配层需要的 Product 持久化与事务能力。
type ProductStore interface {
	catalogproduct.Repository
	Transaction(fn func(tx *gorm.DB) error) error
	BindTx(tx *gorm.DB) catalogproduct.Repository
}

// SKUStore 是装配层需要的 SKU 持久化与事务绑定能力。
type SKUStore interface {
	catalogproduct.SKURepository
	BindTx(tx *gorm.DB) catalogproduct.SKURepository
}

// MappingStore 是商品级联删除所需的映射持久化与事务能力。
type MappingStore interface {
	catalogmapping.MappingRepository
	BindTx(tx *gorm.DB) catalogmapping.MappingRepository
}

// Dependencies 是 Product 三组应用用例的装配依赖。
type Dependencies struct {
	Products          ProductStore
	SKUs              SKUStore
	CardSecrets       repository.CardSecretRepository
	CardSecretBatches repository.CardSecretBatchRepository
	Categories        categorycontract.Repository
	MemberLevelPrices memberLevelPriceCleaner
	Carts             repository.CartRepository
	ProductMappings   MappingStore
	Orders            repository.OrderRepository
	PaymentChannels   repository.PaymentChannelRepository
}

// Services 是 Product 查询、管理和写入用例的显式集合。
type Services struct {
	Read  *productapplication.Service
	Admin *productadmin.AdminService
	Write *productwrite.WriteService
}

// productWriteUnitOfWork 将 Product Application 所需端口绑定到同一事务。
type productWriteUnitOfWork struct {
	products    ProductStore
	skus        SKUStore
	cardSecrets repository.CardSecretRepository
}

func newProductWriteUnitOfWork(
	products ProductStore,
	skus SKUStore,
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
			skus = unit.skus.BindTx(tx)
		}
		var cardSecrets productwrite.CardSecretStockRepository
		if unit.cardSecrets != nil {
			cardSecrets = unit.cardSecrets.WithTx(tx)
		}
		return fn(productwrite.TransactionRepositories{
			Products:    unit.products.BindTx(tx),
			SKUs:        skus,
			CardSecrets: cardSecrets,
		})
	})
}

// productAdminUnitOfWork 将商品级联删除涉及的端口绑定到同一事务。
type productAdminUnitOfWork struct {
	products          ProductStore
	productSKUs       SKUStore
	cardSecrets       repository.CardSecretRepository
	cardSecretBatches repository.CardSecretBatchRepository
	memberLevelPrices memberLevelPriceCleaner
	carts             repository.CartRepository
	productMappings   MappingStore
}

func newProductAdminUnitOfWork(
	products ProductStore,
	productSKUs SKUStore,
	cardSecrets repository.CardSecretRepository,
	cardSecretBatches repository.CardSecretBatchRepository,
	memberLevelPrices memberLevelPriceCleaner,
	carts repository.CartRepository,
	productMappings MappingStore,
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
			Products:          unit.products.BindTx(tx),
			CardSecrets:       unit.cardSecrets.WithTx(tx),
			CardSecretBatches: unit.cardSecretBatches.WithTx(tx),
			SKUs:              unit.productSKUs.BindTx(tx),
			MemberLevelPrices: memberLevelPriceDeleteAdapter{tx: tx, cleaner: unit.memberLevelPrices},
			Carts:             unit.carts.WithTx(tx),
			ProductMappings:   bindMappingDeleteTx(unit.productMappings, tx),
		})
	})
}

func bindMappingDeleteTx(repo MappingStore, tx *gorm.DB) productadmin.ProductMappingDeleteRepository {
	if repo == nil {
		return nil
	}
	return repo.BindTx(tx)
}

type memberLevelPriceDeleteAdapter struct {
	tx      *gorm.DB
	cleaner memberLevelPriceCleaner
}

func (adapter memberLevelPriceDeleteAdapter) DeleteByProduct(productID uint) error {
	return adapter.cleaner.DeleteByProductInTx(adapter.tx, productID)
}

// New 显式装配 Product 查询、管理和写入用例。
func New(dependencies Dependencies) Services {
	read := productapplication.NewService(productapplication.Options{
		Products:                      dependencies.Products,
		Categories:                    dependencies.Categories,
		Stock:                         dependencies.CardSecrets,
		NotFoundError:                 catalogproduct.ErrNotFound,
		ResellerProductNotListedError: catalogproduct.ErrResellerProductNotListed,
	})
	admin := productadmin.NewAdminService(productadmin.Options{
		Products:     dependencies.Products,
		Categories:   dependencies.Categories,
		CardSecrets:  dependencies.CardSecrets,
		Orders:       dependencies.Orders,
		Transactions: newProductAdminUnitOfWork(dependencies.Products, dependencies.SKUs, dependencies.CardSecrets, dependencies.CardSecretBatches, dependencies.MemberLevelPrices, dependencies.Carts, dependencies.ProductMappings),
		Errors: productadmin.ErrorSet{
			NotFound:               catalogproduct.ErrNotFound,
			ProductCategoryInvalid: catalogproduct.ErrProductCategoryInvalid,
			ProductHasStock:        catalogproduct.ErrProductHasStock,
			ProductHasOrderRecord:  catalogproduct.ErrProductHasOrderRecord,
		},
	})
	write := productwrite.NewWriteService(productwrite.Options{
		Products:        dependencies.Products,
		SKUs:            dependencies.SKUs,
		Categories:      dependencies.Categories,
		PaymentChannels: dependencies.PaymentChannels,
		Transactions:    newProductWriteUnitOfWork(dependencies.Products, dependencies.SKUs, dependencies.CardSecrets),
		Errors: productwrite.ErrorSet{
			NotFound:                     catalogproduct.ErrNotFound,
			SlugExists:                   catalogproduct.ErrSlugExists,
			ProductCategoryInvalid:       catalogproduct.ErrProductCategoryInvalid,
			ProductPurchaseInvalid:       catalogproduct.ErrProductPurchaseInvalid,
			FulfillmentInvalid:           catalogproduct.ErrFulfillmentInvalid,
			ProductPriceInvalid:          catalogproduct.ErrProductPriceInvalid,
			ManualStockInvalid:           catalogproduct.ErrManualStockInvalid,
			ProductPurchaseLimitInvalid:  catalogproduct.ErrProductPurchaseLimitInvalid,
			ProductStockDisplayInvalid:   catalogproduct.ErrProductStockDisplayInvalid,
			ProductSKUInvalid:            catalogproduct.ErrProductSKUInvalid,
			ProductSKUHasCardSecretStock: catalogproduct.ErrProductSKUHasCardSecretStock,
		},
	})
	return Services{Read: read, Admin: admin, Write: write}
}
