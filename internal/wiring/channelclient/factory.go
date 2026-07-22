package channelclientwiring

import (
	"github.com/dujiao-next/internal/provider"
	channelclienttransport "github.com/dujiao-next/internal/transport/http/channelclient"
)

func NewAdminHandler(c *provider.Container) *channelclienttransport.AdminHandler {
	return channelclienttransport.NewAdminHandler(channelClientAdminAdapter{svc: c.ChannelClientService})
}
