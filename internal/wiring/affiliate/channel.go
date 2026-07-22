package affiliatewiring

import (
	userauthapp "github.com/dujiao-next/internal/modules/identity/userauth/application"
	affiliatetransport "github.com/dujiao-next/internal/transport/http/affiliate"
)

// affiliateChannelUserAdapter 将 UserAuthService 适配为渠道推广身份开通端口。
type affiliateChannelUserAdapter struct {
	auth *userauthapp.Service
}

func (a affiliateChannelUserAdapter) ProvisionUserID(identity affiliatetransport.ChannelIdentity) (uint, error) {
	user, _, _, err := a.auth.ProvisionTelegramChannelIdentity(userauthapp.TelegramChannelIdentityInput{
		ChannelUserID: identity.ChannelUserID,
		Username:      identity.Username,
		FirstName:     identity.FirstName,
		LastName:      identity.LastName,
		AvatarURL:     identity.AvatarURL,
	})
	if err != nil {
		return 0, err
	}
	if user == nil {
		return 0, userauthapp.ErrNotFound
	}
	return user.ID, nil
}
