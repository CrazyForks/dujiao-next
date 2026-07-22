package dto

import (
	"time"

	"github.com/dujiao-next/internal/models"
	"github.com/dujiao-next/internal/shared/jsonmap"
)

// PostResp 文章/公告公共响应
type PostResp struct {
	ID              uint                 `json:"id"`
	Slug            string               `json:"slug"`
	Type            string               `json:"type"`
	Title           jsonmap.JSON         `json:"title"`
	Summary         jsonmap.JSON         `json:"summary"`
	Content         jsonmap.JSON         `json:"content"`
	Thumbnail       string               `json:"thumbnail,omitempty"`
	PublishedAt     *time.Time           `json:"published_at"`
	RelatedProducts []RelatedProductCard `json:"related_products,omitempty"`
}

// RelatedProductCard 文章详情底部展示的关联商品轻量卡片
type RelatedProductCard struct {
	ID          uint         `json:"id"`
	Slug        string       `json:"slug"`
	Title       jsonmap.JSON `json:"title"`
	PriceAmount models.Money `json:"price_amount"`
	Image       string       `json:"image,omitempty"`
}

// NewRelatedProductCardList 将 Product 列表转为关联卡片列表
func NewRelatedProductCardList(products []models.Product) []RelatedProductCard {
	cards := make([]RelatedProductCard, 0, len(products))
	for i := range products {
		p := &products[i]
		if !p.IsActive {
			continue
		}
		card := RelatedProductCard{
			ID:          p.ID,
			Slug:        p.Slug,
			Title:       p.TitleJSON,
			PriceAmount: p.PriceAmount,
		}
		if len(p.Images) > 0 {
			card.Image = p.Images[0]
		}
		cards = append(cards, card)
	}
	return cards
}

// NewPostResp 从 models.Post 构造响应
func NewPostResp(p *models.Post) PostResp {
	return PostResp{
		ID:          p.ID,
		Slug:        p.Slug,
		Type:        p.Type,
		Title:       p.TitleJSON,
		Summary:     p.SummaryJSON,
		Content:     p.ContentJSON,
		Thumbnail:   p.Thumbnail,
		PublishedAt: p.PublishedAt,
	}
	// 排除：IsPublished(内部状态)、CreatedAt
}

// RelatedPostCard 商品详情底部展示的关联文章轻量卡片
type RelatedPostCard struct {
	ID          uint         `json:"id"`
	Slug        string       `json:"slug"`
	Type        string       `json:"type"`
	Title       jsonmap.JSON `json:"title"`
	Summary     jsonmap.JSON `json:"summary,omitempty"`
	Thumbnail   string       `json:"thumbnail,omitempty"`
	PublishedAt *time.Time   `json:"published_at"`
}

// NewRelatedPostCardList 将 Post 列表转为关联文章卡片
func NewRelatedPostCardList(posts []models.Post) []RelatedPostCard {
	cards := make([]RelatedPostCard, 0, len(posts))
	for i := range posts {
		p := &posts[i]
		cards = append(cards, RelatedPostCard{
			ID:          p.ID,
			Slug:        p.Slug,
			Type:        p.Type,
			Title:       p.TitleJSON,
			Summary:     p.SummaryJSON,
			Thumbnail:   p.Thumbnail,
			PublishedAt: p.PublishedAt,
		})
	}
	return cards
}

// NewPostRespList 批量转换文章列表
func NewPostRespList(posts []models.Post) []PostResp {
	result := make([]PostResp, 0, len(posts))
	for i := range posts {
		result = append(result, NewPostResp(&posts[i]))
	}
	return result
}
