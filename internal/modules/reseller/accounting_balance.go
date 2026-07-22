package reseller

import (
	"strings"
	"time"

	"github.com/dujiao-next/internal/models"
	"github.com/shopspring/decimal"
)

// RefreshBalanceAccount 按账务流水重算余额账户缓存。
func RefreshBalanceAccount(store BalanceAccountStore, resellerID uint, currency string, now time.Time) error {
	currency = strings.TrimSpace(currency)
	if store == nil || resellerID == 0 || currency == "" {
		return nil
	}
	account, err := store.GetOrCreateBalanceAccountForUpdate(resellerID, currency)
	if err != nil {
		return err
	}
	sums, err := store.SumLedgerAmountGroupedByStatus(resellerID, currency, []string{
		models.ResellerLedgerStatusAvailable,
		models.ResellerLedgerStatusLocked,
	})
	if err != nil {
		return err
	}
	available := sums[models.ResellerLedgerStatusAvailable]
	locked := sums[models.ResellerLedgerStatusLocked]
	net := available.Round(2)
	negative := decimal.Zero
	if net.LessThan(decimal.Zero) {
		negative = net.Abs().Round(2)
		account.Status = models.ResellerBalanceStatusNegativeBalance
	} else if account.Status == models.ResellerBalanceStatusNegativeBalance {
		account.Status = models.ResellerBalanceStatusNormal
	}
	account.AvailableAmountCache = models.NewMoneyFromDecimal(net)
	account.LockedAmountCache = models.NewMoneyFromDecimal(locked.Round(2))
	account.NegativeAmountCache = models.NewMoneyFromDecimal(negative)
	account.UpdatedAt = now
	return store.UpdateBalanceAccount(account)
}
