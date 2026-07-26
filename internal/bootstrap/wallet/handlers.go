package walletbootstrap

import (
	wallettransport "github.com/dujiao-next/internal/modules/wallet/transport/http"
	"github.com/dujiao-next/internal/provider"
	channeluserwiring "github.com/dujiao-next/internal/wiring/channeluser"
)

type Handlers struct {
	User    *wallettransport.UserHandler
	Admin   *wallettransport.AdminHandler
	Channel *wallettransport.ChannelHandler
}

func New(c *provider.Container) Handlers {
	wallets := walletTransportAdapter{wallets: c.WalletService, payments: c.PaymentService}
	return Handlers{
		User: wallettransport.NewUserHandler(
			wallets, wallets, c.UserStore, c.SettingService,
		),
		Admin: wallettransport.NewAdminHandler(
			wallets, c.UserStore, c.PaymentChannelRepo, c.PaymentRepo, c.SettingService,
		),
		Channel: wallettransport.NewChannelHandler(
			wallets,
			wallets,
			channeluserwiring.NewSimpleProvisioner(c.UserAuthService),
			c.SettingService,
		),
	}
}
