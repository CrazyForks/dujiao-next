package repository

import (
	"errors"
	"strings"
	"time"

	"github.com/dujiao-next/internal/models"
	resellermodule "github.com/dujiao-next/internal/modules/reseller"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ResellerLedgerScope 表示分销商 + 币种的账户维度，用于到期确认后定位需要刷新的余额账户。
type ResellerLedgerScope = resellermodule.LedgerScope

// CreateLedgerEntryIfNotExists 按幂等键创建分销账务流水。
func (r *GormResellerRepository) CreateLedgerEntryIfNotExists(entry *models.ResellerLedgerEntry) (bool, error) {
	if entry == nil {
		return false, errors.New("reseller ledger entry is nil")
	}
	key := strings.TrimSpace(entry.IdempotencyKey)
	if key == "" {
		return false, errors.New("reseller ledger idempotency key is empty")
	}
	existing, err := r.GetLedgerEntryByIdempotencyKey(key)
	if err != nil {
		return false, err
	}
	if existing != nil {
		return false, nil
	}
	now := time.Now()
	if entry.CreatedAt.IsZero() {
		entry.CreatedAt = now
	}
	entry.UpdatedAt = now
	if err := r.db.Create(entry).Error; err != nil {
		var again models.ResellerLedgerEntry
		if lookupErr := r.db.Where("idempotency_key = ?", key).First(&again).Error; lookupErr == nil {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// GetLedgerEntryByIdempotencyKey 按幂等键获取分销账务流水。
func (r *GormResellerRepository) GetLedgerEntryByIdempotencyKey(key string) (*models.ResellerLedgerEntry, error) {
	key = strings.TrimSpace(key)
	if key == "" {
		return nil, nil
	}
	var row models.ResellerLedgerEntry
	if err := r.db.Where("idempotency_key = ?", key).First(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &row, nil
}

// MarkDueLedgerEntriesAvailable 将到期确认流水转为可提现。
func (r *GormResellerRepository) MarkDueLedgerEntriesAvailable(now time.Time) (int64, error) {
	res := r.db.Model(&models.ResellerLedgerEntry{}).
		Where("status = ? AND available_at IS NOT NULL AND available_at <= ?", models.ResellerLedgerStatusPendingConfirm, now).
		Updates(map[string]interface{}{
			"status":     models.ResellerLedgerStatusAvailable,
			"updated_at": now,
		})
	return res.RowsAffected, res.Error
}

// ListDueLedgerScopes 列出到期待确认流水涉及的分销商与币种组合。
func (r *GormResellerRepository) ListDueLedgerScopes(now time.Time) ([]ResellerLedgerScope, error) {
	scopes := make([]ResellerLedgerScope, 0)
	err := r.db.Model(&models.ResellerLedgerEntry{}).
		Where("status = ? AND available_at IS NOT NULL AND available_at <= ?", models.ResellerLedgerStatusPendingConfirm, now).
		Group("reseller_id, currency").
		Select("reseller_id, currency").
		Scan(&scopes).Error
	if err != nil {
		return nil, err
	}
	return scopes, nil
}

// ListLedgerEntries 分页列出分销账务流水。
func (r *GormResellerRepository) ListLedgerEntries(filter ResellerLedgerListFilter) ([]models.ResellerLedgerEntry, int64, error) {
	rows := make([]models.ResellerLedgerEntry, 0)
	query := r.db.Model(&models.ResellerLedgerEntry{})
	if filter.ResellerID != 0 {
		query = query.Where("reseller_id = ?", filter.ResellerID)
	}
	if currency := strings.TrimSpace(filter.Currency); currency != "" {
		query = query.Where("currency = ?", currency)
	}
	if typ := strings.TrimSpace(filter.Type); typ != "" {
		query = query.Where("type = ?", typ)
	}
	if status := strings.TrimSpace(filter.Status); status != "" {
		query = query.Where("status = ?", status)
	}
	if filter.OrderID != 0 {
		query = query.Where("order_id = ?", filter.OrderID)
	}
	var total int64
	if err := query.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := applyPagination(query.Session(&gorm.Session{}), filter.Page, filter.PageSize).
		Order("id DESC").
		Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

// SumLedgerAmount 汇总指定状态的分销账务金额。
func (r *GormResellerRepository) SumLedgerAmount(resellerID uint, currency string, statuses []string) (decimal.Decimal, error) {
	currency = strings.TrimSpace(currency)
	if resellerID == 0 || currency == "" || len(statuses) == 0 {
		return decimal.Zero, nil
	}
	var total decimal.Decimal
	err := r.db.Model(&models.ResellerLedgerEntry{}).
		Where("reseller_id = ? AND currency = ? AND status IN ?", resellerID, currency, statuses).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&total).Error
	if err != nil {
		return decimal.Zero, err
	}
	return total.Round(2), nil
}

// SumLedgerAmountByOrderAndType 汇总指定订单、指定类型的流水金额（含正负号），用于退款扣减的累计上限保护。
func (r *GormResellerRepository) SumLedgerAmountByOrderAndType(orderID uint, ledgerType string) (decimal.Decimal, error) {
	ledgerType = strings.TrimSpace(ledgerType)
	if orderID == 0 || ledgerType == "" {
		return decimal.Zero, nil
	}
	var total decimal.Decimal
	err := r.db.Model(&models.ResellerLedgerEntry{}).
		Where("order_id = ? AND type = ?", orderID, ledgerType).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&total).Error
	if err != nil {
		return decimal.Zero, err
	}
	return total.Round(2), nil
}

// SumLedgerAmountGroupedByStatus 一次性按状态分组汇总分销账务金额，避免逐状态多次查询。
func (r *GormResellerRepository) SumLedgerAmountGroupedByStatus(resellerID uint, currency string, statuses []string) (map[string]decimal.Decimal, error) {
	currency = strings.TrimSpace(currency)
	result := make(map[string]decimal.Decimal, len(statuses))
	if resellerID == 0 || currency == "" || len(statuses) == 0 {
		return result, nil
	}
	type sumRow struct {
		Status string
		Total  decimal.Decimal
	}
	var rows []sumRow
	err := r.db.Model(&models.ResellerLedgerEntry{}).
		Where("reseller_id = ? AND currency = ? AND status IN ?", resellerID, currency, statuses).
		Group("status").
		Select("status, COALESCE(SUM(amount), 0) AS total").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		result[row.Status] = row.Total.Round(2)
	}
	return result, nil
}

// ListAvailableLedgerEntriesForUpdate 锁定指定币种可提现正向流水。
func (r *GormResellerRepository) ListAvailableLedgerEntriesForUpdate(resellerID uint, currency string) ([]models.ResellerLedgerEntry, error) {
	rows := make([]models.ResellerLedgerEntry, 0)
	currency = strings.TrimSpace(currency)
	if resellerID == 0 || currency == "" {
		return rows, nil
	}
	err := r.db.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("reseller_id = ? AND currency = ? AND status = ? AND withdraw_request_id IS NULL AND amount > 0",
			resellerID,
			currency,
			models.ResellerLedgerStatusAvailable,
		).
		Order("available_at ASC, id ASC").
		Find(&rows).Error
	return rows, err
}

// UpdateLedgerEntry 更新单条分销账务流水。
func (r *GormResellerRepository) UpdateLedgerEntry(entry *models.ResellerLedgerEntry) error {
	if entry == nil {
		return errors.New("reseller ledger entry is nil")
	}
	entry.UpdatedAt = time.Now()
	return r.db.Save(entry).Error
}

// BatchUpdateLedgerEntries 批量更新分销账务流水。
func (r *GormResellerRepository) BatchUpdateLedgerEntries(ids []uint, updates map[string]interface{}) error {
	if len(ids) == 0 {
		return nil
	}
	if updates == nil {
		updates = map[string]interface{}{}
	}
	updates["updated_at"] = time.Now()
	return r.db.Model(&models.ResellerLedgerEntry{}).Where("id IN ?", ids).Updates(updates).Error
}

// BatchUpdateLedgerEntriesByWithdrawID 按提现单 ID 批量更新分销账务流水。
func (r *GormResellerRepository) BatchUpdateLedgerEntriesByWithdrawID(withdrawID uint, updates map[string]interface{}) error {
	if withdrawID == 0 {
		return nil
	}
	if updates == nil {
		updates = map[string]interface{}{}
	}
	updates["updated_at"] = time.Now()
	return r.db.Model(&models.ResellerLedgerEntry{}).Where("withdraw_request_id = ?", withdrawID).Updates(updates).Error
}
