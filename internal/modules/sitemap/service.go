package sitemap

import (
	"context"
	"encoding/xml"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/dujiao-next/internal/cache"
	"github.com/dujiao-next/internal/modules/catalog"
	catalogproduct "github.com/dujiao-next/internal/modules/catalog/product"
)

// SitemapPost 是生成 Sitemap 所需的最小文章投影。
type SitemapPost struct {
	Slug        string
	CreatedAt   time.Time
	PublishedAt *time.Time
}

// PublishedPostReader 由 Sitemap 消费方拥有，只读取可索引文章。
type PublishedPostReader interface {
	ListPublishedPosts(ctx context.Context, limit int) ([]SitemapPost, error)
}

// PublishedPostReaderFunc 将装配层函数适配为 PublishedPostReader。
type PublishedPostReaderFunc func(ctx context.Context, limit int) ([]SitemapPost, error)

// ListPublishedPosts 实现 PublishedPostReader。
func (f PublishedPostReaderFunc) ListPublishedPosts(ctx context.Context, limit int) ([]SitemapPost, error) {
	return f(ctx, limit)
}

// Service 生成 sitemap.xml / robots.txt 内容。
type Service struct {
	productRepo  catalogproduct.Repository
	categoryRepo catalog.CategoryRepository
	posts        PublishedPostReader
}

// NewService 创建 sitemap 服务。
func NewService(
	productRepo catalogproduct.Repository,
	categoryRepo catalog.CategoryRepository,
	posts PublishedPostReader,
) (*Service, error) {
	if productRepo == nil || categoryRepo == nil || posts == nil {
		return nil, fmt.Errorf("sitemap: required dependency is nil")
	}
	return &Service{
		productRepo:  productRepo,
		categoryRepo: categoryRepo,
		posts:        posts,
	}, nil
}

const (
	sitemapCacheTTL    = 5 * time.Minute
	sitemapCachePrefix = "sitemap:xml:"
	sitemapMaxFetch    = 50000 // 单次拉取上限，避免极端数据量打爆内存
)

// Generate 生成 sitemap.xml 内容；baseURL 必须是不带尾斜杠的站点根（如 https://example.com）
func (s *Service) Generate(ctx context.Context, baseURL string) (string, error) {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if baseURL == "" {
		return "", fmt.Errorf("sitemap: empty base url")
	}

	cacheKey := sitemapCachePrefix + baseURL
	if cached, err := cache.GetString(ctx, cacheKey); err == nil && cached != "" {
		return cached, nil
	}

	entries, err := s.collectURLs(ctx, baseURL)
	if err != nil {
		return "", err
	}

	xmlStr, err := renderSitemapXML(entries)
	if err != nil {
		return "", err
	}

	_ = cache.SetString(ctx, cacheKey, xmlStr, sitemapCacheTTL)
	return xmlStr, nil
}

// GenerateRobots 生成 robots.txt 内容
func (s *Service) GenerateRobots(baseURL string) string {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	var b strings.Builder
	b.WriteString("User-agent: *\n")
	b.WriteString("Disallow: /api/\n")
	b.WriteString("Disallow: /admin/\n")
	b.WriteString("Disallow: /me/\n")
	b.WriteString("Disallow: /cart\n")
	b.WriteString("Disallow: /checkout\n")
	b.WriteString("Disallow: /pay\n")
	b.WriteString("Disallow: /orders/\n")
	b.WriteString("Disallow: /recharge-orders/\n")
	b.WriteString("Disallow: /guest/\n")
	b.WriteString("Disallow: /auth/\n")
	if baseURL != "" {
		b.WriteString("\n")
		b.WriteString("Sitemap: ")
		b.WriteString(baseURL)
		b.WriteString("/sitemap.xml\n")
	}
	return b.String()
}

// urlEntry sitemap.xml 中的单条 URL
type urlEntry struct {
	XMLName    xml.Name `xml:"url"`
	Loc        string   `xml:"loc"`
	LastMod    string   `xml:"lastmod,omitempty"`
	ChangeFreq string   `xml:"changefreq,omitempty"`
	Priority   string   `xml:"priority,omitempty"`
}

type urlSet struct {
	XMLName xml.Name   `xml:"urlset"`
	Xmlns   string     `xml:"xmlns,attr"`
	URLs    []urlEntry `xml:"url"`
}

func (s *Service) collectURLs(ctx context.Context, baseURL string) ([]urlEntry, error) {
	now := time.Now().UTC().Format("2006-01-02")
	entries := make([]urlEntry, 0, 64)

	// 1. 静态页面
	staticPages := []struct {
		Path       string
		ChangeFreq string
		Priority   string
	}{
		{"/", "daily", "1.0"},
		{"/products", "daily", "0.9"},
		{"/blog", "weekly", "0.6"},
		{"/notice", "weekly", "0.5"},
		{"/about", "monthly", "0.3"},
		{"/terms", "yearly", "0.2"},
		{"/privacy", "yearly", "0.2"},
	}
	for _, p := range staticPages {
		entries = append(entries, urlEntry{
			Loc:        baseURL + p.Path,
			LastMod:    now,
			ChangeFreq: p.ChangeFreq,
			Priority:   p.Priority,
		})
	}

	// 2. 启用的分类
	categories, err := s.categoryRepo.ListActive()
	if err != nil {
		return nil, fmt.Errorf("sitemap: list categories: %w", err)
	}
	for _, cat := range categories {
		entries = append(entries, urlEntry{
			Loc:        baseURL + "/categories/" + url.PathEscape(cat.Slug),
			LastMod:    cat.CreatedAt.UTC().Format("2006-01-02"),
			ChangeFreq: "weekly",
			Priority:   "0.7",
		})
	}

	// 3. 上架的商品（OnlyActive 已含分类启用过滤）
	products, _, err := s.productRepo.List(catalogproduct.ListFilter{
		Page:       1,
		PageSize:   sitemapMaxFetch,
		OnlyActive: true,
	})
	if err != nil {
		return nil, fmt.Errorf("sitemap: list products: %w", err)
	}
	for _, p := range products {
		entries = append(entries, urlEntry{
			Loc:        baseURL + "/products/" + url.PathEscape(p.Slug),
			LastMod:    p.UpdatedAt.UTC().Format("2006-01-02"),
			ChangeFreq: "daily",
			Priority:   "0.8",
		})
	}

	// 4. 已发布的博客 / 公告
	posts, err := s.posts.ListPublishedPosts(ctx, sitemapMaxFetch)
	if err != nil {
		return nil, fmt.Errorf("sitemap: list posts: %w", err)
	}
	for _, post := range posts {
		lastmod := post.CreatedAt
		if post.PublishedAt != nil {
			lastmod = *post.PublishedAt
		}
		// blog 与 notice 共用 /blog/:slug 详情页（user 前台 Notice.vue 跳转到 /blog/{slug}）
		entries = append(entries, urlEntry{
			Loc:        baseURL + "/blog/" + url.PathEscape(post.Slug),
			LastMod:    lastmod.UTC().Format("2006-01-02"),
			ChangeFreq: "monthly",
			Priority:   "0.5",
		})
	}

	return entries, nil
}

func renderSitemapXML(entries []urlEntry) (string, error) {
	set := urlSet{
		Xmlns: "http://www.sitemaps.org/schemas/sitemap/0.9",
		URLs:  entries,
	}
	body, err := xml.MarshalIndent(set, "", "  ")
	if err != nil {
		return "", err
	}
	return xml.Header + string(body) + "\n", nil
}
