package service

import (
	resellermodule "github.com/dujiao-next/internal/modules/reseller"
	"github.com/dujiao-next/internal/repository"
)

const (
	resellerWithdrawActionReject = resellermodule.WithdrawActionReject
	resellerWithdrawActionPay    = resellermodule.WithdrawActionPay

	ResellerWithdrawDisabledReasonProfileInactive       = resellermodule.WithdrawDisabledReasonProfileInactive
	ResellerWithdrawDisabledReasonSettlementUnavailable = resellermodule.WithdrawDisabledReasonSettlementUnavailable
)

type ResellerAccountingOptions struct {
	ConfirmDays int
}

type ResellerAccountingService struct {
	repo        repository.ResellerRepository
	query       *resellermodule.AccountingQueryService
	withdraw    *resellermodule.AccountingWithdrawService
	ledger      *resellermodule.AccountingLedgerService
	confirmDays int
}

func NewResellerAccountingService(repo repository.ResellerRepository, opts ResellerAccountingOptions) *ResellerAccountingService {
	const maxConfirmDays = 3650
	days := opts.ConfirmDays
	if days < 0 {
		days = 0
	}
	if days > maxConfirmDays {
		days = maxConfirmDays
	}
	withdrawStore := newResellerAccountingWithdrawStoreAdapter(repo)
	ledgerStore := newResellerAccountingLedgerStoreAdapter(repo)
	return &ResellerAccountingService{
		repo:        repo,
		query:       resellermodule.NewAccountingQueryService(repo),
		withdraw:    resellermodule.NewAccountingWithdrawService(withdrawStore),
		ledger:      resellermodule.NewAccountingLedgerService(ledgerStore, days),
		confirmDays: days,
	}
}
