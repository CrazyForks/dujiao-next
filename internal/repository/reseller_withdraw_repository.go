package repository

import (
	"errors"
	"strings"
	"time"

	"github.com/dujiao-next/internal/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// CreateWithdrawRequest 创建分销提现申请。
func (r *GormResellerRepository) CreateWithdrawRequest(req *models.ResellerWithdrawRequest) error {
	if req == nil {
		return errors.New("reseller withdraw request is nil")
	}
	now := time.Now()
	if req.CreatedAt.IsZero() {
		req.CreatedAt = now
	}
	req.UpdatedAt = now
	return r.db.Create(req).Error
}

// GetWithdrawRequestByID 按 ID 获取分销提现申请。
func (r *GormResellerRepository) GetWithdrawRequestByID(id uint) (*models.ResellerWithdrawRequest, error) {
	if id == 0 {
		return nil, nil
	}
	var row models.ResellerWithdrawRequest
	if err := r.db.First(&row, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &row, nil
}

// GetWithdrawRequestByIDForUpdate 按 ID 获取并锁定分销提现申请。
func (r *GormResellerRepository) GetWithdrawRequestByIDForUpdate(id uint) (*models.ResellerWithdrawRequest, error) {
	if id == 0 {
		return nil, nil
	}
	var row models.ResellerWithdrawRequest
	if err := r.db.Clauses(clause.Locking{Strength: "UPDATE"}).First(&row, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &row, nil
}

// UpdateWithdrawRequest 更新分销提现申请。
func (r *GormResellerRepository) UpdateWithdrawRequest(req *models.ResellerWithdrawRequest) error {
	if req == nil {
		return errors.New("reseller withdraw request is nil")
	}
	req.UpdatedAt = time.Now()
	return r.db.Save(req).Error
}

// ListWithdrawRequests 分页列出分销提现申请。
func (r *GormResellerRepository) ListWithdrawRequests(filter ResellerWithdrawListFilter) ([]models.ResellerWithdrawRequest, int64, error) {
	rows := make([]models.ResellerWithdrawRequest, 0)
	query := r.db.Model(&models.ResellerWithdrawRequest{})
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
		Order("id DESC").
		Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}
