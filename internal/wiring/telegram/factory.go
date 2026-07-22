package telegramwiring

import (
	"github.com/dujiao-next/internal/provider"
	telegramtransport "github.com/dujiao-next/internal/transport/http/telegram"
)

func NewAdminBroadcastHandler(c *provider.Container) *telegramtransport.AdminBroadcastHandler {
	return telegramtransport.NewAdminBroadcastHandler(telegramBroadcastAdapter{svc: c.TelegramBroadcastService})
}

func NewChannelBotHandler(c *provider.Container) *telegramtransport.ChannelBotHandler {
	return telegramtransport.NewChannelBotHandler(
		c.SettingService,
		telegramChannelBotTokenAdapter{svc: c.ChannelClientService},
	)
}
