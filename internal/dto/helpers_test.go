package dto

import (
	"github.com/dujiao-next/internal/shared/money"
	"github.com/shopspring/decimal"
)

func newMoney(s string) money.Amount {
	return money.FromDecimal(decimal.RequireFromString(s))
}
