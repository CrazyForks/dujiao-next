package content

import (
	"context"
	"io"
	"io/fs"
	"time"

	productdomain "github.com/dujiao-next/internal/modules/catalog/product/domain"

	"github.com/dujiao-next/internal/models"
)

// PostWriteUnitOfWork 保证文章本体与商品关联在同一个持久化事务内写入。
type PostWriteUnitOfWork interface {
	WithinPostWriteTransaction(
		ctx context.Context,
		operation func(posts PostStore, relations PostProductRelationStore) error,
	) error
}

// PostStore 持久化文章本体；商品关联由独立端口负责。
type PostStore interface {
	PostWriteUnitOfWork
	List(ctx context.Context, query PostQuery) ([]models.Post, int64, error)
	GetBySlug(ctx context.Context, slug string, onlyPublished bool) (*models.Post, error)
	GetByID(ctx context.Context, id string) (*models.Post, error)
	Create(ctx context.Context, post *models.Post) error
	Update(ctx context.Context, post *models.Post) error
	Delete(ctx context.Context, id string) error
	CountBySlug(ctx context.Context, slug string, excludeID *string) (int64, error)
}

// PostProductRelationStore 持久化文章与商品的有序关联。
type PostProductRelationStore interface {
	GetRelatedProductIDs(ctx context.Context, postID uint) ([]uint, error)
	SetRelatedProductIDs(ctx context.Context, postID uint, productIDs []uint) error
	ListRelatedProducts(ctx context.Context, postID uint) ([]productdomain.Product, error)
	ListPostsForProduct(ctx context.Context, productID uint, postType string, onlyPublished bool, limit int) ([]models.Post, error)
}

// PostCategoryStore 持久化文章分类和分类占用关系。
type PostCategoryStore interface {
	ListAll(ctx context.Context, parentID *uint) ([]models.PostCategory, error)
	ListActive(ctx context.Context) ([]models.PostCategory, error)
	ListTree(ctx context.Context) ([]models.PostCategory, error)
	GetByID(ctx context.Context, id uint) (*models.PostCategory, error)
	Create(ctx context.Context, category *models.PostCategory) error
	Update(ctx context.Context, category *models.PostCategory) error
	UpdateActive(ctx context.Context, id uint, active bool) error
	Delete(ctx context.Context, id uint) error
	CountBySlug(ctx context.Context, slug string, excludeID *uint) (int64, error)
	CountChildren(ctx context.Context, parentID uint) (int64, error)
	CountPostsByCategory(ctx context.Context, categoryID uint) (int64, error)
}

// BannerStore 持久化后台 Banner，并提供按时间窗口读取公开 Banner 的查询。
type BannerStore interface {
	List(ctx context.Context, query BannerQuery) ([]models.Banner, int64, error)
	ListValidByPosition(ctx context.Context, position string, limit int, now time.Time) ([]models.Banner, error)
	GetByID(ctx context.Context, id string) (*models.Banner, error)
	Create(ctx context.Context, banner *models.Banner) error
	Update(ctx context.Context, banner *models.Banner) error
	Delete(ctx context.Context, id string) error
}

// MediaStore 持久化素材元数据，不负责物理文件操作。
type MediaStore interface {
	List(ctx context.Context, query MediaQuery) ([]models.Media, int64, error)
	GetByID(ctx context.Context, id uint) (*models.Media, error)
	GetByPath(ctx context.Context, path string) (*models.Media, error)
	Create(ctx context.Context, media *models.Media) error
	Update(ctx context.Context, media *models.Media) error
	Delete(ctx context.Context, id uint) error
}

// FileStore 描述 Media 用例真实需要的本地文件能力。
type FileStore interface {
	Stat(name string) (fs.FileInfo, error)
	Open(name string) (io.ReadCloser, error)
	Remove(name string) error
}

// Clock 让发布时间和 Banner 时间窗口可以在测试中固定。
type Clock interface {
	Now() time.Time
}

// WarningLogger 只暴露 Media 用例记录非致命文件副作用失败所需的能力。
type WarningLogger interface {
	Warnw(message string, keysAndValues ...interface{})
}
