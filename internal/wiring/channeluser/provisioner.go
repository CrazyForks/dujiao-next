package channeluserwiring

import "github.com/dujiao-next/internal/service"

type SimpleProvisioner struct {
	auth *service.UserAuthService
}

func NewSimpleProvisioner(auth *service.UserAuthService) SimpleProvisioner {
	return SimpleProvisioner{auth: auth}
}

func (p SimpleProvisioner) ProvisionUserID(channelUserID string) (uint, error) {
	user, _, _, err := p.auth.ProvisionTelegramChannelIdentity(service.TelegramChannelIdentityInput{
		ChannelUserID: channelUserID,
	})
	if err != nil {
		return 0, err
	}
	if user == nil {
		return 0, service.ErrNotFound
	}
	return user.ID, nil
}
