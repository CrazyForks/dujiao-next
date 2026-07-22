package repository

import (
	"errors"
	"strings"
	"time"

	"github.com/dujiao-next/internal/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// GetOrCreateBalanceAccountForUpdate 获取或创建并锁定同币种余额账户。
func (r *GormResellerRepository) GetOrCreateBalanceAccountForUpdate(resellerID uint, currency string) (*models.ResellerBalanceAccount, error) {
	currency = strings.TrimSpace(currency)
	if resellerID == 0 || currency == "" {
		return nil, errors.New("invalid reseller balance account")
	}
	var row models.ResellerBalanceAccount
	err := r.db.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("reseller_id = ? AND currency = ?", resellerID, currency).
		First(&row).Error
	if err == nil {
		return &row, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	now := time.Now()
	row = models.ResellerBalanceAccount{
		ResellerID: resellerID,
		Currency:   currency,
		Status:     models.ResellerBalanceStatusNormal,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	if err := r.db.Create(&row).Error; err != nil {
		return nil, err
	}
	if err := r.db.Clauses(clause.Locking{Strength: "UPDATE"}).First(&row, row.ID).Error; err != nil {
		return nil, err
	}
	return &row, nil
}

// ListBalanceAccounts 分页列出分销商余额账户。
func (r *GormResellerRepository) ListBalanceAccounts(filter ResellerBalanceAccountListFilter) ([]models.ResellerBalanceAccount, int64, error) {
	rows := make([]models.ResellerBalanceAccount, 0)
	query := r.db.Model(&models.ResellerBalanceAccount{})
	if filter.ResellerID != 0 {
		query = query.Where("reseller_id = ?", filter.ResellerID)
	}
	if currency := strings.TrimSpace(filter.Currency); currency != "" {
		query = query.Where("currency = ?", currency)
	}
	if status := strings.TrimSpace(filter.Status); status != "" {
		query = query.Where("status = ?", status)
	}
	var total int64
	if err := query.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := applyPagination(query.Session(&gorm.Session{}), filter.Page, filter.PageSize).
		Order("currency ASC, id DESC").
		Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

// UpdateBalanceAccount 更新分销余额账户缓存。
func (r *GormResellerRepository) UpdateBalanceAccount(account *models.ResellerBalanceAccount) error {
	if account == nil {
		return errors.New("reseller balance account is nil")
	}
	account.UpdatedAt = time.Now()
	return r.db.Save(account).Error
}
