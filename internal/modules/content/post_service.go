package content

import (
	"context"

	"github.com/dujiao-next/internal/constants"
	"github.com/dujiao-next/internal/models"
)

// CreatePostInput 描述文章创建和更新所需字段。
type CreatePostInput struct {
	Slug        string
	Type        string
	TitleJSON   map[string]interface{}
	SummaryJSON map[string]interface{}
	ContentJSON map[string]interface{}
	Thumbnail   string
	IsPublished *bool
	ProductIDs  *[]uint
	CategoryID  *uint
}

// PublicPostQuery 描述公开文章列表查询。
type PublicPostQuery struct {
	Type     string
	Search   string
	Page     int
	PageSize int
}

// AdminPostQuery 描述后台文章列表查询。
type AdminPostQuery struct {
	Type     string
	Search   string
	Page     int
	PageSize int
}

// PostService 实现文章用例。
type PostService struct {
	posts      PostStore
	relations  PostProductRelationStore
	categories PostCategoryStore
	clock      Clock
}

// NewPostService 创建文章用例服务。
func NewPostService(posts PostStore, relations PostProductRelationStore, categories PostCategoryStore, clock Clock) *PostService {
	if clock == nil {
		clock = SystemClock{}
	}
	return &PostService{
		posts:      posts,
		relations:  relations,
		categories: categories,
		clock:      clock,
	}
}

// ListPublic 获取公开文章列表。
func (s *PostService) ListPublic(ctx context.Context, query PublicPostQuery) ([]models.Post, int64, error) {
	return s.posts.List(ctx, PostQuery{
		Page:          query.Page,
		PageSize:      query.PageSize,
		Type:          query.Type,
		Search:        query.Search,
		OnlyPublished: true,
		Order:         PostOrderPublishedDesc,
	})
}

// GetPublicBySlug 获取公开文章详情。
func (s *PostService) GetPublicBySlug(ctx context.Context, slug string) (*models.Post, error) {
	post, err := s.posts.GetBySlug(ctx, slug, true)
	if err != nil {
		return nil, err
	}
	if post == nil {
		return nil, ErrNotFound
	}
	return post, nil
}

// ListAdmin 获取后台文章列表。
func (s *PostService) ListAdmin(ctx context.Context, query AdminPostQuery) ([]models.Post, int64, error) {
	return s.posts.List(ctx, PostQuery{
		Page:     query.Page,
		PageSize: query.PageSize,
		Type:     query.Type,
		Search:   query.Search,
		Order:    PostOrderCreatedDesc,
	})
}

// Create 创建文章。
func (s *PostService) Create(ctx context.Context, input CreatePostInput) (*models.Post, error) {
	if !isAllowedPostType(input.Type) {
		return nil, ErrInvalidPostType
	}
	categoryID := normalizePostCategoryID(input.CategoryID)
	if err := s.validateCategoryAssignment(ctx, input.Type, categoryID, nil); err != nil {
		return nil, err
	}

	count, err := s.posts.CountBySlug(ctx, input.Slug, nil)
	if err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, ErrSlugExists
	}

	isPublished := false
	if input.IsPublished != nil {
		isPublished = *input.IsPublished
	}
	post := models.Post{
		Slug:        input.Slug,
		Type:        input.Type,
		TitleJSON:   models.JSON(input.TitleJSON),
		SummaryJSON: models.JSON(input.SummaryJSON),
		ContentJSON: models.JSON(input.ContentJSON),
		Thumbnail:   input.Thumbnail,
		IsPublished: isPublished,
		CategoryID:  categoryID,
	}
	if isPublished {
		now := s.clock.Now()
		post.PublishedAt = &now
	}

	if err := s.posts.WithinPostWriteTransaction(ctx, func(posts PostStore, relations PostProductRelationStore) error {
		if err := posts.Create(ctx, &post); err != nil {
			return err
		}
		if input.ProductIDs == nil {
			return nil
		}
		return relations.SetRelatedProductIDs(ctx, post.ID, *input.ProductIDs)
	}); err != nil {
		return nil, err
	}
	return &post, nil
}

