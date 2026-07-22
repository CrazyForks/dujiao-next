package adproxywiring

import (
	"github.com/dujiao-next/internal/provider"
	adproxytransport "github.com/dujiao-next/internal/transport/http/adproxy"
)

func NewAdminHandler(c *provider.Container) *adproxytransport.AdminHandler {
	return adproxytransport.NewAdminHandler(adProxyAdminAdapter{svc: c.AdProxyService})
}
