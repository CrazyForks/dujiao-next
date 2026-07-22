package reseller

import (
	"time"

	"github.com/dujiao-next/internal/models"
	"github.com/shopspring/decimal"
)

const (
	WithdrawDisabledReasonProfileInactive       = "profile_inactive"
	WithdrawDisabledReasonSettlementUnavailable = "settlement_unavailable"
)

// AdminLedgerListFilter 管理端账务流水列表过滤条件。
type AdminLedgerListFilter struct {
	Page        int
	PageSize    int
	ResellerID  uint
	UserID      uint
	Keyword     string
	Currency    string
	Type        string
	Status      string
	OrderID     uint
	OrderNo     string
	CreatedFrom *time.Time
	CreatedTo   *time.Time
}

// AdminBalanceAccountListFilter 管理端余额账户列表过滤条件。
type AdminBalanceAccountListFilter struct {
	Page       int
	PageSize   int
	ResellerID uint
	UserID     uint
	Keyword    string
	Currency   string
	Status     string
}

// AdminWithdrawListFilter 管理端提现申请列表过滤条件。
type AdminWithdrawListFilter struct {
	Page        int
	PageSize    int
	ResellerID  uint
	UserID      uint
	Keyword     string
	Currency    string
	Status      string
	CreatedFrom *time.Time
	CreatedTo   *time.Time
}

// UserFinanceDashboard 用户侧分销财务看板。
type UserFinanceDashboard struct {
	Opened                 bool
	Profile                *models.ResellerProfile
	Balances               []models.ResellerBalanceAccount
	WithdrawEnabled        bool
	WithdrawDisabledReason string
}

// UserLedgerListFilter 用户侧账务流水列表过滤条件。
type UserLedgerListFilter struct {
	Page     int
	PageSize int
	Currency string
	Type     string
	Status   string
	OrderID  uint
}

// UserBalanceAccountListFilter 用户侧余额账户列表过滤条件。
type UserBalanceAccountListFilter struct {
	Page     int
	PageSize int
	Currency string
	Status   string
}

// UserWithdrawListFilter 用户侧提现申请列表过滤条件。
type UserWithdrawListFilter struct {
	Page     int
	PageSize int
	Currency string
	Status   string
}

// WithdrawApplyInput 分销提现申请输入。
type WithdrawApplyInput struct {
	Amount   decimal.Decimal
	Currency string
	Channel  string
	Account  string
}

// LedgerListFilter 分销商账务流水持久化过滤条件。
type LedgerListFilter struct {
	Page       int
	PageSize   int
	ResellerID uint
	Currency   string
	Type       string
	Status     string
	OrderID    uint
}

// BalanceAccountListFilter 分销商余额账户持久化过滤条件。
type BalanceAccountListFilter struct {
	Page       int
	PageSize   int
	ResellerID uint
	Currency   string
	Status     string
}

// WithdrawListFilter 分销商提现申请持久化过滤条件。
type WithdrawListFilter struct {
	Page       int
	PageSize   int
	ResellerID uint
	Currency   string
	Status     string
}

// AccountingQueryStore 是分销财务只读查询用例所需的最小持久化端口。
type AccountingQueryStore interface {
	GetProfileByUserID(userID uint) (*models.ResellerProfile, error)
	ListBalanceAccounts(filter BalanceAccountListFilter) ([]models.ResellerBalanceAccount, int64, error)
	ListLedgerEntries(filter LedgerListFilter) ([]models.ResellerLedgerEntry, int64, error)
	ListWithdrawRequests(filter WithdrawListFilter) ([]models.ResellerWithdrawRequest, int64, error)
	ListAdminResellerLedgerEntries(filter AdminLedgerListFilter) ([]models.ResellerLedgerEntry, int64, error)
	ListAdminResellerBalanceAccounts(filter AdminBalanceAccountListFilter) ([]models.ResellerBalanceAccount, int64, error)
	ListAdminResellerWithdrawRequests(filter AdminWithdrawListFilter) ([]models.ResellerWithdrawRequest, int64, error)
}

const (
	WithdrawActionReject = "reject"
	WithdrawActionPay    = "pay"
)

// BalanceAccountStore 是余额缓存刷新所需的最小持久化端口。
type BalanceAccountStore interface {
	GetOrCreateBalanceAccountForUpdate(resellerID uint, currency string) (*models.ResellerBalanceAccount, error)
	SumLedgerAmountGroupedByStatus(resellerID uint, currency string, statuses []string) (map[string]decimal.Decimal, error)
	UpdateBalanceAccount(account *models.ResellerBalanceAccount) error
}

// AccountingWithdrawStore 是分销提现用例所需的持久化与事务端口。
// WithinTransaction 不得向用例暴露具体数据库连接类型。
type AccountingWithdrawStore interface {
	BalanceAccountStore
	WithinTransaction(fn func(store AccountingWithdrawStore) error) error
	GetProfileByUserID(userID uint) (*models.ResellerProfile, error)
	ListAvailableLedgerEntriesForUpdate(resellerID uint, currency string) ([]models.ResellerLedgerEntry, error)
	UpdateLedgerEntry(entry *models.ResellerLedgerEntry) error
	CreateLedgerEntryIfNotExists(entry *models.ResellerLedgerEntry) (bool, error)
	BatchUpdateLedgerEntries(ids []uint, updates map[string]interface{}) error
	BatchUpdateLedgerEntriesByWithdrawID(withdrawID uint, updates map[string]interface{}) error
	CreateWithdrawRequest(req *models.ResellerWithdrawRequest) error
	GetWithdrawRequestByID(id uint) (*models.ResellerWithdrawRequest, error)
	GetWithdrawRequestByIDForUpdate(id uint) (*models.ResellerWithdrawRequest, error)
	UpdateWithdrawRequest(req *models.ResellerWithdrawRequest) error
}

// LedgerScope 表示分销商 + 币种的账户维度。
type LedgerScope struct {
	ResellerID uint
	Currency   string
}

// AccountingLedgerStore 是利润入账/退款扣减/到期确认所需的持久化与事务端口。
// WithinTransaction 不得向用例暴露具体数据库连接类型。
type AccountingLedgerStore interface {
	BalanceAccountStore
	WithinTransaction(fn func(store AccountingLedgerStore) error) error
	GetOrderSnapshotByOrderID(orderID uint) (*models.ResellerOrderSnapshot, error)
	CreateLedgerEntryIfNotExists(entry *models.ResellerLedgerEntry) (bool, error)
	ListDueLedgerScopes(now time.Time) ([]LedgerScope, error)
	MarkDueLedgerEntriesAvailable(now time.Time) (int64, error)
	SumLedgerAmountByOrderAndType(orderID uint, ledgerType string) (decimal.Decimal, error)
	GetLedgerEntryByIdempotencyKey(key string) (*models.ResellerLedgerEntry, error)
}
