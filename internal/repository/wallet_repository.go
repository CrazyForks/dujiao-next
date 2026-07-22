package repository

import (
	"github.com/dujiao-next/internal/models"
	walletmodule "github.com/dujiao-next/internal/modules/wallet"
	walletgormstore "github.com/dujiao-next/internal/modules/wallet/store/gormstore"
	"gorm.io/gorm"
)

// WalletRepository is a compatibility port for legacy wallet write paths.
// New query consumers should depend on wallet.Repository.
type WalletRepository interface {
	walletmodule.Repository
	GetAccountByUserIDForUpdate(userID uint) (*models.WalletAccount, error)
	UpdateAccount(account *models.WalletAccount) error
	CreateTransaction(txn *models.WalletTransaction) error
	GetTransactionByReference(reference string) (*models.WalletTransaction, error)
	CreateRechargeOrder(order *models.WalletRechargeOrder) error
	UpdateRechargeOrder(order *models.WalletRechargeOrder) error
	GetRechargeOrderByPaymentID(paymentID uint) (*models.WalletRechargeOrder, error)
	GetRechargeOrderByPaymentIDForUpdate(paymentID uint) (*models.WalletRechargeOrder, error)
	GetRechargeOrdersByPaymentIDs(paymentIDs []uint) ([]models.WalletRechargeOrder, error)
	Transaction(fn func(tx *gorm.DB) error) error
	WithTx(tx *gorm.DB) *GormWalletRepository
}

// GormWalletRepository adapts the wallet GORM store for remaining legacy writers.
type GormWalletRepository struct {
	*walletgormstore.Store
}

func NewWalletRepository(db *gorm.DB) *GormWalletRepository {
	return AdaptWalletStore(walletgormstore.New(db))
}

func AdaptWalletStore(store *walletgormstore.Store) *GormWalletRepository {
	return &GormWalletRepository{Store: store}
}

func (r *GormWalletRepository) WithTx(tx *gorm.DB) *GormWalletRepository {
	if tx == nil {
		return r
	}
	return AdaptWalletStore(r.Store.WithTx(tx))
}
