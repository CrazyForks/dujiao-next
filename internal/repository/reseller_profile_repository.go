package repository

import (
	"errors"
	"strings"

	"github.com/dujiao-next/internal/models"
	"gorm.io/gorm"
)

// CreateProfile 创建分销商资料。
func (r *GormResellerRepository) CreateProfile(profile *models.ResellerProfile) error {
	if profile == nil {
		return errors.New("reseller profile is nil")
	}
	return r.db.Create(profile).Error
}

// GetProfileByID 按 ID 获取分销商资料。
func (r *GormResellerRepository) GetProfileByID(id uint) (*models.ResellerProfile, error) {
	if id == 0 {
		return nil, nil
	}
	var profile models.ResellerProfile
	if err := r.db.Preload("User").First(&profile, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &profile, nil
}

// GetProfileByUserID 按用户 ID 获取分销商资料。
func (r *GormResellerRepository) GetProfileByUserID(userID uint) (*models.ResellerProfile, error) {
	if userID == 0 {
		return nil, nil
	}
	var profile models.ResellerProfile
	if err := r.db.Preload("User").Where("user_id = ?", userID).First(&profile).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &profile, nil
}

// UpdateProfile 更新分销商资料。
func (r *GormResellerRepository) UpdateProfile(profile *models.ResellerProfile) error {
	if profile == nil || profile.ID == 0 {
		return errors.New("invalid reseller profile")
	}
	return r.db.Save(profile).Error
}

// ListProfiles 分页列出分销商资料。
func (r *GormResellerRepository) ListProfiles(filter ResellerProfileListFilter) ([]models.ResellerProfile, int64, error) {
	rows := make([]models.ResellerProfile, 0)
	query := r.db.Model(&models.ResellerProfile{}).Preload("User")
	if filter.UserID > 0 {
		query = query.Where("reseller_profiles.user_id = ?", filter.UserID)
	}
	if status := strings.TrimSpace(filter.Status); status != "" {
		query = query.Where("reseller_profiles.status = ?", status)
	}
	if settlement := strings.TrimSpace(filter.SettlementStatus); settlement != "" {
		query = query.Where("reseller_profiles.settlement_status = ?", settlement)
	}
	if keyword := strings.TrimSpace(filter.Keyword); keyword != "" {
		like := "%" + strings.ToLower(keyword) + "%"
		query = query.Joins("LEFT JOIN users ON users.id = reseller_profiles.user_id").
			Where("LOWER(users.email) LIKE ? OR LOWER(users.display_name) LIKE ? OR CAST(reseller_profiles.id AS TEXT) = ?", like, like, keyword)
	}
	if filter.CreatedFrom != nil {
		query = query.Where("reseller_profiles.created_at >= ?", *filter.CreatedFrom)
	}
	if filter.CreatedTo != nil {
		query = query.Where("reseller_profiles.created_at <= ?", *filter.CreatedTo)
	}
	var total int64
	if err := query.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := applyPagination(query.Session(&gorm.Session{}), filter.Page, filter.PageSize).
		Order("reseller_profiles.id DESC").
		Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

// IsActiveRelatedAccount 判断用户是否为分销商已启用的关联账号。
func (r *GormResellerRepository) IsActiveRelatedAccount(resellerID uint, userID uint) (bool, error) {
	if resellerID == 0 || userID == 0 {
		return false, nil
	}
	var count int64
	if err := r.db.Model(&models.ResellerRelatedAccount{}).
		Where("reseller_id = ? AND user_id = ? AND status = ?", resellerID, userID, models.ResellerRelatedAccountStatusActive).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}
