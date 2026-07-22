package mapping

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/dujiao-next/internal/constants"
	"github.com/dujiao-next/internal/models"
	"github.com/dujiao-next/internal/modules/catalog"
	catalogproduct "github.com/dujiao-next/internal/modules/catalog/product"
	productdomain "github.com/dujiao-next/internal/modules/catalog/product/domain"
	"github.com/dujiao-next/internal/modules/settings"
	"github.com/dujiao-next/internal/upstream"
)

// 文件组织约定:
//   service.go       — 端口/错误/装配 + 基础查询与 CRUD
//   import.go        — 单品导入 + 上游元数据列表 + 图片下载
//   batch_import.go  — 批量导入与上游分类自动创建
//   sync.go          — 同步流程（单品 / 全量库存 / 下单前兜底）
//   markup.go        — 加价重算
//   pricing.go       — 汇率与加价换算

var (
	ErrMappingNotFound           = errors.New("product mapping not found")
	ErrMappingAlreadyExists      = errors.New("product mapping already exists for this upstream product")
	ErrUpstreamProductNotFound   = errors.New("upstream product not found")
	ErrMappingInactive           = errors.New("product mapping is inactive")
	ErrMediaRecorderRequired     = errors.New("product mapping media recorder is required")
	ErrUpstreamStockInsufficient = errors.New("upstream stock insufficient")
	ErrConnectionNotFound        = errors.New("site connection not found")
)

// ListFilter 商品映射列表筛选条件。
type ListFilter struct {
	ConnectionID   uint
	UpstreamStatus string // active / inactive / deleted，空值不筛选
	ProductStatus  string // active / inactive，空值不筛选
	Search         string // 商品名称模糊搜索，空值不筛选
	Page           int
	PageSize       int
}

// MappingRepository 商品映射持久化端口。
type MappingRepository interface {
	GetByID(id uint) (*models.ProductMapping, error)
	GetByLocalProductID(productID uint) (*models.ProductMapping, error)
	GetByConnectionAndUpstreamID(connectionID, upstreamProductID uint) (*models.ProductMapping, error)
	Create(mapping *models.ProductMapping) error
	Update(mapping *models.ProductMapping) error
	Delete(id uint) error
	DeleteByLocalProduct(productID uint) error
	List(filter ListFilter) ([]models.ProductMapping, int64, error)
	ListByLocalProductIDs(productIDs []uint) ([]models.ProductMapping, error)
	ListActiveByConnection(connectionID uint) ([]models.ProductMapping, error)
	ListAllActive() ([]models.ProductMapping, error)
	ListUpstreamIDsByConnection(connectionID uint) ([]uint, error)
}

// SKUMappingRepository SKU 映射持久化端口。
type SKUMappingRepository interface {
	GetByLocalSKUID(skuID uint) (*models.SKUMapping, error)
	ListByProductMapping(productMappingID uint) ([]models.SKUMapping, error)
	ListByProductMappingIDs(productMappingIDs []uint) ([]models.SKUMapping, error)
	Create(mapping *models.SKUMapping) error
	Update(mapping *models.SKUMapping) error
	DeleteByProductMapping(productMappingID uint) error
}

// ProductRepository 是映射上下文所需的最小本地商品端口。
type ProductRepository interface {
	GetByID(id string) (*models.Product, error)
	Update(item *models.Product) error
	QuickUpdate(id string, fields map[string]interface{}) error
}

// SKURepository 是映射上下文所需的最小本地 SKU 端口。
type SKURepository interface {
	GetByID(id uint) (*models.ProductSKU, error)
	ListByProduct(productID uint, onlyActive bool) ([]models.ProductSKU, error)
	Create(item *models.ProductSKU) error
	Update(item *models.ProductSKU) error
}

// CategoryRepository 分类查找与复活端口（含商品分类归属校验）。
type CategoryRepository interface {
	productdomain.CategoryAssignmentRepository
	GetBySlug(slug string) (*models.Category, error)
	GetBySlugUnscoped(slug string) (*models.Category, error)
	Restore(category *models.Category) error
}

// ConnectionProvider 隔离站点连接的读取与上游协议适配器构造。
type ConnectionProvider interface {
	GetByID(id uint) (*models.SiteConnection, error)
	GetAdapter(conn *models.SiteConnection) (upstream.Adapter, error)
}

// MediaRecorder 是下载上游图片后所需的最小 Content 写入端口。
type MediaRecorder interface {
	RecordLocalFile(ctx context.Context, localPath, scene string)
}

// CategoryCreator 自动建分类端口，由 Catalog CategoryService 实现，setter 注入避免装配顺序耦合。
type CategoryCreator interface {
	Create(input catalog.CreateCategoryInput) (*models.Category, error)
}

// SettingsProvider 读取上游同步动态配置。
type SettingsProvider interface {
	GetUpstreamSyncConfig(fallbackInterval string) (settings.UpstreamSyncConfig, error)
	GetUpstreamSyncInterval(fallbackInterval string) (time.Duration, error)
}

// ImportTxProductRepository 导入事务内的本地商品写入端口。
type ImportTxProductRepository interface {
	Create(item *models.Product) error
	QuickUpdate(id string, fields map[string]interface{}) error
}

