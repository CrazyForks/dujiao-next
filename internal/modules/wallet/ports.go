package wallet

import (
	"time"

	"github.com/dujiao-next/internal/models"
)

// AccountListFilter filters wallet accounts.
type AccountListFilter struct {
	Page     int
	PageSize int
	UserID   uint
}

// TransactionListFilter filters wallet transactions.
type TransactionListFilter struct {
	Page        int
	PageSize    int
	UserID      uint
	OrderID     uint
	Type        string
	Direction   string
	CreatedFrom *time.Time
	CreatedTo   *time.Time
}

// RechargeListFilter filters wallet recharge orders.
type RechargeListFilter struct {
	Page         int
	PageSize     int
	RechargeNo   string
	UserID       uint
	UserKeyword  string
	PaymentID    uint
	ChannelID    uint
	ProviderType string
	ChannelType  string
	Status       string
	CreatedFrom  *time.Time
	CreatedTo    *time.Time
	PaidFrom     *time.Time
	PaidTo       *time.Time
}

// Repository supplies wallet query and account-provisioning operations.
type Repository interface {
	GetAccountByUserID(userID uint) (*models.WalletAccount, error)
	GetAccountsByUserIDs(userIDs []uint) ([]models.WalletAccount, error)
	CreateAccount(account *models.WalletAccount) error
	ListAccounts(filter AccountListFilter) ([]models.WalletAccount, int64, error)
	ListTransactions(filter TransactionListFilter) ([]models.WalletTransaction, int64, error)
	GetRechargeOrderByRechargeNo(userID uint, rechargeNo string) (*models.WalletRechargeOrder, error)
	GetRechargeOrderByPaymentIDAndUser(paymentID uint, userID uint) (*models.WalletRechargeOrder, error)
	ListRechargeOrdersAdmin(filter RechargeListFilter) ([]models.WalletRechargeOrder, int64, error)
	StatsRechargeOrders(filter RechargeListFilter) (map[string]int64, error)
}
