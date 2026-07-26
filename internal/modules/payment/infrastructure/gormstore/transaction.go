package gormstore

import (
	ordercontract "github.com/dujiao-next/internal/modules/order/contract"
	ordergormstore "github.com/dujiao-next/internal/modules/order/infrastructure/gormstore"
	paymentcontract "github.com/dujiao-next/internal/modules/payment/contract"

	"gorm.io/gorm"
)

type transaction struct {
	ordercontract.Transaction
	db *gorm.DB
}

var _ paymentcontract.Transaction = transaction{}

// UseTransaction 将调用方已打开的数据库事务适配为支付工作单元。
func UseTransaction(tx *gorm.DB) paymentcontract.Transaction {
	if tx == nil {
		return nil
	}
	return transaction{
		Transaction: ordergormstore.UseTransaction(tx),
		db:          tx,
	}
}

func (tx transaction) Payments() paymentcontract.Store {
	return New(tx.db)
}

func (tx transaction) PaymentChannels() paymentcontract.ChannelStore {
	return NewChannelStore(tx.db)
}

func (s *Store) WithinTransaction(fn func(paymentcontract.Transaction) error) error {
	if s == nil || s.db == nil || fn == nil {
		return nil
	}
	return s.db.Transaction(func(tx *gorm.DB) error {
		return fn(UseTransaction(tx))
	})
}
