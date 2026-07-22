package sitemapwiring

import (
	"github.com/dujiao-next/internal/provider"
	sitemaptransport "github.com/dujiao-next/internal/transport/http/sitemap"
)

func NewHandler(c *provider.Container) *sitemaptransport.Handler {
	return sitemaptransport.NewHandler(
		c.SitemapService,
		sitemapSiteBrandAdapter{settings: c.SettingService},
	)
}
