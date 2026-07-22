package repository

import (
	"errors"
	"strings"
	"time"

	"github.com/dujiao-next/internal/models"
	"gorm.io/gorm"
)

// UpsertSiteConfig 创建或恢复分销站点配置。
func (r *GormResellerRepository) UpsertSiteConfig(input models.ResellerSiteConfig) (*models.ResellerSiteConfig, error) {
	if input.ResellerID == 0 {
		return nil, errors.New("invalid reseller site config")
	}
	now := time.Now()
	var existing models.ResellerSiteConfig
	err := r.db.Unscoped().Where("reseller_id = ?", input.ResellerID).First(&existing).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		input.CreatedAt = now
		input.UpdatedAt = now
		if err := r.db.Create(&input).Error; err != nil {
			return nil, err
		}
		return &input, nil
	}
	existing.SiteName = input.SiteName
	existing.Logo = input.Logo
	existing.Favicon = input.Favicon
	existing.AnnouncementJSON = input.AnnouncementJSON
	existing.SupportJSON = input.SupportJSON
	existing.SEOJSON = input.SEOJSON
	existing.FooterLinksJSON = input.FooterLinksJSON
	existing.NavConfigJSON = input.NavConfigJSON
	existing.ThemeJSON = input.ThemeJSON
	existing.DeletedAt = gorm.DeletedAt{}
	existing.UpdatedAt = now
	if err := r.db.Unscoped().Save(&existing).Error; err != nil {
		return nil, err
	}
	return &existing, nil
}

// GetSiteConfigByResellerID 按分销商资料 ID 获取站点配置。
func (r *GormResellerRepository) GetSiteConfigByResellerID(resellerID uint) (*models.ResellerSiteConfig, error) {
	if resellerID == 0 {
		return nil, nil
	}
	var row models.ResellerSiteConfig
	err := r.db.Preload("Profile").Preload("Profile.User", "deleted_at IS NULL").
		Where("reseller_id = ?", resellerID).
		First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

// DeleteSiteConfigByResellerID 软删除分销商站点配置。
func (r *GormResellerRepository) DeleteSiteConfigByResellerID(resellerID uint) error {
	if resellerID == 0 {
		return nil
	}
	return r.db.Where("reseller_id = ?", resellerID).Delete(&models.ResellerSiteConfig{}).Error
}

// ListSiteConfigs 查询分销商站点配置列表。
func (r *GormResellerRepository) ListSiteConfigs(filter ResellerSiteConfigListFilter) ([]models.ResellerSiteConfig, int64, error) {
	var rows []models.ResellerSiteConfig
	query := r.db.Model(&models.ResellerSiteConfig{}).
		Preload("Profile").
		Preload("Profile.User", "deleted_at IS NULL")
	if filter.ResellerID > 0 {
		query = query.Where("reseller_site_configs.reseller_id = ?", filter.ResellerID)
	}
	if keyword := strings.TrimSpace(filter.Keyword); keyword != "" {
		like := "%" + strings.ToLower(keyword) + "%"
		query = query.Joins("LEFT JOIN reseller_profiles ON reseller_profiles.id = reseller_site_configs.reseller_id").
			Joins("LEFT JOIN users ON users.id = reseller_profiles.user_id").
			Where("LOWER(reseller_site_configs.site_name) LIKE ? OR LOWER(users.email) LIKE ? OR LOWER(users.display_name) LIKE ? OR CAST(reseller_site_configs.reseller_id AS TEXT) = ?", like, like, like, keyword)
	}
	if filter.CreatedFrom != nil {
		query = query.Where("reseller_site_configs.created_at >= ?", *filter.CreatedFrom)
	}
	if filter.CreatedTo != nil {
		query = query.Where("reseller_site_configs.created_at <= ?", *filter.CreatedTo)
	}
	var total int64
	if err := query.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := applyPagination(query.Session(&gorm.Session{}), filter.Page, filter.PageSize).
		Order("reseller_site_configs.id DESC").
		Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}
