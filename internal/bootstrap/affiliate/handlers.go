package affiliatebootstrap

import (
	affiliatetransport "github.com/dujiao-next/internal/modules/affiliate/transport/http"
	"github.com/dujiao-next/internal/provider"
)

func NewStorefrontHandler(c *provider.Container) *affiliatetransport.Handler {
	return affiliatetransport.NewHandler(c.AffiliateService)
}

func NewAdminHandler(c *provider.Container) *affiliatetransport.AdminHandler {
	return affiliatetransport.NewAdminHandler(c.AffiliateService)
}

func NewChannelHandler(c *provider.Container) *affiliatetransport.ChannelHandler {
	return affiliatetransport.NewChannelHandler(
		c.AffiliateService,
		affiliateChannelUserAdapter{auth: c.UserAuthService},
		c.SettingService,
	)
}
