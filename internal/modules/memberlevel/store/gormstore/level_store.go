package gormstore

import (
	"errors"

	"github.com/dujiao-next/internal/models"
	"github.com/dujiao-next/internal/modules/memberlevel"

	"gorm.io/gorm"
)

type LevelStore struct {
	db *gorm.DB
}

func NewLevelStore(db *gorm.DB) *LevelStore {
	return &LevelStore{db: db}
}

func (r *LevelStore) GetByID(id uint) (*models.MemberLevel, error) {
	var level models.MemberLevel
	if err := r.db.First(&level, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &level, nil
}

func (r *LevelStore) GetBySlug(slug string) (*models.MemberLevel, error) {
	var level models.MemberLevel
	if err := r.db.Where("slug = ?", slug).First(&level).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &level, nil
}

func (r *LevelStore) GetDefault() (*models.MemberLevel, error) {
	var level models.MemberLevel
	if err := r.db.Where("is_default = ? AND is_active = ?", true, true).First(&level).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &level, nil
}

func (r *LevelStore) GetActiveBySortOrder(sortOrder int, excludeID uint) (*models.MemberLevel, error) {
	var level models.MemberLevel
	query := r.db.Where("is_active = ? AND sort_order = ?", true, sortOrder)
	if excludeID > 0 {
		query = query.Where("id != ?", excludeID)
	}
	if err := query.First(&level).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &level, nil
}

// ListAllActive 获取所有启用的等级，按 sort_order DESC
func (r *LevelStore) ListAllActive() ([]models.MemberLevel, error) {
	var levels []models.MemberLevel
	if err := r.db.Where("is_active = ?", true).Order("sort_order desc, id asc").Find(&levels).Error; err != nil {
		return nil, err
	}
	return levels, nil
}

func (r *LevelStore) Create(level *models.MemberLevel) error {
	return r.db.Create(level).Error
}

func (r *LevelStore) Update(level *models.MemberLevel) error {
	return r.db.Save(level).Error
}

func (r *LevelStore) Delete(id uint) error {
	return r.db.Delete(&models.MemberLevel{}, id).Error
}

func (r *LevelStore) List(filter memberlevel.ListFilter) ([]models.MemberLevel, int64, error) {
	var levels []models.MemberLevel
	query := r.db.Model(&models.MemberLevel{})

	if filter.IsActive != nil {
		query = query.Where("is_active = ?", *filter.IsActive)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if filter.Page > 0 && filter.PageSize > 0 {
		query = query.Offset((filter.Page - 1) * filter.PageSize).Limit(filter.PageSize)
	}

	if err := query.Order("sort_order desc, id asc").Find(&levels).Error; err != nil {
		return nil, 0, err
	}
	return levels, total, nil
}

// ClearDefault 清除默认标记（排除指定ID）
func (r *LevelStore) ClearDefault(excludeID uint) error {
	query := r.db.Model(&models.MemberLevel{}).Where("is_default = ?", true)
	if excludeID > 0 {
		query = query.Where("id != ?", excludeID)
	}
	return query.Update("is_default", false).Error
}

var _ memberlevel.LevelRepository = (*LevelStore)(nil)
