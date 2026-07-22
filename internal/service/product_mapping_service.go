package service

import (
	"errors"

	settingsapp "github.com/dujiao-next/internal/modules/settings/application"

	"github.com/dujiao-next/internal/modules/catalog"
	catalogmapping "github.com/dujiao-next/internal/modules/catalog/mapping"
	mappinggormstore "github.com/dujiao-next/internal/modules/catalog/mapping/store/gormstore"
	"github.com/dujiao-next/internal/modules/siteconnection"
	"github.com/dujiao-next/internal/repository"

	"gorm.io/gorm"
)

// 商品映射实现已迁入 Catalog 上游映射上下文（internal/modules/catalog/mapping），
// 持久化在 internal/modules/catalog/mapping/store/gormstore。
// 本文件仅保留兼容门面：装配模块服务与导入事务工作单元。

// 错误别名指向模块哨兵，保留现有调用方的 errors.Is 语义。
var (
	ErrMappingNotFound         = catalogmapping.ErrMappingNotFound
	ErrMappingAlreadyExists    = catalogmapping.ErrMappingAlreadyExists
	ErrUpstreamProductNotFound = catalogmapping.ErrUpstreamProductNotFound
	ErrMappingInactive         = catalogmapping.ErrMappingInactive
	ErrMediaRecorderRequired   = catalogmapping.ErrMediaRecorderRequired
)

// LocalMediaRecorder 是 ProductMapping 下载图片后所需的最小 Content 写入接口。
type LocalMediaRecorder = catalogmapping.MediaRecorder

type BatchUpstreamProductImportOutcome = catalogmapping.BatchUpstreamProductImportOutcome
type BatchImportByCategoryResult = catalogmapping.BatchImportByCategoryResult

// ProductMappingService 商品映射业务服务（兼容门面）
type ProductMappingService struct {
	*catalogmapping.Service
}

// NewProductMappingService 创建商品映射服务
func NewProductMappingService(
	mappingRepo catalogmapping.MappingRepository,
	skuMappingRepo catalogmapping.SKUMappingRepository,
	productRepo repository.ProductRepository,
	productSKURepo repository.ProductSKURepository,
	categoryRepo catalog.CategoryRepository,
	connService *siteconnection.Service,
	mediaRecorder LocalMediaRecorder,
) (*ProductMappingService, error) {
	core, err := catalogmapping.NewService(catalogmapping.Options{
		Mappings:     mappingRepo,
		SKUMappings:  skuMappingRepo,
		Products:     productRepo,
		SKUs:         productSKURepo,
		Categories:   categoryRepo,
		Connections:  connService,
		Media:        mediaRecorder,
		Transactions: newProductMappingUnitOfWork(productRepo, productSKURepo, mappingRepo, skuMappingRepo),
		Errors: catalogmapping.ErrorSet{
			ConnectionNotFound:     siteconnection.ErrNotFound,
			ProductCategoryInvalid: ErrProductCategoryInvalid,
		},
	})
	if err != nil {
		return nil, err
	}
	return &ProductMappingService{Service: core}, nil
}

// SetCategoryService 设置分类服务（避免循环依赖）
func (s *ProductMappingService) SetCategoryService(cs *catalog.CategoryService) {
	if cs == nil {
		return
	}
	s.Service.SetCategoryCreator(cs)
}

// SetSettingService 注入设置服务（用于读取上游同步动态配置）
func (s *ProductMappingService) SetSettingService(ss *settingsapp.Service) {
	if ss == nil {
		return
	}
	s.Service.SetSettings(ss)
}

// productMappingUnitOfWork 把商品事务与映射端口绑定收口在兼容边界内。
type productMappingUnitOfWork struct {
	products    repository.ProductRepository
	skus        repository.ProductSKURepository
	mappings    catalogmapping.MappingRepository
	skuMappings catalogmapping.SKUMappingRepository
}

func newProductMappingUnitOfWork(
	products repository.ProductRepository,
	skus repository.ProductSKURepository,
	mappings catalogmapping.MappingRepository,
	skuMappings catalogmapping.SKUMappingRepository,
) catalogmapping.UnitOfWork {
	return &productMappingUnitOfWork{
		products:    products,
		skus:        skus,
		mappings:    mappings,
		skuMappings: skuMappings,
	}
}

func (unit *productMappingUnitOfWork) WithinTransaction(fn func(catalogmapping.ImportRepositories) error) error {
	if fn == nil {
		return nil
	}
	if unit == nil || unit.products == nil {
		return errors.New("product transaction repository is nil")
	}
	return unit.products.Transaction(func(tx *gorm.DB) error {
		var skus catalogmapping.ImportTxSKURepository
		if unit.skus != nil {
			skus = unit.skus.WithTx(tx)
		}
		return fn(catalogmapping.ImportRepositories{
			Products:    unit.products.WithTx(tx),
			SKUs:        skus,
			Mappings:    bindMappingImportTx(unit.mappings, tx),
			SKUMappings: bindSKUMappingImportTx(unit.skuMappings, tx),
		})
	})
}

func bindMappingImportTx(repo catalogmapping.MappingRepository, tx *gorm.DB) catalogmapping.ImportTxMappingRepository {
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

func bindSKUMappingImportTx(repo catalogmapping.SKUMappingRepository, tx *gorm.DB) catalogmapping.ImportTxSKUMappingRepository {
	switch binder := repo.(type) {
	case interface {
		WithTx(tx *gorm.DB) *mappinggormstore.SKUMappingStore
	}:
		return binder.WithTx(tx)
	case interface {
		WithTx(tx *gorm.DB) repository.SKUMappingRepository
	}:
		return binder.WithTx(tx)
	default:
		return repo
	}
}
