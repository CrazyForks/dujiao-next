package affiliatewiring

import (
	"github.com/dujiao-next/internal/provider"
	affiliatetransport "github.com/dujiao-next/internal/transport/http/affiliate"
)

func NewStorefrontHandler(c *provider.Container) *affiliatetransport.Handler {
	return affiliatetransport.NewHandler(affiliateStorefrontAdapter{svc: c.AffiliateService})
}

func NewAdminHandler(c *provider.Container) *affiliatetransport.AdminHandler {
	return affiliatetransport.NewAdminHandler(affiliateAdminAdapter{svc: c.AffiliateService})
}

func NewChannelHandler(c *provider.Container) *affiliatetransport.ChannelHandler {
	return affiliatetransport.NewChannelHandler(
		affiliateStorefrontAdapter{svc: c.AffiliateService},
		affiliateChannelUserAdapter{auth: c.UserAuthService},
		c.SettingService,
	)
}
