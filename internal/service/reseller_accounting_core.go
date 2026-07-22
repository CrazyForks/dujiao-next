package service

import (
	"time"

	"github.com/dujiao-next/internal/models"
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

func (s *ResellerAccountingService) getResellerProfileByUserID(userID uint) (*models.ResellerProfile, error) {
	if s == nil || s.repo == nil || userID == 0 {
		return nil, ErrResellerNotOpened
	}
	profile, err := s.repo.GetProfileByUserID(userID)
	if err != nil {
		return nil, err
	}
	if profile == nil {
		return nil, ErrResellerNotOpened
	}
	return profile, nil
}

func requireActiveResellerProfile(profile *models.ResellerProfile) error {
	return resellermodule.RequireActiveProfile(profile)
}

func resellerWithdrawAvailability(profile *models.ResellerProfile) (bool, string) {
	return resellermodule.WithdrawAvailability(profile)
}

func (s *ResellerAccountingService) refreshBalanceAccountTx(repo repository.ResellerRepository, resellerID uint, currency string, now time.Time) error {
	return resellermodule.RefreshBalanceAccount(repo, resellerID, currency, now)
}
