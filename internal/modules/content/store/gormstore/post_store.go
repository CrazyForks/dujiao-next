package gormstore

import (
	"context"
	"errors"
	"strings"

	"github.com/dujiao-next/internal/models"
	"github.com/dujiao-next/internal/modules/content"
	"gorm.io/gorm"
)

// PostStore 使用 GORM 持久化文章和文章商品关联。
type PostStore struct {
	db *gorm.DB
}

var (
	_ content.PostStore                = (*PostStore)(nil)
	_ content.PostProductRelationStore = (*PostStore)(nil)
)

// NewPostStore 创建文章持久化适配器。
func NewPostStore(db *gorm.DB) *PostStore {
	return &PostStore{db: db}
}

// WithinPostWriteTransaction 在同一个数据库事务中提供文章与商品关联 Store。
func (s *PostStore) WithinPostWriteTransaction(
	ctx context.Context,
	operation func(content.PostStore, content.PostProductRelationStore) error,
) error {
	return withContext(s.db, ctx).Transaction(func(tx *gorm.DB) error {
		transactionStore := NewPostStore(tx)
		return operation(transactionStore, transactionStore)
	})
}

func (s *PostStore) List(ctx context.Context, query content.PostQuery) ([]models.Post, int64, error) {
	var posts []models.Post
	db := withContext(s.db, ctx)
	statement := db.Model(&models.Post{})

	if query.OnlyPublished {
		statement = statement.Where("is_published = ?", true)
	}
	if query.Type != "" {
		statement = statement.Where("type = ?", query.Type)
	}
	if search := strings.TrimSpace(query.Search); search != "" {
		like := "%" + search + "%"
		condition, argCount := buildLocalizedLikeCondition(db, []string{"slug"}, []string{"title_json"})
		statement = statement.Where(condition, repeatLikeArgs(like, argCount)...)
	}

	var total int64
	if err := statement.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	statement = applyPagination(statement, query.Page, query.PageSize)
	orderBy := "created_at DESC"
	if query.Order == content.PostOrderPublishedDesc {
		orderBy = "published_at DESC, created_at DESC"
	}
	if err := statement.Order(orderBy).Find(&posts).Error; err != nil {
		return nil, 0, err
	}
	return posts, total, nil
}

func (s *PostStore) GetBySlug(ctx context.Context, slug string, onlyPublished bool) (*models.Post, error) {
	statement := withContext(s.db, ctx).Where("slug = ?", slug)
	if onlyPublished {
		statement = statement.Where("is_published = ?", true)
	}

	var post models.Post
	if err := statement.First(&post).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &post, nil
}

func (s *PostStore) GetByID(ctx context.Context, id string) (*models.Post, error) {
	var post models.Post
	if err := withContext(s.db, ctx).First(&post, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &post, nil
}

func (s *PostStore) Create(ctx context.Context, post *models.Post) error {
	return withContext(s.db, ctx).Create(post).Error
}

func (s *PostStore) Update(ctx context.Context, post *models.Post) error {
	return withContext(s.db, ctx).Save(post).Error
}

func (s *PostStore) Delete(ctx context.Context, id string) error {
	return withContext(s.db, ctx).Delete(&models.Post{}, id).Error
}

func (s *PostStore) CountBySlug(ctx context.Context, slug string, excludeID *string) (int64, error) {
	var count int64
	statement := withContext(s.db, ctx).Model(&models.Post{}).Where("slug = ?", slug)
	if excludeID != nil {
		statement = statement.Where("id != ?", *excludeID)
	}
	if err := statement.Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (s *PostStore) GetRelatedProductIDs(ctx context.Context, postID uint) ([]uint, error) {
	var ids []uint
	err := withContext(s.db, ctx).Model(&models.PostProduct{}).
		Where("post_id = ?", postID).
		Order("sort ASC, id ASC").
		Pluck("product_id", &ids).Error
	return ids, err
}

func (s *PostStore) SetRelatedProductIDs(ctx context.Context, postID uint, productIDs []uint) error {
	return withContext(s.db, ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("post_id = ?", postID).Delete(&models.PostProduct{}).Error; err != nil {
			return err
		}
		if len(productIDs) == 0 {
			return nil
		}

		seen := make(map[uint]struct{}, len(productIDs))
		records := make([]models.PostProduct, 0, len(productIDs))
		for index, productID := range productIDs {
			if productID == 0 {
				continue
			}
			if _, exists := seen[productID]; exists {
				continue
			}
			seen[productID] = struct{}{}
			records = append(records, models.PostProduct{
				PostID:    postID,
				ProductID: productID,
				Sort:      index,
			})
		}
		if len(records) == 0 {
			return nil
		}
		return tx.Create(&records).Error
	})
}

func (s *PostStore) ListRelatedProducts(ctx context.Context, postID uint) ([]models.Product, error) {
	var products []models.Product
	err := withContext(s.db, ctx).
		Joins("INNER JOIN post_products pp ON pp.product_id = products.id").
		Where("pp.post_id = ?", postID).
		Order("pp.sort ASC, pp.id ASC").
		Find(&products).Error
	return products, err
}

func (s *PostStore) ListPostsForProduct(ctx context.Context, productID uint, postType string, onlyPublished bool, limit int) ([]models.Post, error) {
	var posts []models.Post
	statement := withContext(s.db, ctx).
		Joins("INNER JOIN post_products pp ON pp.post_id = posts.id").
		Where("pp.product_id = ?", productID)
	if postType != "" {
		statement = statement.Where("posts.type = ?", postType)
	}
	if onlyPublished {
		statement = statement.Where("posts.is_published = ?", true)
	}
	if limit > 0 {
		statement = statement.Limit(limit)
	}
	if err := statement.Order("pp.sort ASC, pp.id ASC").Find(&posts).Error; err != nil {
		return nil, err
	}
	return posts, nil
}