// ImportTxSKURepository 导入事务内的本地 SKU 写入端口。
type ImportTxSKURepository interface {
	Create(item *models.ProductSKU) error
}

// ImportTxMappingRepository 导入事务内的映射写入端口。
type ImportTxMappingRepository interface {
	Create(mapping *models.ProductMapping) error
}

// ImportTxSKUMappingRepository 导入事务内的 SKU 映射写入端口。
type ImportTxSKUMappingRepository interface {
	Create(mapping *models.SKUMapping) error
}

// ImportRepositories 是一次上游商品导入事务中绑定的全部窄端口。
type ImportRepositories struct {
	Products    ImportTxProductRepository
	SKUs        ImportTxSKURepository
	Mappings    ImportTxMappingRepository
	SKUMappings ImportTxSKUMappingRepository
}

// UnitOfWork 隔离应用层与具体数据库事务实现。
type UnitOfWork interface {
	WithinTransaction(fn func(ImportRepositories) error) error
}

// ErrorSet 保留旧服务层公开共享错误的 errors.Is 身份。
type ErrorSet struct {
	ConnectionNotFound     error
	ProductCategoryInvalid error
}

type Options struct {
	Mappings     MappingRepository
	SKUMappings  SKUMappingRepository
	Products     ProductRepository
	SKUs         SKURepository
	Categories   CategoryRepository
	Connections  ConnectionProvider
	Media        MediaRecorder
	Transactions UnitOfWork
	Errors       ErrorSet
}

// Service 承载本地商品与上游站点之间的映射用例。
type Service struct {
	mappings        MappingRepository
	skuMappings     SKUMappingRepository
	products        ProductRepository
	skus            SKURepository
	categories      CategoryRepository
	connections     ConnectionProvider
	media           MediaRecorder
	transactions    UnitOfWork
	categoryCreator CategoryCreator
	settings        SettingsProvider
	errors          ErrorSet
}

func NewService(options Options) (*Service, error) {
	if options.Media == nil {
		return nil, ErrMediaRecorderRequired
	}
	return &Service{
		mappings:     options.Mappings,
		skuMappings:  options.SKUMappings,
		products:     options.Products,
		skus:         options.SKUs,
		categories:   options.Categories,
		connections:  options.Connections,
		media:        options.Media,
		transactions: options.Transactions,
		errors:       resolveErrorSet(options.Errors),
	}, nil
}

func resolveErrorSet(values ErrorSet) ErrorSet {
	return ErrorSet{
		ConnectionNotFound:     errorOrDefault(values.ConnectionNotFound, ErrConnectionNotFound),
		ProductCategoryInvalid: errorOrDefault(values.ProductCategoryInvalid, catalogproduct.ErrProductCategoryInvalid),
	}
}

func errorOrDefault(value, fallback error) error {
	if value != nil {
		return value
	}
	return fallback
}

// SetCategoryCreator 注入分类创建端口（装配时调用）。
func (s *Service) SetCategoryCreator(creator CategoryCreator) {
	s.categoryCreator = creator
}

// SetSettings 注入设置端口（用于读取上游同步动态配置）。
func (s *Service) SetSettings(settings SettingsProvider) {
	s.settings = settings
}

// GetByID 获取映射详情
func (s *Service) GetByID(id uint) (*models.ProductMapping, error) {
	return s.mappings.GetByID(id)
}

// List 列表查询映射
func (s *Service) List(filter ListFilter) ([]models.ProductMapping, int64, error) {
	return s.mappings.List(filter)
}

// SetActive 启用/禁用映射
func (s *Service) SetActive(id uint, active bool) error {
	mapping, err := s.mappings.GetByID(id)
	if err != nil {
		return err
	}
	if mapping == nil {
		return ErrMappingNotFound
	}
	mapping.IsActive = active
	return s.mappings.Update(mapping)
}

// Delete 删除映射（不删除本地商品）
func (s *Service) Delete(id uint) error {
	mapping, err := s.mappings.GetByID(id)
	if err != nil {
		return err
	}
	if mapping == nil {
		return ErrMappingNotFound
	}

	// 删除 SKU 映射
	if err := s.skuMappings.DeleteByProductMapping(id); err != nil {
		return err
	}

	// 还原本地商品状态：取消映射标记、交付类型改回 manual、自动下架
	if mapping.LocalProductID > 0 {
		localProduct, err := s.products.GetByID(strconv.FormatUint(uint64(mapping.LocalProductID), 10))
		if err == nil && localProduct != nil {
			localProduct.IsMapped = false
			if localProduct.FulfillmentType == constants.FulfillmentTypeUpstream {
				localProduct.FulfillmentType = constants.FulfillmentTypeManual
				localProduct.IsActive = false // 下架，防止用户下单后无法交付
			}
			_ = s.products.Update(localProduct)
		}
	}

	return s.mappings.Delete(id)
}

// GetSKUMappings 获取映射的 SKU 映射列表
func (s *Service) GetSKUMappings(mappingID uint) ([]models.SKUMapping, error) {
	return s.skuMappings.ListByProductMapping(mappingID)
}

// GetMappedUpstreamIDs 获取指定连接下所有已映射的上游商品 ID
func (s *Service) GetMappedUpstreamIDs(connectionID uint) ([]uint, error) {
	return s.mappings.ListUpstreamIDsByConnection(connectionID)
}
