package repository

import (
	"time"

	"github.com/dujiao-next/internal/models"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// ResellerRepository 分销商数据访问接口。
type ResellerRepository interface {
	Transaction(fn func(tx *gorm.DB) error) error
	WithTx(tx *gorm.DB) ResellerRepository
	CreateProfile(profile *models.ResellerProfile) error
	GetProfileByID(id uint) (*models.ResellerProfile, error)
	GetProfileByUserID(userID uint) (*models.ResellerProfile, error)
	UpdateProfile(profile *models.ResellerProfile) error
	ListProfiles(filter ResellerProfileListFilter) ([]models.ResellerProfile, int64, error)
	UpsertDomain(domain models.ResellerDomain) (*models.ResellerDomain, error)
	GetDomainByID(id uint) (*models.ResellerDomain, error)
	GetDomainByIDForUpdate(id uint) (*models.ResellerDomain, error)
	UpdateDomain(domain *models.ResellerDomain) error
	FindDomainByHost(host string) (*models.ResellerDomain, error)
	FindActiveVerifiedDomain(host string) (*models.ResellerDomain, error)
	ListDomains(filter ResellerDomainListFilter) ([]models.ResellerDomain, int64, error)
	ListDomainsByResellerID(resellerID uint) ([]models.ResellerDomain, error)
	UpsertSiteConfig(config models.ResellerSiteConfig) (*models.ResellerSiteConfig, error)
	GetSiteConfigByResellerID(resellerID uint) (*models.ResellerSiteConfig, error)
	DeleteSiteConfigByResellerID(resellerID uint) error
	ListSiteConfigs(filter ResellerSiteConfigListFilter) ([]models.ResellerSiteConfig, int64, error)
	ListProductSettingsForPricing(resellerID uint, productIDs []uint, skuIDs []uint) ([]models.ResellerProductSetting, error)
	ListHiddenProductIDs(resellerID uint) ([]uint, error)
	IsActiveRelatedAccount(resellerID uint, userID uint) (bool, error)
	CreateOrderSnapshot(snapshot *models.ResellerOrderSnapshot) error
	GetOrderSnapshotByOrderID(orderID uint) (*models.ResellerOrderSnapshot, error)
	ListOrderSnapshotsByReseller(filter ResellerOrderListFilter) ([]ResellerOrderSnapshotRow, int64, error)
	StatsOrderSnapshotsByReseller(filter ResellerOrderListFilter) (ResellerOrderStatsRow, error)
	GetOrderSnapshotByResellerOrderNo(resellerID uint, orderNo string) (*ResellerOrderSnapshotRow, error)
	CreateLedgerEntryIfNotExists(entry *models.ResellerLedgerEntry) (bool, error)
	GetLedgerEntryByIdempotencyKey(key string) (*models.ResellerLedgerEntry, error)
	MarkDueLedgerEntriesAvailable(now time.Time) (int64, error)
	ListDueLedgerScopes(now time.Time) ([]ResellerLedgerScope, error)
	ListLedgerEntries(filter ResellerLedgerListFilter) ([]models.ResellerLedgerEntry, int64, error)
	SumLedgerAmount(resellerID uint, currency string, statuses []string) (decimal.Decimal, error)
	SumLedgerAmountByOrderAndType(orderID uint, ledgerType string) (decimal.Decimal, error)
	SumLedgerAmountGroupedByStatus(resellerID uint, currency string, statuses []string) (map[string]decimal.Decimal, error)
	GetOrCreateBalanceAccountForUpdate(resellerID uint, currency string) (*models.ResellerBalanceAccount, error)
	ListBalanceAccounts(filter ResellerBalanceAccountListFilter) ([]models.ResellerBalanceAccount, int64, error)
	UpdateBalanceAccount(account *models.ResellerBalanceAccount) error
	ListAvailableLedgerEntriesForUpdate(resellerID uint, currency string) ([]models.ResellerLedgerEntry, error)
	UpdateLedgerEntry(entry *models.ResellerLedgerEntry) error
	BatchUpdateLedgerEntries(ids []uint, updates map[string]interface{}) error
	BatchUpdateLedgerEntriesByWithdrawID(withdrawID uint, updates map[string]interface{}) error
	CreateWithdrawRequest(req *models.ResellerWithdrawRequest) error
	GetWithdrawRequestByID(id uint) (*models.ResellerWithdrawRequest, error)
	GetWithdrawRequestByIDForUpdate(id uint) (*models.ResellerWithdrawRequest, error)
	UpdateWithdrawRequest(req *models.ResellerWithdrawRequest) error
	ListWithdrawRequests(filter ResellerWithdrawListFilter) ([]models.ResellerWithdrawRequest, int64, error)
	ListAdminResellerLedgerEntries(filter ResellerAdminLedgerListFilter) ([]models.ResellerLedgerEntry, int64, error)
	ListAdminResellerBalanceAccounts(filter ResellerAdminBalanceAccountListFilter) ([]models.ResellerBalanceAccount, int64, error)
	ListAdminResellerWithdrawRequests(filter ResellerAdminWithdrawListFilter) ([]models.ResellerWithdrawRequest, int64, error)
}

// GormResellerRepository GORM 分销商仓储。
type GormResellerRepository struct {
	BaseRepository
}

// NewResellerRepository 创建分销商仓储。
func NewResellerRepository(db *gorm.DB) *GormResellerRepository {
	return &GormResellerRepository{BaseRepository: BaseRepository{db: db}}
}

// WithTx 绑定事务。
func (r *GormResellerRepository) WithTx(tx *gorm.DB) ResellerRepository {
	if tx == nil {
		return r
	}
	return &GormResellerRepository{BaseRepository: BaseRepository{db: tx}}
}