// Update 更新文章。
func (s *PostService) Update(ctx context.Context, id string, input CreatePostInput) (*models.Post, error) {
	if !isAllowedPostType(input.Type) {
		return nil, ErrInvalidPostType
	}

	post, err := s.posts.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if post == nil {
		return nil, ErrNotFound
	}
	categoryID := normalizePostCategoryID(input.CategoryID)
	if err := s.validateCategoryAssignment(ctx, input.Type, categoryID, post.CategoryID); err != nil {
		return nil, err
	}

	count, err := s.posts.CountBySlug(ctx, input.Slug, &id)
	if err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, ErrSlugExists
	}

	post.Slug = input.Slug
	post.Type = input.Type
	post.TitleJSON = models.JSON(input.TitleJSON)
	post.SummaryJSON = models.JSON(input.SummaryJSON)
	post.ContentJSON = models.JSON(input.ContentJSON)
	post.Thumbnail = input.Thumbnail
	post.CategoryID = categoryID
	if input.IsPublished != nil {
		wasPublished := post.IsPublished
		post.IsPublished = *input.IsPublished
		if *input.IsPublished && !wasPublished && post.PublishedAt == nil {
			now := s.clock.Now()
			post.PublishedAt = &now
		}
	}

	if err := s.posts.WithinPostWriteTransaction(ctx, func(posts PostStore, relations PostProductRelationStore) error {
		if err := posts.Update(ctx, post); err != nil {
			return err
		}
		if input.ProductIDs == nil {
			return nil
		}
		return relations.SetRelatedProductIDs(ctx, post.ID, *input.ProductIDs)
	}); err != nil {
		return nil, err
	}
	return post, nil
}

// GetRelatedProductIDs 获取文章关联商品 ID 列表。
func (s *PostService) GetRelatedProductIDs(ctx context.Context, postID uint) ([]uint, error) {
	return s.relations.GetRelatedProductIDs(ctx, postID)
}

// ListRelatedProducts 获取文章关联商品列表。
func (s *PostService) ListRelatedProducts(ctx context.Context, postID uint) ([]models.Product, error) {
	return s.relations.ListRelatedProducts(ctx, postID)
}

// ListPostsForProduct 获取与商品关联的已发布博客列表。
func (s *PostService) ListPostsForProduct(ctx context.Context, productID uint, limit int) ([]models.Post, error) {
	return s.relations.ListPostsForProduct(ctx, productID, constants.PostTypeBlog, true, limit)
}

// Delete 删除文章。
func (s *PostService) Delete(ctx context.Context, id string) error {
	post, err := s.posts.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if post == nil {
		return ErrNotFound
	}
	return s.posts.Delete(ctx, id)
}

func isAllowedPostType(postType string) bool {
	return postType == constants.PostTypeBlog || postType == constants.PostTypeNotice
}

func (s *PostService) validateCategoryAssignment(ctx context.Context, postType string, categoryID, currentCategoryID *uint) error {
	if postType == constants.PostTypeNotice {
		if categoryID != nil && *categoryID > 0 {
			return ErrPostNoticeCategoryUnsupported
		}
		return nil
	}
	if categoryID == nil || *categoryID == 0 {
		return nil
	}
	if s.categories == nil {
		return ErrPostCategoryInvalid
	}

	category, err := s.categories.GetByID(ctx, *categoryID)
	if err != nil {
		return err
	}
	if category == nil {
		return ErrPostCategoryInvalid
	}
	if !category.IsActive && !sameOptionalUint(currentCategoryID, categoryID) {
		return ErrPostCategoryInvalid
	}

	childCount, err := s.categories.CountChildren(ctx, *categoryID)
	if err != nil {
		return err
	}
	if childCount > 0 && !sameOptionalUint(currentCategoryID, categoryID) {
		return ErrPostCategoryInvalid
	}
	return nil
}

func normalizePostCategoryID(categoryID *uint) *uint {
	if categoryID != nil && *categoryID == 0 {
		return nil
	}
	return categoryID
}

func sameOptionalUint(left, right *uint) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}
	return *left == *right
}
