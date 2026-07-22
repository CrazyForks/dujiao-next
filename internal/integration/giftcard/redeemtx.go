package giftcardintegration

import (
	"github.com/dujiao-next/internal/models"
	giftcardcontract "github.com/dujiao-next/internal/modules/giftcard/contract"
	giftcarddomain "github.com/dujiao-next/internal/modules/giftcard/domain"
	"github.com/dujiao-next/internal/modules/giftcard/infrastructure/gormstore"
	"github.com/dujiao-next/internal/service"

	"gorm.io/gorm"
)

// Runner 将礼品卡仓储锁与钱包入账绑定到同一个 GORM 事务。
type Runner struct {
	cards  *gormstore.Store
	wallet *service.WalletService
}

func New(cards *gormstore.Store, wallet *service.WalletService) *Runner {
	return &Runner{cards: cards, wallet: wallet}
}

func (r *Runner) WithinRedeemTransaction(fn func(tx giftcardcontract.RedeemTransaction) error) error {
	if r == nil || r.cards == nil || r.wallet == nil {
		return giftcardcontract.ErrFetchFailed
	}
	return r.cards.Transaction(func(tx *gorm.DB) error {
		return fn(&transaction{
			db:     tx,
			cards:  r.cards.WithTx(tx),
			wallet: r.wallet,
		})
	})
}

type transaction struct {
	db     *gorm.DB
	cards  *gormstore.Store
	wallet *service.WalletService
}

func (tx *transaction) GetByCodeForUpdate(code string) (*giftcarddomain.GiftCard, error) {
	return tx.cards.GetByCodeForUpdate(code)
}

func (tx *transaction) UpdateCard(card *giftcarddomain.GiftCard) error {
	return tx.cards.Update(card)
}

func (tx *transaction) CreditWallet(input giftcardcontract.WalletCreditInput) (*models.WalletAccount, *models.WalletTransaction, error) {
	return tx.wallet.CreditInTx(tx.db, service.WalletCreditInput{
		UserID:    input.UserID,
		Amount:    input.Amount,
		Currency:  input.Currency,
		TxnType:   input.TxnType,
		Reference: input.Reference,
		Remark:    input.Remark,
		OrderID:   input.OrderID,
	})
}

var _ giftcardcontract.RedeemTransactionRunner = (*Runner)(nil)
var _ giftcardcontract.RedeemTransaction = (*transaction)(nil)
