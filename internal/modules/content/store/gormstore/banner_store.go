package gormstore

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/dujiao-next/internal/models"
	"github.com/dujiao-next/internal/modules/content"
	"gorm.io/gorm"
)

// BannerStore 使用 GORM 持久化 Banner。
type BannerStore struct {
	db *gorm.DB
}

var _ content.BannerStore = (*BannerStore)(nil)

// NewBannerStore 创建 Banner 持久化适配器。
func NewBannerStore(db *gorm.DB) *BannerStore {
	return &BannerStore{db: db}
}

func (s *BannerStore) List(ctx context.Context, query content.BannerQuery) ([]models.Banner, int64, error) {
	var banners []models.Banner
	db := withContext(s.db, ctx)
	statement := db.Model(&models.Banner{})

	if query.Position != "" {
		statement = statement.Where("position = ?", query.Position)
	}
	if query.IsActive != nil {
		statement = statement.Where("is_active = ?", *query.IsActive)
	}
	if search := strings.TrimSpace(query.Search); search != "" {
		like := "%" + search + "%"
		condition, argCount := buildLocalizedLikeCondition(db, []string{"name"}, []string{"title_json"})
		statement = statement.Where(condition, repeatLikeArgs(like, argCount)...)
	}

	var total int64
	if err := statement.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	statement = applyPagination(statement, query.Page, query.PageSize)
	if err := statement.Order("sort_order DESC, created_at DESC").Find(&banners).Error; err != nil {
		return nil, 0, err
	}
	return banners, total, nil
}

func (s *BannerStore) ListValidByPosition(ctx context.Context, position string, limit int, now time.Time) ([]models.Banner, error) {
	var banners []models.Banner
	statement := withContext(s.db, ctx).Model(&models.Banner{}).
		Where("is_active = ?", true).
		Where("(start_at IS NULL OR start_at <= ?)", now).
		Where("(end_at IS NULL OR end_at >= ?)", now)
	if position != "" {
		statement = statement.Where("position = ?", position)
	}
	if limit > 0 {
		statement = statement.Limit(limit)
	}
	if err := statement.Order("sort_order DESC, created_at DESC").Find(&banners).Error; err != nil {
		return nil, err
	}
	return banners, nil
}

func (s *BannerStore) GetByID(ctx context.Context, id string) (*models.Banner, error) {
	var banner models.Banner
	if err := withContext(s.db, ctx).First(&banner, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &banner, nil
}

func (s *BannerStore) Create(ctx context.Context, banner *models.Banner) error {
	requestedActive := banner.IsActive
	return withContext(s.db, ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(banner).Error; err != nil {
			return err
		}
		if requestedActive {
			return nil
		}

		// IsActive 带有 default:true，GORM Create 会把 false 当作零值省略并
		// 回填为 true。用同一事务内的显式列更新保留调用方的 false 语义。
		if err := tx.Model(banner).UpdateColumn("is_active", false).Error; err != nil {
			return err
		}
		banner.IsActive = false
		return nil
	})
}

func (s *BannerStore) Update(ctx context.Context, banner *models.Banner) error {
	return withContext(s.db, ctx).Save(banner).Error
}

func (s *BannerStore) Delete(ctx context.Context, id string) error {
	return withContext(s.db, ctx).Delete(&models.Banner{}, id).Error
}
