package application

import (
	"fmt"
	"strings"
	"time"

	walletcontract "github.com/dujiao-next/internal/modules/wallet/contract"
)

const defaultCurrency = "CNY"

type Options struct {
	Repository   walletcontract.Repository
	Transactions walletcontract.UnitOfWork
}

type Service struct {
	repository   walletcontract.Repository
	transactions walletcontract.UnitOfWork
}

var _ walletcontract.UseCase = (*Service)(nil)

func NewService(options Options) *Service {
	return &Service{repository: options.Repository, transactions: options.Transactions}
}

func normalizeCurrency(currency string) string {
	normalized := strings.ToUpper(strings.TrimSpace(currency))
	if normalized == "" {
		return defaultCurrency
	}
	return normalized
}

func cleanRemark(raw, fallback string) string {
	remark := strings.TrimSpace(raw)
	if remark == "" {
		return fallback
	}
	return remark
}

func orderReference(orderID uint, action string) string {
	action = strings.TrimSpace(action)
	if action == "" {
		action = "wallet"
	}
	return fmt.Sprintf("order:%d:%s", orderID, action)
}

func uniqueReference(prefix string, id uint) string {
	normalized := strings.TrimSpace(prefix)
	if normalized == "" {
		normalized = "wallet"
	}
	return fmt.Sprintf("%s:%d:%d", normalized, id, time.Now().UnixNano())
}
