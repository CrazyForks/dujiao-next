package telegramwiring

import (
	"github.com/dujiao-next/internal/provider"
	telegramtransport "github.com/dujiao-next/internal/transport/http/telegram"
)

func NewChannelBotHandler(c *provider.Container) *telegramtransport.ChannelBotHandler {
	return telegramtransport.NewChannelBotHandler(
		c.SettingService,
		c.ChannelClientService,
	)
}
