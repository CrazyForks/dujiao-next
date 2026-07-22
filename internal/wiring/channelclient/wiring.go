package channelclientwiring

import (
	"github.com/dujiao-next/internal/modules/channelclient"
)

// channelClientAdminAdapter 将 ChannelClient Service 适配为 transport 端口。
// ChannelClientResponse 已是 modules/channelclient.ClientDetail 别名，可直接透传。
type channelClientAdminAdapter struct {
	svc *channelclient.Service
}

func (a channelClientAdminAdapter) ListChannelClientDetails() ([]channelclient.ClientDetail, error) {
	return a.svc.ListChannelClientDetails()
}

func (a channelClientAdminAdapter) CreateChannelClient(name, channelType, description, botToken, callbackURL string) (*channelclient.ClientDetail, error) {
	return a.svc.CreateChannelClient(name, channelType, description, botToken, callbackURL)
}

func (a channelClientAdminAdapter) GetChannelClientDetail(id uint) (*channelclient.ClientDetail, error) {
	return a.svc.GetChannelClientDetail(id)
}

func (a channelClientAdminAdapter) UpdateChannelClientStatus(id uint, status int) error {
	return a.svc.UpdateChannelClientStatus(id, status)
}

func (a channelClientAdminAdapter) UpdateChannelClient(id uint, name, description string, botToken *string, callbackURL *string) (*channelclient.ClientDetail, error) {
	return a.svc.UpdateChannelClient(id, name, description, botToken, callbackURL)
}

func (a channelClientAdminAdapter) ResetChannelClientSecret(id uint) (*channelclient.ClientDetail, error) {
	return a.svc.ResetChannelClientSecret(id)
}

func (a channelClientAdminAdapter) DeleteChannelClient(id uint) error {
	return a.svc.DeleteChannelClient(id)
}
