package gormstore

import (
	"context"
	"errors"
	"strings"

	"github.com/dujiao-next/internal/models"
	"github.com/dujiao-next/internal/modules/content"
	"gorm.io/gorm"
)

// MediaStore 使用 GORM 持久化素材元数据。
type MediaStore struct {
	db *gorm.DB
}

var _ content.MediaStore = (*MediaStore)(nil)

// NewMediaStore 创建素材元数据持久化适配器。
func NewMediaStore(db *gorm.DB) *MediaStore {
	return &MediaStore{db: db}
}

func (s *MediaStore) List(ctx context.Context, query content.MediaQuery) ([]models.Media, int64, error) {
	var items []models.Media
	db := withContext(s.db, ctx)
	statement := db.Model(&models.Media{})
	if query.Scene != "" {
		statement = statement.Where("scene = ?", query.Scene)
	}
	if search := strings.TrimSpace(query.Search); search != "" {
		like := "%" + search + "%"
		operator := likeOperatorByDialect(dbDialectName(db))
		statement = statement.Where("name "+operator+" ? OR filename "+operator+" ?", like, like)
	}

	var total int64
	if err := statement.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	statement = applyPagination(statement, query.Page, query.PageSize)
	if err := statement.Order("created_at DESC").Find(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (s *MediaStore) GetByID(ctx context.Context, id uint) (*models.Media, error) {
	var media models.Media
	if err := withContext(s.db, ctx).First(&media, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &media, nil
}

func (s *MediaStore) GetByPath(ctx context.Context, path string) (*models.Media, error) {
	var media models.Media
	if err := withContext(s.db, ctx).Where("path = ?", path).First(&media).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &media, nil
}

func (s *MediaStore) Create(ctx context.Context, media *models.Media) error {
	return withContext(s.db, ctx).Create(media).Error
}

func (s *MediaStore) Update(ctx context.Context, media *models.Media) error {
	return withContext(s.db, ctx).Save(media).Error
}

func (s *MediaStore) Delete(ctx context.Context, id uint) error {
	return withContext(s.db, ctx).Delete(&models.Media{}, id).Error
}
