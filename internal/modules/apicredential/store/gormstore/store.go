package gormstore

import (
	"errors"

	"github.com/dujiao-next/internal/models"
	"github.com/dujiao-next/internal/modules/apicredential"

	"gorm.io/gorm"
)

type Store struct {
	db *gorm.DB
}

func New(db *gorm.DB) *Store {
	return &Store{db: db}
}

// GetByID 根据 ID 获取
func (r *Store) GetByID(id uint) (*models.ApiCredential, error) {
	var cred models.ApiCredential
	if err := r.db.First(&cred, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &cred, nil
}

// GetByUserID 根据用户 ID 获取
func (r *Store) GetByUserID(userID uint) (*models.ApiCredential, error) {
	var cred models.ApiCredential
	if err := r.db.Where("user_id = ?", userID).First(&cred).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &cred, nil
}

// GetAnyByUserID 根据用户 ID 获取，包含软删除记录。
func (r *Store) GetAnyByUserID(userID uint) (*models.ApiCredential, error) {
	var cred models.ApiCredential
	if err := r.db.Unscoped().Where("user_id = ?", userID).First(&cred).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &cred, nil
}

// GetByApiKey 根据 API Key 获取（预加载 User 用于状态校验）
func (r *Store) GetByApiKey(apiKey string) (*models.ApiCredential, error) {
	var cred models.ApiCredential
	if err := r.db.Preload("User", "deleted_at IS NULL").Where("api_key = ?", apiKey).First(&cred).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &cred, nil
}

// Create 创建凭证
func (r *Store) Create(cred *models.ApiCredential) error {
	return r.db.Create(cred).Error
}

// Update 更新凭证
func (r *Store) Update(cred *models.ApiCredential) error {
	return r.db.Save(cred).Error
}

// UpdateAny 更新凭证，包含软删除记录。
func (r *Store) UpdateAny(cred *models.ApiCredential) error {
	return r.db.Unscoped().Save(cred).Error
}

// Delete 软删除凭证
func (r *Store) Delete(id uint) error {
	return r.db.Delete(&models.ApiCredential{}, id).Error
}

// List 列表查询
func (r *Store) List(filter apicredential.ListFilter) ([]models.ApiCredential, int64, error) {
	var creds []models.ApiCredential
	var total int64

	q := r.db.Model(&models.ApiCredential{})
	if filter.Status != "" {
		q = q.Where("status = ?", filter.Status)
	}
	if filter.UserID > 0 {
		q = q.Where("user_id = ?", filter.UserID)
	}
	if filter.Search != "" {
		// 按邮箱或昵称搜索：需要 JOIN users 表
		q = q.Joins("JOIN users ON users.id = api_credentials.user_id").
			Where("users.email LIKE ? OR users.display_name LIKE ?",
				"%"+filter.Search+"%", "%"+filter.Search+"%")
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	q = q.Order("api_credentials.created_at DESC")
	if filter.Page > 0 && filter.PageSize > 0 {
		q = q.Offset((filter.Page - 1) * filter.PageSize).Limit(filter.PageSize)
	}

	if err := q.Preload("User", "deleted_at IS NULL").Find(&creds).Error; err != nil {
		return nil, 0, err
	}

	return creds, total, nil
}

var _ apicredential.Repository = (*Store)(nil)
