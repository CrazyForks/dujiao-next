package repository

import (
	"errors"
	"strings"
	"time"

	"github.com/dujiao-next/internal/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// UpsertDomain 创建域名，或恢复同域名的软删除记录。
func (r *GormResellerRepository) UpsertDomain(input models.ResellerDomain) (*models.ResellerDomain, error) {
	input.Domain = normalizeDomainForRepository(input.Domain)
	if input.ResellerID == 0 || input.Domain == "" {
		return nil, errors.New("invalid reseller domain")
	}
	now := time.Now()
	var existing models.ResellerDomain
	err := r.db.Unscoped().Where("domain = ?", input.Domain).First(&existing).Error
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
	if !existing.DeletedAt.Valid {
		return nil, errors.New("reseller domain already exists")
	}
	existing.ResellerID = input.ResellerID
	existing.Type = input.Type
	existing.VerificationToken = input.VerificationToken
	existing.VerificationStatus = input.VerificationStatus
	existing.Status = input.Status
	existing.IsPrimary = input.IsPrimary
	existing.VerifiedAt = input.VerifiedAt
	existing.DeletedAt = gorm.DeletedAt{}
	existing.UpdatedAt = now
	if err := r.db.Unscoped().Save(&existing).Error; err != nil {
		return nil, err
	}
	return &existing, nil
}

// GetDomainByID 按 ID 获取域名。
func (r *GormResellerRepository) GetDomainByID(id uint) (*models.ResellerDomain, error) {
	if id == 0 {
		return nil, nil
	}
	var row models.ResellerDomain
	if err := r.db.Preload("Profile.User", "deleted_at IS NULL").First(&row, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &row, nil
}

// GetDomainByIDForUpdate 按 ID 获取并锁定域名。
func (r *GormResellerRepository) GetDomainByIDForUpdate(id uint) (*models.ResellerDomain, error) {
	if id == 0 {
		return nil, nil
	}
	var row models.ResellerDomain
	if err := r.db.Clauses(clause.Locking{Strength: "UPDATE"}).Preload("Profile.User", "deleted_at IS NULL").First(&row, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &row, nil
}

// UpdateDomain 更新域名。
func (r *GormResellerRepository) UpdateDomain(domain *models.ResellerDomain) error {
	if domain == nil || domain.ID == 0 {
		return errors.New("invalid reseller domain")
	}
	domain.Domain = normalizeDomainForRepository(domain.Domain)
	return r.db.Save(domain).Error
}

// FindDomainByHost 按域名获取未删除域名记录。
func (r *GormResellerRepository) FindDomainByHost(host string) (*models.ResellerDomain, error) {
	domain := normalizeDomainForRepository(host)
	if domain == "" {
		return nil, nil
	}
	var row models.ResellerDomain
	err := r.db.Preload("Profile").Where("domain = ?", domain).First(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &row, nil
}

// FindActiveVerifiedDomain 按域名获取已验证且启用的分销域名。
func (r *GormResellerRepository) FindActiveVerifiedDomain(host string) (*models.ResellerDomain, error) {
	domain := normalizeDomainForRepository(host)
	if domain == "" {
		return nil, nil
	}
	var row models.ResellerDomain
	err := r.db.Preload("Profile").
		Joins("JOIN reseller_profiles ON reseller_profiles.id = reseller_domains.reseller_id AND reseller_profiles.deleted_at IS NULL").
		Where("reseller_domains.domain = ? AND reseller_domains.status = ? AND reseller_domains.verification_status = ?", domain, models.ResellerDomainStatusActive, models.ResellerDomainVerificationVerified).
		Where("reseller_profiles.status = ?", models.ResellerProfileStatusActive).
		First(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &row, nil
}

// ListDomainsByResellerID 列出分销商名下所有未删除域名。
func (r *GormResellerRepository) ListDomainsByResellerID(resellerID uint) ([]models.ResellerDomain, error) {
	rows := make([]models.ResellerDomain, 0)
	if resellerID == 0 {
		return rows, nil
	}
	if err := r.db.Where("reseller_id = ?", resellerID).Order("is_primary DESC, id ASC").Find(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

// ListDomains 分页列出分销商域名。
func (r *GormResellerRepository) ListDomains(filter ResellerDomainListFilter) ([]models.ResellerDomain, int64, error) {
	rows := make([]models.ResellerDomain, 0)
	query := r.db.Model(&models.ResellerDomain{}).Preload("Profile.User", "deleted_at IS NULL")
	if filter.ResellerID > 0 {
		query = query.Where("reseller_domains.reseller_id = ?", filter.ResellerID)
	}
	if filter.UserID > 0 {
		query = query.Joins("JOIN reseller_profiles rp_user_filter ON rp_user_filter.id = reseller_domains.reseller_id").
			Where("rp_user_filter.user_id = ?", filter.UserID)
	}
	if domain := strings.TrimSpace(filter.Domain); domain != "" {
		query = query.Where("reseller_domains.domain = ?", normalizeDomainForRepository(domain))
	}
	if typ := strings.TrimSpace(filter.Type); typ != "" {
		query = query.Where("reseller_domains.type = ?", typ)
	}
	if status := strings.TrimSpace(filter.Status); status != "" {
		query = query.Where("reseller_domains.status = ?", status)
	}
	if verification := strings.TrimSpace(filter.VerificationStatus); verification != "" {
		query = query.Where("reseller_domains.verification_status = ?", verification)
	}
	if keyword := strings.TrimSpace(filter.Keyword); keyword != "" {
		like := "%" + strings.ToLower(keyword) + "%"
		query = query.Joins("LEFT JOIN reseller_profiles rp_keyword ON rp_keyword.id = reseller_domains.reseller_id").
			Joins("LEFT JOIN users ON users.id = rp_keyword.user_id").
			Where("LOWER(reseller_domains.domain) LIKE ? OR LOWER(users.email) LIKE ? OR LOWER(users.display_name) LIKE ? OR CAST(reseller_domains.reseller_id AS TEXT) = ?", like, like, like, keyword)
	}
	if filter.CreatedFrom != nil {
		query = query.Where("reseller_domains.created_at >= ?", *filter.CreatedFrom)
	}
	if filter.CreatedTo != nil {
		query = query.Where("reseller_domains.created_at <= ?", *filter.CreatedTo)
	}
	var total int64
	if err := query.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := applyPagination(query.Session(&gorm.Session{}), filter.Page, filter.PageSize).
		Order("reseller_domains.id DESC").
		Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

func normalizeDomainForRepository(raw string) string {
	return strings.TrimSuffix(strings.ToLower(strings.TrimSpace(raw)), ".")
}
