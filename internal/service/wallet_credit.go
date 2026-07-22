package service

import (
	"strings"
	"time"

	"github.com/dujiao-next/internal/constants"
	"github.com/dujiao-next/internal/models"
	"github.com/dujiao-next/internal/repository"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// WalletCreditInput 事务内入账输入
type WalletCreditInput struct {
	UserID    uint
	Amount    models.Money
	Currency  string
	TxnType   string
	Reference string
	Remark    string
	OrderID   *uint
}

// CreditInTx 在事务内执行钱包入账并写入唯一参考号流水
func (s *WalletService) CreditInTx(tx *gorm.DB, input WalletCreditInput) (*models.WalletAccount, *models.WalletTransaction, error) {
	if tx == nil {
		return nil, nil, ErrOrderUpdateFailed
	}
	if input.UserID == 0 {
		return nil, nil, ErrWalletAccountNotFound
	}
	amount := input.Amount.Decimal.Round(2)
	if amount.LessThanOrEqual(decimal.Zero) {
		return nil, nil, ErrWalletInvalidAmount
	}
	reference := strings.TrimSpace(input.Reference)
	if reference == "" {
		return nil, nil, ErrWalletTransactionCreateFailed
	}
	txnType := strings.TrimSpace(input.TxnType)
	if txnType == "" {
		txnType = constants.WalletTxnTypeRecharge
	}
	remark := cleanWalletRemark(input.Remark, "钱包入账")
	now := time.Now()
	repo := s.walletRepo.WithTx(tx)

	exists, err := repo.GetTransactionByReference(reference)
	if err != nil {
		return nil, nil, err
	}
	if exists != nil {
		account, accountErr := repo.GetAccountByUserID(input.UserID)
		if accountErr != nil {
			return nil, nil, accountErr
		}
		if account == nil {
			account, accountErr = s.ensureAccountForUpdate(repo, input.UserID, now)
			if accountErr != nil {
				return nil, nil, accountErr
			}
		}
		return account, exists, nil
	}

	account, err := s.ensureAccountForUpdate(repo, input.UserID, now)
	if err != nil {
		return nil, nil, err
	}
	before := account.Balance.Decimal.Round(2)
	after := before.Add(amount).Round(2)
	account.Balance = models.NewMoneyFromDecimal(after)
	account.UpdatedAt = now
	if err := repo.UpdateAccount(account); err != nil {
		return nil, nil, ErrWalletAccountUpdateFailed
	}

	txn := &models.WalletTransaction{
		UserID:        input.UserID,
		OrderID:       input.OrderID,
		Type:          txnType,
		Direction:     constants.WalletTxnDirectionIn,
		Amount:        models.NewMoneyFromDecimal(amount),
		BalanceBefore: models.NewMoneyFromDecimal(before),
		BalanceAfter:  models.NewMoneyFromDecimal(after),
		Currency:      normalizeWalletCurrency(input.Currency),
		Reference:     reference,
		Remark:        remark,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	if err := repo.CreateTransaction(txn); err != nil {
		return nil, nil, ErrWalletTransactionCreateFailed
	}
	return account, txn, nil
}

func (s *WalletService) changeBalance(userID uint, delta decimal.Decimal, txnType string, orderID *uint, reference, remark, currency string) (*models.WalletAccount, *models.WalletTransaction, error) {
	var accountResult *models.WalletAccount
	var txnResult *models.WalletTransaction
	if err := s.walletRepo.Transaction(func(tx *gorm.DB) error {
		repo := s.walletRepo.WithTx(tx)
		now := time.Now()
		account, err := s.ensureAccountForUpdate(repo, userID, now)
		if err != nil {
			return err
		}

		before := account.Balance.Decimal.Round(2)
		after := before.Add(delta).Round(2)
		if after.LessThan(decimal.Zero) {
			return ErrWalletInsufficientBalance
		}
		direction := constants.WalletTxnDirectionIn
		amount := delta.Round(2)
		if delta.LessThan(decimal.Zero) {
			direction = constants.WalletTxnDirectionOut
			amount = delta.Abs().Round(2)
		}

		account.Balance = models.NewMoneyFromDecimal(after)
		account.UpdatedAt = now
		if err := repo.UpdateAccount(account); err != nil {
			return ErrWalletAccountUpdateFailed
		}

		txn := &models.WalletTransaction{
			UserID:        userID,
			OrderID:       orderID,
			Type:          txnType,
			Direction:     direction,
			Amount:        models.NewMoneyFromDecimal(amount),
			BalanceBefore: models.NewMoneyFromDecimal(before),
			BalanceAfter:  models.NewMoneyFromDecimal(after),
			Currency:      normalizeWalletCurrency(currency),
			Reference:     strings.TrimSpace(reference),
			Remark:        remark,
			CreatedAt:     now,
			UpdatedAt:     now,
		}
		if err := repo.CreateTransaction(txn); err != nil {
			return ErrWalletTransactionCreateFailed
		}

		accountResult = account
		txnResult = txn
		return nil
	}); err != nil {
		return nil, nil, err
	}
	return accountResult, txnResult, nil
}

func (s *WalletService) ensureAccountForUpdate(repo *repository.GormWalletRepository, userID uint, now time.Time) (*models.WalletAccount, error) {
	account, err := repo.GetAccountByUserIDForUpdate(userID)
	if err != nil {
		return nil, err
	}
	if account != nil {
		return account, nil
	}
	account = &models.WalletAccount{
		UserID:    userID,
		Balance:   models.NewMoneyFromDecimal(decimal.Zero),
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := repo.CreateAccount(account); err != nil {
		created, queryErr := repo.GetAccountByUserIDForUpdate(userID)
		if queryErr == nil && created != nil {
			return created, nil
		}
		return nil, ErrWalletAccountCreateFailed
	}
	return account, nil
}
