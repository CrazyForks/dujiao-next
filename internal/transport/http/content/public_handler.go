package contenthttp

import (
	"context"
	"errors"
	"strconv"

	"github.com/dujiao-next/internal/constants"
	"github.com/dujiao-next/internal/dto"
	"github.com/dujiao-next/internal/http/handlers/shared"
	"github.com/dujiao-next/internal/http/response"
	"github.com/dujiao-next/internal/models"
	domaincontent "github.com/dujiao-next/internal/modules/content"
	"github.com/gin-gonic/gin"
)

// PublicPostQueries 是公开 Content Handler 实际需要的文章读取能力。
type PublicPostQueries interface {
	ListPublic(ctx context.Context, query domaincontent.PublicPostQuery) ([]models.Post, int64, error)
	GetPublicBySlug(ctx context.Context, slug string) (*models.Post, error)
	ListRelatedProducts(ctx context.Context, postID uint) ([]models.Product, error)
}

// PublicPostCategoryQueries 是公开 Handler 实际需要的分类读取能力。
type PublicPostCategoryQueries interface {
	ListActive(ctx context.Context) ([]models.PostCategory, error)
}

// PublicBannerQueries 是公开 Handler 实际需要的 Banner 读取能力。
type PublicBannerQueries interface {
	ListPublic(ctx context.Context, query domaincontent.PublicBannerQuery) ([]models.Banner, error)
}

// PublicHandler 处理 Content 公开 HTTP 接口，仅持有窄用例依赖。
type PublicHandler struct {
	posts      PublicPostQueries
	categories PublicPostCategoryQueries
	banners    PublicBannerQueries
}

// NewPublicHandler 创建公开 Content Handler。
func NewPublicHandler(posts PublicPostQueries, categories PublicPostCategoryQueries, banners PublicBannerQueries) *PublicHandler {
	return &PublicHandler{posts: posts, categories: categories, banners: banners}
}

// GetPosts 获取公开文章或公告列表。
func (h *PublicHandler) GetPosts(c *gin.Context) {
	page, pageSize := shared.ParsePagination(c)
	posts, total, err := h.posts.ListPublic(c.Request.Context(), domaincontent.PublicPostQuery{
		Type:     c.Query("type"),
		Search:   c.Query("search"),
		Page:     page,
		PageSize: pageSize,
	})
	if err != nil {
		shared.RespondError(c, response.CodeInternal, "error.post_fetch_failed", err)
		return
	}

	pagination := response.BuildPagination(page, pageSize, total)
	response.SuccessWithPage(c, dto.NewPostRespList(posts), pagination)
}

// GetPostBySlug 根据 slug 获取公开文章详情。
func (h *PublicHandler) GetPostBySlug(c *gin.Context) {
	requestContext := c.Request.Context()
	post, err := h.posts.GetPublicBySlug(requestContext, c.Param("slug"))
	if err != nil {
		if errors.Is(err, domaincontent.ErrNotFound) {
			shared.RespondError(c, response.CodeNotFound, "error.post_not_found", nil)
			return
		}
		shared.RespondError(c, response.CodeInternal, "error.post_fetch_failed", err)
		return
	}

	result := dto.NewPostResp(post)
	if post.Type == constants.PostTypeBlog {
		products, relatedErr := h.posts.ListRelatedProducts(requestContext, post.ID)
		if relatedErr == nil {
			result.RelatedProducts = dto.NewRelatedProductCardList(products)
		}
	}
	response.Success(c, result)
}

// GetPublicBanners 获取公开 Banner 列表。
func (h *PublicHandler) GetPublicBanners(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if limit <= 0 {
		limit = 10
	}
	if limit > 50 {
		limit = 50
	}

	banners, err := h.banners.ListPublic(c.Request.Context(), domaincontent.PublicBannerQuery{
		Position: c.DefaultQuery("position", constants.BannerPositionHomeHero),
		Limit:    limit,
	})
	if err != nil {
		shared.RespondError(c, response.CodeInternal, "error.banner_fetch_failed", err)
		return
	}
	response.Success(c, dto.NewBannerRespList(banners))
}

// GetPostCategories 获取公开文章分类平铺列表。
func (h *PublicHandler) GetPostCategories(c *gin.Context) {
	categories, err := h.categories.ListActive(c.Request.Context())
	if err != nil {
		shared.RespondError(c, response.CodeInternal, "error.post_category_fetch_failed", err)
		return
	}
	response.Success(c, newPostCategoryDTOs(categories))
}
