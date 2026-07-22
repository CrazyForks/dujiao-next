package siteconnectionwiring

import (
	"github.com/dujiao-next/internal/provider"
	siteconnectiontransport "github.com/dujiao-next/internal/transport/http/siteconnection"
)

func NewAdminHandler(c *provider.Container) *siteconnectiontransport.AdminHandler {
	return siteconnectiontransport.NewAdminHandler(
		c.SiteConnectionService,
		siteConnectionMarkupAdapter{svc: c.ProductMappingService},
	)
}
