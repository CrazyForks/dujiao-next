package reseller

import (
	"fmt"
	"strings"
	"time"

	productcontract "github.com/dujiao-next/internal/modules/catalog/product/contract"

	"github.com/dujiao-next/internal/models"
	"github.com/dujiao-next/internal/shared/money"
	"github.com/shopspring/decimal"
)

// AccountingWithdrawService 分销提现申请与审核用例。
type AccountingWithdrawService struct {
	store AccountingWithdrawStore
}

func NewAccountingWithdrawService(store AccountingWithdrawStore) *AccountingWithdrawService {
	return &AccountingWithdrawService{store: store}
}

func (s *AccountingWithdrawService) ApplyUserWithdraw(userID uint, input WithdrawApplyInput) (*models.ResellerWithdrawRequest, error) {
	if s == nil || s.store == nil || userID == 0 {
		return nil, ErrAccountingUnavailable
	}
	profile, err := s.store.GetProfileByUserID(userID)
	if err != nil {
		return nil, err
	}
	if profile == nil {
		return nil, ErrNotOpened
	}
	if err := RequireActiveProfile(profile); err != nil {
		return nil, err
	}
	return s.ApplyWithdraw(profile.ID, input)
}

func (s *AccountingWithdrawService) ApplyWithdraw(resellerID uint, input WithdrawApplyInput) (*models.ResellerWithdrawRequest, error) {
	if s == nil || s.store == nil || resellerID == 0 {
		return nil, ErrAccountingUnavailable
	}
	amount := input.Amount.Round(2)
	currency := strings.TrimSpace(input.Currency)
	channel := strings.TrimSpace(input.Channel)
	account := strings.TrimSpace(input.Account)
	if amount.LessThanOrEqual(decimal.Zero) {
		return nil, ErrWithdrawAmountInvalid
	}
	if currency == "" {
		return nil, ErrWithdrawCurrencyUnavailable
	}
	if channel == "" || account == "" {
		return nil, ErrWithdrawAmountInvalid
	}
	var createdID uint
	err := s.store.WithinTransaction(func(store AccountingWithdrawStore) error {
		balance, err := store.GetOrCreateBalanceAccountForUpdate(resellerID, currency)
		if err != nil {
			return err
		}
		if balance.Status == models.ResellerBalanceStatusNegativeBalance ||
			balance.Status == models.ResellerBalanceStatusFrozenReview ||
			balance.Status == models.ResellerBalanceStatusDisabled {
			return ErrBalanceAccountFrozen
		}
		// 可提现额必须以「净可用余额」为准（含退款扣减等负数流水），
		// 防止仅凭正数流水之和超额提现，导致账户被提成负余额、造成平台资损。
		availableSums, err := store.SumLedgerAmountGroupedByStatus(resellerID, currency, []string{models.ResellerLedgerStatusAvailable})
		if err != nil {
			return err
		}
		if amount.GreaterThan(availableSums[models.ResellerLedgerStatusAvailable].Round(2)) {
			return ErrWithdrawInsufficient
		}
		ledgers, err := store.ListAvailableLedgerEntriesForUpdate(resellerID, currency)
		if err != nil {
			return err
		}
		remaining := amount
		selectedIDs := make([]uint, 0)
		now := time.Now()
		for i := range ledgers {
			if remaining.LessThanOrEqual(decimal.Zero) {
				break
			}
			row := ledgers[i]
			rowAmount := row.Amount.Decimal.Round(2)
			if rowAmount.LessThanOrEqual(decimal.Zero) {
				continue
			}
			if rowAmount.LessThanOrEqual(remaining) {
				selectedIDs = append(selectedIDs, row.ID)
				remaining = remaining.Sub(rowAmount).Round(2)
				continue
			}
			lockAmount := remaining.Round(2)
			remainAmount := rowAmount.Sub(lockAmount).Round(2)
			row.Amount = money.FromDecimal(lockAmount)
			row.UpdatedAt = now
			if err := store.UpdateLedgerEntry(&row); err != nil {
				return err
			}
			remainRow := row
			remainRow.ID = 0
			remainRow.Amount = money.FromDecimal(remainAmount)
			remainRow.Status = models.ResellerLedgerStatusAvailable
			remainRow.WithdrawRequestID = nil
			remainRow.IdempotencyKey = fmt.Sprintf("split:%d:%d", row.ID, now.UnixNano())
			remainRow.CreatedAt = now
			remainRow.UpdatedAt = now
			if _, err := store.CreateLedgerEntryIfNotExists(&remainRow); err != nil {
				return err
			}
			selectedIDs = append(selectedIDs, row.ID)
			remaining = decimal.Zero
			break
		}
		if remaining.GreaterThan(decimal.Zero) {
			return ErrWithdrawInsufficient
		}
		req := &models.ResellerWithdrawRequest{
			ResellerID: resellerID,
			Amount:     money.FromDecimal(amount),
			Currency:   currency,
			Channel:    channel,
			Account:    account,
			Status:     models.ResellerWithdrawStatusPending,
			CreatedAt:  now,
			UpdatedAt:  now,
		}
		if err := store.CreateWithdrawRequest(req); err != nil {
			return err
		}
		if err := store.BatchUpdateLedgerEntries(selectedIDs, map[string]interface{}{
			"status":              models.ResellerLedgerStatusLocked,
			"withdraw_request_id": req.ID,
		}); err != nil {
			return err
		}
		createdID = req.ID
		return RefreshBalanceAccount(store, resellerID, currency, now)
	})
	if err != nil {
		return nil, err
	}
	return s.store.GetWithdrawRequestByID(createdID)
}

func (s *AccountingWithdrawService) ReviewWithdraw(adminID uint, withdrawID uint, action string, rejectReason string) (*models.ResellerWithdrawRequest, error) {
	if s == nil || s.store == nil || withdrawID == 0 {
		return nil, productcontract.ErrNotFound
	}
	act := strings.ToLower(strings.TrimSpace(action))
	if act != WithdrawActionReject && act != WithdrawActionPay {
		return nil, ErrWithdrawStatusInvalid
	}
	err := s.store.WithinTransaction(func(store AccountingWithdrawStore) error {
		req, err := store.GetWithdrawRequestByIDForUpdate(withdrawID)
		if err != nil {
			return err
		}
		if req == nil {
			return productcontract.ErrNotFound
		}
		if req.Status != models.ResellerWithdrawStatusPending {
			return ErrWithdrawStatusInvalid
		}
		now := time.Now()
		req.ProcessedBy = &adminID
		req.ProcessedAt = &now
		req.UpdatedAt = now
		if act == WithdrawActionReject {
			req.Status = models.ResellerWithdrawStatusRejected
			req.RejectReason = strings.TrimSpace(rejectReason)
			if err := store.BatchUpdateLedgerEntriesByWithdrawID(withdrawID, map[string]interface{}{
				"status":              models.ResellerLedgerStatusAvailable,
				"withdraw_request_id": nil,
			}); err != nil {
				return err
			}
		} else {
			req.Status = models.ResellerWithdrawStatusPaid
			req.RejectReason = ""
			if err := store.BatchUpdateLedgerEntriesByWithdrawID(withdrawID, map[string]interface{}{
				"status": models.ResellerLedgerStatusWithdrawn,
			}); err != nil {
				return err
			}
		}
		if err := store.UpdateWithdrawRequest(req); err != nil {
			return err
		}
		return RefreshBalanceAccount(store, req.ResellerID, req.Currency, now)
	})
	if err != nil {
		return nil, err
	}
	return s.store.GetWithdrawRequestByID(withdrawID)
}
