package gormstore

import (
	"context"
	"errors"

	"github.com/dujiao-next/internal/models"
	"github.com/dujiao-next/internal/modules/content"
	"gorm.io/gorm"
)

// PostCategoryStore 使用 GORM 持久化文章分类。
type PostCategoryStore struct {
	db *gorm.DB
}

var _ content.PostCategoryStore = (*PostCategoryStore)(nil)

// NewPostCategoryStore 创建文章分类持久化适配器。
func NewPostCategoryStore(db *gorm.DB) *PostCategoryStore {
	return &PostCategoryStore{db: db}
}

func (s *PostCategoryStore) ListAll(ctx context.Context, parentID *uint) ([]models.PostCategory, error) {
	var categories []models.PostCategory
	statement := withContext(s.db, ctx).Order("sort_order ASC, id ASC")
	if parentID != nil {
		statement = statement.Where("parent_id = ?", *parentID)
	}
	if err := statement.Find(&categories).Error; err != nil {
		return nil, err
	}
	return categories, nil
}

func (s *PostCategoryStore) ListActive(ctx context.Context) ([]models.PostCategory, error) {
	var categories []models.PostCategory
	err := withContext(s.db, ctx).
		Where("is_active = ?", true).
		Order("sort_order ASC, id ASC").
		Find(&categories).Error
	return categories, err
}

func (s *PostCategoryStore) ListTree(ctx context.Context) ([]models.PostCategory, error) {
	var all []models.PostCategory
	if err := withContext(s.db, ctx).Order("sort_order ASC, id ASC").Find(&all).Error; err != nil {
		return nil, err
	}

	byID := make(map[uint]*models.PostCategory, len(all))
	for index := range all {
		byID[all[index].ID] = &all[index]
	}
	for index := range all {
		if all[index].ParentID != nil && *all[index].ParentID != 0 {
			if parent, exists := byID[*all[index].ParentID]; exists {
				parent.Children = append(parent.Children, all[index])
			}
		}
	}

	roots := make([]models.PostCategory, 0, len(all))
	for index := range all {
		if all[index].ParentID == nil || *all[index].ParentID == 0 {
			roots = append(roots, *byID[all[index].ID])
		}
	}
	return roots, nil
}

func (s *PostCategoryStore) GetByID(ctx context.Context, id uint) (*models.PostCategory, error) {
	var category models.PostCategory
	if err := withContext(s.db, ctx).First(&category, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &category, nil
}

func (s *PostCategoryStore) Create(ctx context.Context, category *models.PostCategory) error {
	return withContext(s.db, ctx).Create(category).Error
}

func (s *PostCategoryStore) Update(ctx context.Context, category *models.PostCategory) error {
	return withContext(s.db, ctx).Save(category).Error
}

func (s *PostCategoryStore) UpdateActive(ctx context.Context, id uint, active bool) error {
	return withContext(s.db, ctx).Model(&models.PostCategory{}).Where("id = ?", id).Update("is_active", active).Error
}

func (s *PostCategoryStore) Delete(ctx context.Context, id uint) error {
	return withContext(s.db, ctx).Delete(&models.PostCategory{}, id).Error
}

func (s *PostCategoryStore) CountBySlug(ctx context.Context, slug string, excludeID *uint) (int64, error) {
	var count int64
	statement := withContext(s.db, ctx).Model(&models.PostCategory{}).Where("slug = ?", slug)
	if excludeID != nil {
		statement = statement.Where("id != ?", *excludeID)
	}
	if err := statement.Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (s *PostCategoryStore) CountChildren(ctx context.Context, parentID uint) (int64, error) {
	var count int64
	err := withContext(s.db, ctx).Model(&models.PostCategory{}).Where("parent_id = ?", parentID).Count(&count).Error
	return count, err
}

func (s *PostCategoryStore) CountPostsByCategory(ctx context.Context, categoryID uint) (int64, error) {
	var count int64
	err := withContext(s.db, ctx).Model(&models.Post{}).Where("category_id = ?", categoryID).Count(&count).Error
	return count, err
}
