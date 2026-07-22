package reseller

import (
	"errors"
	"strings"

	"github.com/dujiao-next/internal/models"
)

// AccountingQueryService 分销财务只读查询用例。
type AccountingQueryService struct {
	store AccountingQueryStore
}

func NewAccountingQueryService(store AccountingQueryStore) *AccountingQueryService {
	return &AccountingQueryService{store: store}
}

// RequireActiveProfile 校验分销资料处于可结算的激活状态。
func RequireActiveProfile(profile *models.ResellerProfile) error {
	if profile == nil {
		return ErrNotOpened
	}
	if profile.Status != models.ResellerProfileStatusActive {
		return ErrProfileInactive
	}
	if profile.SettlementStatus != "" && profile.SettlementStatus != models.ResellerSettlementStatusNormal {
		return ErrSettlementUnavailable
	}
	return nil
}

// WithdrawAvailability 返回当前资料是否允许提现及禁用原因。
func WithdrawAvailability(profile *models.ResellerProfile) (bool, string) {
	if profile == nil {
		return false, ""
	}
	if profile.Status != models.ResellerProfileStatusActive {
		return false, WithdrawDisabledReasonProfileInactive
	}
	if profile.SettlementStatus != "" && profile.SettlementStatus != models.ResellerSettlementStatusNormal {
		return false, WithdrawDisabledReasonSettlementUnavailable
	}
	return true, ""
}

func (s *AccountingQueryService) getProfileByUserID(userID uint) (*models.ResellerProfile, error) {
	if s == nil || s.store == nil || userID == 0 {
		return nil, ErrNotOpened
	}
	profile, err := s.store.GetProfileByUserID(userID)
	if err != nil {
		return nil, err
	}
	if profile == nil {
		return nil, ErrNotOpened
	}
	return profile, nil
}

func (s *AccountingQueryService) GetUserFinanceDashboard(userID uint) (UserFinanceDashboard, error) {
	profile, err := s.getProfileByUserID(userID)
	if errors.Is(err, ErrNotOpened) {
		return UserFinanceDashboard{Opened: false}, nil
	}
	if err != nil {
		return UserFinanceDashboard{}, err
	}
	balances, _, err := s.store.ListBalanceAccounts(BalanceAccountListFilter{
		Page:       1,
		PageSize:   100,
		ResellerID: profile.ID,
	})
	if err != nil {
		return UserFinanceDashboard{}, err
	}
	withdrawEnabled, withdrawDisabledReason := WithdrawAvailability(profile)
	return UserFinanceDashboard{
		Opened:                 true,
		Profile:                profile,
		Balances:               balances,
		WithdrawEnabled:        withdrawEnabled,
		WithdrawDisabledReason: withdrawDisabledReason,
	}, nil
}

func (s *AccountingQueryService) ListUserBalanceAccounts(userID uint, filter UserBalanceAccountListFilter) ([]models.ResellerBalanceAccount, int64, error) {
	profile, err := s.getProfileByUserID(userID)
	if err != nil {
		return nil, 0, err
	}
	if err := RequireActiveProfile(profile); err != nil {
		return nil, 0, err
	}
	return s.store.ListBalanceAccounts(BalanceAccountListFilter{
		Page:       filter.Page,
		PageSize:   filter.PageSize,
		ResellerID: profile.ID,
		Currency:   strings.TrimSpace(filter.Currency),
		Status:     strings.TrimSpace(filter.Status),
	})
}

func (s *AccountingQueryService) ListUserLedgerEntries(userID uint, filter UserLedgerListFilter) ([]models.ResellerLedgerEntry, int64, error) {
	profile, err := s.getProfileByUserID(userID)
	if err != nil {
		return nil, 0, err
	}
	if err := RequireActiveProfile(profile); err != nil {
		return nil, 0, err
	}
	return s.store.ListLedgerEntries(LedgerListFilter{
		Page:       filter.Page,
		PageSize:   filter.PageSize,
		ResellerID: profile.ID,
		Currency:   strings.TrimSpace(filter.Currency),
		Type:       strings.TrimSpace(filter.Type),
		Status:     strings.TrimSpace(filter.Status),
		OrderID:    filter.OrderID,
	})
}

func (s *AccountingQueryService) ListUserWithdrawRequests(userID uint, filter UserWithdrawListFilter) ([]models.ResellerWithdrawRequest, int64, error) {
	profile, err := s.getProfileByUserID(userID)
	if err != nil {
		return nil, 0, err
	}
	if err := RequireActiveProfile(profile); err != nil {
		return nil, 0, err
	}
	return s.store.ListWithdrawRequests(WithdrawListFilter{
		Page:       filter.Page,
		PageSize:   filter.PageSize,
		ResellerID: profile.ID,
		Currency:   strings.TrimSpace(filter.Currency),
		Status:     strings.TrimSpace(filter.Status),
	})
}

func (s *AccountingQueryService) ListAdminLedgerEntries(filter AdminLedgerListFilter) ([]models.ResellerLedgerEntry, int64, error) {
	if s == nil || s.store == nil {
		return []models.ResellerLedgerEntry{}, 0, nil
	}
	return s.store.ListAdminResellerLedgerEntries(AdminLedgerListFilter{
		Page:        filter.Page,
		PageSize:    filter.PageSize,
		ResellerID:  filter.ResellerID,
		UserID:      filter.UserID,
		Keyword:     strings.TrimSpace(filter.Keyword),
		Currency:    strings.TrimSpace(filter.Currency),
		Type:        strings.TrimSpace(filter.Type),
		Status:      strings.TrimSpace(filter.Status),
		OrderID:     filter.OrderID,
		OrderNo:     strings.TrimSpace(filter.OrderNo),
		CreatedFrom: filter.CreatedFrom,
		CreatedTo:   filter.CreatedTo,
	})
}

func (s *AccountingQueryService) ListAdminBalanceAccounts(filter AdminBalanceAccountListFilter) ([]models.ResellerBalanceAccount, int64, error) {
	if s == nil || s.store == nil {
		return []models.ResellerBalanceAccount{}, 0, nil
	}
	return s.store.ListAdminResellerBalanceAccounts(AdminBalanceAccountListFilter{
		Page:       filter.Page,
		PageSize:   filter.PageSize,
		ResellerID: filter.ResellerID,
		UserID:     filter.UserID,
		Keyword:    strings.TrimSpace(filter.Keyword),
		Currency:   strings.TrimSpace(filter.Currency),
		Status:     strings.TrimSpace(filter.Status),
	})
}

func (s *AccountingQueryService) ListAdminWithdrawRequests(filter AdminWithdrawListFilter) ([]models.ResellerWithdrawRequest, int64, error) {
	if s == nil || s.store == nil {
		return []models.ResellerWithdrawRequest{}, 0, nil
	}
	return s.store.ListAdminResellerWithdrawRequests(AdminWithdrawListFilter{
		Page:        filter.Page,
		PageSize:    filter.PageSize,
		ResellerID:  filter.ResellerID,
		UserID:      filter.UserID,
		Keyword:     strings.TrimSpace(filter.Keyword),
		Currency:    strings.TrimSpace(filter.Currency),
		Status:      strings.TrimSpace(filter.Status),
		CreatedFrom: filter.CreatedFrom,
		CreatedTo:   filter.CreatedTo,
	})
}
