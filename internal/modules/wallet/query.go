package wallet

import (
	"time"

	"github.com/dujiao-next/internal/models"
	"github.com/dujiao-next/internal/shared/money"
	"github.com/shopspring/decimal"
)

// QueryService provides the wallet read use-cases.
type QueryService struct {
	repo Repository
}

func NewQueryService(repo Repository) *QueryService {
	return &QueryService{repo: repo}
}

// GetAccount returns an account, provisioning it when absent.
func (s *QueryService) GetAccount(userID uint) (*models.WalletAccount, error) {
	if userID == 0 {
		return nil, ErrAccountNotFound
	}
	account, err := s.repo.GetAccountByUserID(userID)
	if err != nil || account != nil {
		return account, err
	}
	now := time.Now()
	account = &models.WalletAccount{
		UserID: userID, Balance: money.FromDecimal(decimal.Zero), CreatedAt: now, UpdatedAt: now,
	}
	if err := s.repo.CreateAccount(account); err != nil {
		created, queryErr := s.repo.GetAccountByUserID(userID)
		if queryErr == nil && created != nil {
			return created, nil
		}
		return nil, ErrAccountCreateFailed
	}
	return account, nil
}

func (s *QueryService) ListTransactions(filter TransactionListFilter) ([]models.WalletTransaction, int64, error) {
	return s.repo.ListTransactions(filter)
}

func (s *QueryService) ListRechargeOrdersAdmin(filter RechargeListFilter) ([]models.WalletRechargeOrder, int64, error) {
	return s.repo.ListRechargeOrdersAdmin(filter)
}

func (s *QueryService) ListUserRechargeOrders(userID uint, page, pageSize int, status, rechargeNo string) ([]models.WalletRechargeOrder, int64, error) {
	if userID == 0 {
		return nil, 0, ErrAccountNotFound
	}
	return s.repo.ListRechargeOrdersAdmin(RechargeListFilter{Page: page, PageSize: pageSize, UserID: userID, Status: status, RechargeNo: rechargeNo})
}

func (s *QueryService) StatsUserRechargeOrders(userID uint, rechargeNo string) (map[string]int64, error) {
	if userID == 0 {
		return nil, ErrAccountNotFound
	}
	return s.repo.StatsRechargeOrders(RechargeListFilter{UserID: userID, RechargeNo: rechargeNo})
}

func (s *QueryService) GetRechargeOrderByRechargeNo(userID uint, rechargeNo string) (*models.WalletRechargeOrder, error) {
	if userID == 0 {
		return nil, ErrRechargeNotFound
	}
	order, err := s.repo.GetRechargeOrderByRechargeNo(userID, rechargeNo)
	if err != nil || order != nil {
		return order, err
	}
	return nil, ErrRechargeNotFound
}

func (s *QueryService) GetRechargeOrderByPaymentIDAndUser(paymentID, userID uint) (*models.WalletRechargeOrder, error) {
	if paymentID == 0 || userID == 0 {
		return nil, ErrRechargeNotFound
	}
	order, err := s.repo.GetRechargeOrderByPaymentIDAndUser(paymentID, userID)
	if err != nil || order != nil {
		return order, err
	}
	return nil, ErrRechargeNotFound
}

func (s *QueryService) GetBalancesByUserIDs(userIDs []uint) (map[uint]money.Amount, error) {
	result := make(map[uint]money.Amount, len(userIDs))
	if len(userIDs) == 0 {
		return result, nil
	}
	accounts, err := s.repo.GetAccountsByUserIDs(userIDs)
	if err != nil {
		return nil, err
	}
	for _, account := range accounts {
		result[account.UserID] = account.Balance
	}
	return result, nil
}
