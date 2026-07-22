package repository

import (
	"time"

	resellermodule "github.com/dujiao-next/internal/modules/reseller"
	walletmodule "github.com/dujiao-next/internal/modules/wallet"
)

// Pagination 通用分页参数
type Pagination struct {
	Page     int
	PageSize int
}

// OrderListFilter 查询订单列表的过滤条件
type OrderListFilter struct {
	Page           int
	PageSize       int
	UserID         uint
	UserKeyword    string
	Status         string
	OrderNo        string
	GuestEmail     string
	ProductKeyword string
	CreatedFrom    *time.Time
	CreatedTo      *time.Time
	SortBy         string
	SortOrder      string
}

// ResellerOrderScope 表示前台订单查询的分销租户范围。
//
// ResellerID == nil 明确表示主站范围: orders.reseller_id IS NULL。
// 后台列表不要使用该结构，后台 nil 语义是“不按分销商过滤”。
type ResellerOrderScope struct {
	ResellerID *uint
}

// ResellerLedgerListFilter 分销商账务流水过滤条件。
type ResellerLedgerListFilter = resellermodule.LedgerListFilter

// ResellerOrderListFilter 分销商视角销售订单过滤条件。
type ResellerOrderListFilter = resellermodule.OrderSnapshotListFilter

// ResellerOrderSnapshotRow 聚合订单快照、订单展示字段、商品行和账务流水。
type ResellerOrderSnapshotRow = resellermodule.OrderSnapshotRow

// ResellerOrderStatsRow 分销商视角销售订单统计。
type ResellerOrderStatsRow = resellermodule.OrderStatsRow

// ResellerAdminLedgerListFilter 管理端分销商账务流水过滤条件。
type ResellerAdminLedgerListFilter = resellermodule.AdminLedgerListFilter

// ResellerAdminBalanceAccountListFilter 管理端分销商余额账户过滤条件。
type ResellerAdminBalanceAccountListFilter = resellermodule.AdminBalanceAccountListFilter

// ResellerBalanceAccountListFilter 分销商余额账户过滤条件。
type ResellerBalanceAccountListFilter = resellermodule.BalanceAccountListFilter

// ResellerWithdrawListFilter 分销商提现申请过滤条件。
type ResellerWithdrawListFilter = resellermodule.WithdrawListFilter

// ResellerAdminWithdrawListFilter 管理端分销商提现过滤条件。
type ResellerAdminWithdrawListFilter = resellermodule.AdminWithdrawListFilter

// ResellerProfileListFilter 管理端分销商资料过滤条件。
type ResellerProfileListFilter struct {
	Page             int
	PageSize         int
	UserID           uint
	Status           string
	SettlementStatus string
	Keyword          string
	CreatedFrom      *time.Time
	CreatedTo        *time.Time
}

// ResellerDomainListFilter 管理端分销商域名过滤条件。
type ResellerDomainListFilter struct {
	Page               int
	PageSize           int
	ResellerID         uint
	UserID             uint
	Domain             string
	Type               string
	Status             string
	VerificationStatus string
	Keyword            string
	CreatedFrom        *time.Time
	CreatedTo          *time.Time
}

// ResellerSiteConfigListFilter 分销站点配置列表过滤条件。
type ResellerSiteConfigListFilter struct {
	Page        int
	PageSize    int
	ResellerID  uint
	Keyword     string
	CreatedFrom *time.Time
	CreatedTo   *time.Time
}

// PaymentListFilter 查询支付列表的过滤条件
type PaymentListFilter struct {
	Page         int
	PageSize     int
	UserID       uint
	OrderID      uint
	ChannelID    uint
	ProviderType string
	ChannelType  string
	Status       string
	CreatedFrom  *time.Time
	CreatedTo    *time.Time
	SkipCount    bool
	Lightweight  bool
}

// OrderRefundRecordListFilter 查询订单退款记录列表的过滤条件
type OrderRefundRecordListFilter struct {
	Page           int
	PageSize       int
	UserID         uint
	UserKeyword    string
	OrderNo        string
	GuestEmail     string
	ProductKeyword string
	CreatedFrom    *time.Time
	CreatedTo      *time.Time
}

// PaymentChannelListFilter 查询支付渠道列表的过滤条件
type PaymentChannelListFilter struct {
	Page         int
	PageSize     int
	ProviderType string
	ChannelType  string
	ActiveOnly   bool
}

// WalletAccountListFilter is retained for legacy callers.
type WalletAccountListFilter = walletmodule.AccountListFilter

// WalletTransactionListFilter is retained for legacy callers.
type WalletTransactionListFilter = walletmodule.TransactionListFilter

// WalletRechargeListFilter is retained for legacy callers.
type WalletRechargeListFilter = walletmodule.RechargeListFilter
