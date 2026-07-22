package service

import (
	"github.com/dujiao-next/internal/models"
	walletmodule "github.com/dujiao-next/internal/modules/wallet"
	"github.com/dujiao-next/internal/repository"
)

func (s *WalletService) queries() *walletmodule.QueryService {
	return walletmodule.NewQueryService(s.walletRepo)
}

// GetAccount is retained as a compatibility facade.
func (s *WalletService) GetAccount(userID uint) (*models.WalletAccount, error) {
	return s.queries().GetAccount(userID)
}

// ListTransactions is retained as a compatibility facade.
func (s *WalletService) ListTransactions(filter repository.WalletTransactionListFilter) ([]models.WalletTransaction, int64, error) {
	return s.queries().ListTransactions(filter)
}

// ListRechargeOrdersAdmin is retained as a compatibility facade.
func (s *WalletService) ListRechargeOrdersAdmin(filter repository.WalletRechargeListFilter) ([]models.WalletRechargeOrder, int64, error) {
	return s.queries().ListRechargeOrdersAdmin(filter)
}

// ListUserRechargeOrders is retained as a compatibility facade.
func (s *WalletService) ListUserRechargeOrders(userID uint, page, pageSize int, status, rechargeNo string) ([]models.WalletRechargeOrder, int64, error) {
	return s.queries().ListUserRechargeOrders(userID, page, pageSize, status, rechargeNo)
}

// StatsUserRechargeOrders is retained as a compatibility facade.
func (s *WalletService) StatsUserRechargeOrders(userID uint, rechargeNo string) (map[string]int64, error) {
	return s.queries().StatsUserRechargeOrders(userID, rechargeNo)
}

// GetRechargeOrderByRechargeNo is retained as a compatibility facade.
func (s *WalletService) GetRechargeOrderByRechargeNo(userID uint, rechargeNo string) (*models.WalletRechargeOrder, error) {
	return s.queries().GetRechargeOrderByRechargeNo(userID, rechargeNo)
}

// GetRechargeOrderByPaymentIDAndUser is retained as a compatibility facade.
func (s *WalletService) GetRechargeOrderByPaymentIDAndUser(paymentID uint, userID uint) (*models.WalletRechargeOrder, error) {
	return s.queries().GetRechargeOrderByPaymentIDAndUser(paymentID, userID)
}

// GetBalancesByUserIDs is retained as a compatibility facade.
func (s *WalletService) GetBalancesByUserIDs(userIDs []uint) (map[uint]models.Money, error) {
	return s.queries().GetBalancesByUserIDs(userIDs)
}
