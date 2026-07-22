package reseller

import (
	"github.com/dujiao-next/internal/models"
	productdomain "github.com/dujiao-next/internal/modules/catalog/product/domain"
)

// DomainLookupRepository 是租户解析所需的最小域名查询端口。
type DomainLookupRepository interface {
	FindActiveVerifiedDomain(host string) (*models.ResellerDomain, error)
}

// SiteConfigRepository 是站点配置用例所需的最小持久化端口。
type SiteConfigRepository interface {
	GetProfileByUserID(userID uint) (*models.ResellerProfile, error)
	GetProfileByID(id uint) (*models.ResellerProfile, error)
	UpsertSiteConfig(config models.ResellerSiteConfig) (*models.ResellerSiteConfig, error)
	GetSiteConfigByResellerID(resellerID uint) (*models.ResellerSiteConfig, error)
	DeleteSiteConfigByResellerID(resellerID uint) error
}

// ManagementStore 是入驻/审批/域名管理用例所需的持久化与事务端口。
// WithinTransaction 不得向用例暴露具体数据库连接类型。
type ManagementStore interface {
	WithinTransaction(fn func(store ManagementStore) error) error
	GetProfileByUserID(userID uint) (*models.ResellerProfile, error)
	GetProfileByID(id uint) (*models.ResellerProfile, error)
	CreateProfile(profile *models.ResellerProfile) error
	UpdateProfile(profile *models.ResellerProfile) error
	ListDomainsByResellerID(resellerID uint) ([]models.ResellerDomain, error)
	UpsertDomain(domain models.ResellerDomain) (*models.ResellerDomain, error)
	GetDomainByID(id uint) (*models.ResellerDomain, error)
	GetDomainByIDForUpdate(id uint) (*models.ResellerDomain, error)
	UpdateDomain(domain *models.ResellerDomain) error
	FindDomainByHost(host string) (*models.ResellerDomain, error)
}

// ProductSettingListFilter 用户侧分销商品配置列表过滤条件。
type ProductSettingListFilter struct {
	Page       int
	PageSize   int
	ResellerID uint
	CategoryID uint
	Keyword    string
	Configured string
	Listed     string
	OnlyActive bool
}

// ProductSettingAdminListFilter 后台分销商品配置列表过滤条件。
type ProductSettingAdminListFilter struct {
	Page        int
	PageSize    int
	ResellerID  uint
	UserID      uint
	ProductID   uint
	Keyword     string
	PricingMode string
	Configured  string
	Listed      string
}

// ProductSettingProductRow 商品及其分销配置。
type ProductSettingProductRow struct {
	Product  productdomain.Product
	Settings []models.ResellerProductSetting
}

// ProductSettingSummary 分销商品配置汇总。
type ProductSettingSummary struct {
	ConfiguredProducts int64
	HiddenProducts     int64
	SKUOverrides       int64
	PricingOverrides   int64
}

// ProductSettingStore 是分销商品配置用例所需的持久化与事务端口。
// WithinTransaction 不得向用例暴露具体数据库连接类型。
type ProductSettingStore interface {
	WithinTransaction(fn func(store ProductSettingStore) error) error
	GetProfileByUserID(userID uint) (*models.ResellerProfile, error)
	GetProfileByID(id uint) (*models.ResellerProfile, error)
	ListProductsWithSettings(filter ProductSettingListFilter) ([]ProductSettingProductRow, int64, error)
	GetProductWithSettings(resellerID, productID uint) (*ProductSettingProductRow, error)
	UpsertSetting(setting models.ResellerProductSetting) (*models.ResellerProductSetting, error)
	DeleteSetting(resellerID, productID, skuID uint) error
	ListAdminSettings(filter ProductSettingAdminListFilter) ([]models.ResellerProductSetting, int64, error)
	SummarizeByResellerID(resellerID uint) (ProductSettingSummary, error)
}
