package catalogmappingbootstrap

import (
	"errors"

	categorycontract "github.com/dujiao-next/internal/modules/catalog/category/contract"

	catalogmapping "github.com/dujiao-next/internal/modules/catalog/mapping"
	catalogproduct "github.com/dujiao-next/internal/modules/catalog/product"
	"github.com/dujiao-next/internal/modules/siteconnection"

	"gorm.io/gorm"
)

// ProductStore 是映射服务装配所需的商品持久化与事务能力。
type ProductStore interface {
	catalogproduct.Repository
	Transaction(fn func(tx *gorm.DB) error) error
	BindTx(tx *gorm.DB) catalogproduct.Repository
}

// SKUStore 是映射服务装配所需的 SKU 持久化与事务绑定能力。
type SKUStore interface {
	catalogproduct.SKURepository
	BindTx(tx *gorm.DB) catalogproduct.SKURepository
}

// MappingStore 是映射持久化与事务绑定能力。
type MappingStore interface {
	catalogmapping.MappingRepository
	BindTx(tx *gorm.DB) catalogmapping.MappingRepository
}

// SKUMappingStore 是 SKU 映射持久化与事务绑定能力。
type SKUMappingStore interface {
	catalogmapping.SKUMappingRepository
	BindTx(tx *gorm.DB) catalogmapping.SKUMappingRepository
}

// Dependencies 集中声明 Catalog Mapping 模块的启动装配依赖。
type Dependencies struct {
	Mappings    MappingStore
	SKUMappings SKUMappingStore
	Products    ProductStore
	SKUs        SKUStore
	Categories  categorycontract.Repository
	Connections *siteconnection.Service
	Media       catalogmapping.MediaRecorder
}

// New 创建可直接注入调用方的 Catalog Mapping 应用服务。
func New(dependencies Dependencies) (*catalogmapping.Service, error) {
	core, err := catalogmapping.NewService(catalogmapping.Options{
		Mappings:     dependencies.Mappings,
		SKUMappings:  dependencies.SKUMappings,
		Products:     dependencies.Products,
		SKUs:         dependencies.SKUs,
		Categories:   dependencies.Categories,
		Connections:  dependencies.Connections,
		Media:        dependencies.Media,
		Transactions: newUnitOfWork(dependencies.Products, dependencies.SKUs, dependencies.Mappings, dependencies.SKUMappings),
		Errors: catalogmapping.ErrorSet{
			ConnectionNotFound:     siteconnection.ErrNotFound,
			ProductCategoryInvalid: catalogproduct.ErrProductCategoryInvalid,
		},
	})
	if err != nil {
		return nil, err
	}
	return core, nil
}

// unitOfWork 把商品与映射写入端口绑定到同一数据库事务。
type unitOfWork struct {
	products    ProductStore
	skus        SKUStore
	mappings    MappingStore
	skuMappings SKUMappingStore
}

func newUnitOfWork(
	products ProductStore,
	skus SKUStore,
	mappings MappingStore,
	skuMappings SKUMappingStore,
) catalogmapping.UnitOfWork {
	return &unitOfWork{
		products:    products,
		skus:        skus,
		mappings:    mappings,
		skuMappings: skuMappings,
	}
}

func (unit *unitOfWork) WithinTransaction(fn func(catalogmapping.ImportRepositories) error) error {
	if fn == nil {
		return nil
	}
	if unit == nil || unit.products == nil {
		return errors.New("product transaction repository is nil")
	}
	return unit.products.Transaction(func(tx *gorm.DB) error {
		var skus catalogmapping.ImportTxSKURepository
		if unit.skus != nil {
			skus = unit.skus.BindTx(tx)
		}
		var mappings catalogmapping.ImportTxMappingRepository
		if unit.mappings != nil {
			mappings = unit.mappings.BindTx(tx)
		}
		var skuMappings catalogmapping.ImportTxSKUMappingRepository
		if unit.skuMappings != nil {
			skuMappings = unit.skuMappings.BindTx(tx)
		}
		return fn(catalogmapping.ImportRepositories{
			Products:    unit.products.BindTx(tx),
			SKUs:        skus,
			Mappings:    mappings,
			SKUMappings: skuMappings,
		})
	})
}
