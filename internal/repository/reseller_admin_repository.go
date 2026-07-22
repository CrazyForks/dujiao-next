package repository

import (
	"strconv"
	"strings"

	"github.com/dujiao-next/internal/models"
	"gorm.io/gorm"
)

// ListAdminResellerLedgerEntries 分页列出管理端分销账务流水。
func (r *GormResellerRepository) ListAdminResellerLedgerEntries(filter ResellerAdminLedgerListFilter) ([]models.ResellerLedgerEntry, int64, error) {
	rows := make([]models.ResellerLedgerEntry, 0)
	query := r.db.Model(&models.ResellerLedgerEntry{}).
		Preload("Profile").
		Preload("Profile.User").
		Preload("Order")

	query = r.applyAdminResellerProfileFilters(query, "reseller_ledger_entries", filter.ResellerID, filter.UserID, filter.Keyword, "")
	if currency := strings.TrimSpace(filter.Currency); currency != "" {
		query = query.Where("reseller_ledger_entries.currency = ?", currency)
	}
	if typ := strings.TrimSpace(filter.Type); typ != "" {
		query = query.Where("reseller_ledger_entries.type = ?", typ)
	}
	if status := strings.TrimSpace(filter.Status); status != "" {
		query = query.Where("reseller_ledger_entries.status = ?", status)
	}
	if filter.OrderID != 0 {
		query = query.Where("reseller_ledger_entries.order_id = ?", filter.OrderID)
	}
	if orderNo := strings.TrimSpace(filter.OrderNo); orderNo != "" {
		query = query.Joins("LEFT JOIN orders o_filter ON o_filter.id = reseller_ledger_entries.order_id").
			Where("o_filter.order_no = ?", orderNo)
	}
	if filter.CreatedFrom != nil {
		query = query.Where("reseller_ledger_entries.created_at >= ?", *filter.CreatedFrom)
	}
	if filter.CreatedTo != nil {
		query = query.Where("reseller_ledger_entries.created_at <= ?", *filter.CreatedTo)
	}

	var total int64
	if err := query.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := applyPagination(query.Session(&gorm.Session{}), filter.Page, filter.PageSize).
		Order("reseller_ledger_entries.id DESC").
		Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

// ListAdminResellerBalanceAccounts 分页列出管理端分销余额账户。
func (r *GormResellerRepository) ListAdminResellerBalanceAccounts(filter ResellerAdminBalanceAccountListFilter) ([]models.ResellerBalanceAccount, int64, error) {
	rows := make([]models.ResellerBalanceAccount, 0)
	query := r.db.Model(&models.ResellerBalanceAccount{}).
		Preload("Profile").
		Preload("Profile.User")

	query = r.applyAdminResellerProfileFilters(query, "reseller_balance_accounts", filter.ResellerID, filter.UserID, filter.Keyword, "")
	if currency := strings.TrimSpace(filter.Currency); currency != "" {
		query = query.Where("reseller_balance_accounts.currency = ?", currency)
	}
	if status := strings.TrimSpace(filter.Status); status != "" {
		query = query.Where("reseller_balance_accounts.status = ?", status)
	}

	var total int64
	if err := query.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := applyPagination(query.Session(&gorm.Session{}), filter.Page, filter.PageSize).
		Order("reseller_balance_accounts.id DESC").
		Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

// ListAdminResellerWithdrawRequests 分页列出管理端分销提现申请。
func (r *GormResellerRepository) ListAdminResellerWithdrawRequests(filter ResellerAdminWithdrawListFilter) ([]models.ResellerWithdrawRequest, int64, error) {
	rows := make([]models.ResellerWithdrawRequest, 0)
	query := r.db.Model(&models.ResellerWithdrawRequest{}).
		Preload("Profile").
		Preload("Profile.User").
		Preload("Processor", "deleted_at IS NULL")

	query = r.applyAdminResellerProfileFilters(query, "reseller_withdraw_requests", filter.ResellerID, filter.UserID, filter.Keyword, "reseller_withdraw_requests.account")
	if currency := strings.TrimSpace(filter.Currency); currency != "" {
		query = query.Where("reseller_withdraw_requests.currency = ?", currency)
	}
	if status := strings.TrimSpace(filter.Status); status != "" {
		query = query.Where("reseller_withdraw_requests.status = ?", status)
	}
	if filter.CreatedFrom != nil {
		query = query.Where("reseller_withdraw_requests.created_at >= ?", *filter.CreatedFrom)
	}
	if filter.CreatedTo != nil {
		query = query.Where("reseller_withdraw_requests.created_at <= ?", *filter.CreatedTo)
	}

	var total int64
	if err := query.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := applyPagination(query.Session(&gorm.Session{}), filter.Page, filter.PageSize).
		Order("reseller_withdraw_requests.id DESC").
		Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

func (r *GormResellerRepository) applyAdminResellerProfileFilters(query *gorm.DB, table string, resellerID uint, userID uint, keyword string, accountColumn string) *gorm.DB {
	if resellerID != 0 {
		query = query.Where(table+".reseller_id = ?", resellerID)
	}
	keyword = strings.TrimSpace(keyword)
	if userID == 0 && keyword == "" {
		return query
	}

	query = query.
		Joins("LEFT JOIN reseller_profiles rp_filter ON rp_filter.id = " + table + ".reseller_id").
		Joins("LEFT JOIN users u_filter ON u_filter.id = rp_filter.user_id")
	if userID != 0 {
		query = query.Where("rp_filter.user_id = ?", userID)
	}
	if keyword == "" {
		return query
	}

	like := "%" + keyword + "%"
	conditions := []string{"u_filter.email LIKE ?", "u_filter.display_name LIKE ?"}
	args := []interface{}{like, like}
	if id, err := strconv.ParseUint(keyword, 10, 64); err == nil && id > 0 {
		conditions = append(conditions, "rp_filter.id = ?")
		args = append(args, uint(id))
	}
	if accountColumn != "" {
		conditions = append(conditions, accountColumn+" LIKE ?")
		args = append(args, like)
	}
	return query.Where("("+strings.Join(conditions, " OR ")+")", args...)
}
