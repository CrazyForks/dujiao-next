package affiliatewiring

import (
	"github.com/dujiao-next/internal/service"
	affiliatetransport "github.com/dujiao-next/internal/transport/http/affiliate"
)

// affiliateChannelUserAdapter 将 UserAuthService 适配为渠道推广身份开通端口。
type affiliateChannelUserAdapter struct {
	auth *service.UserAuthService
}

func (a affiliateChannelUserAdapter) ProvisionUserID(identity affiliatetransport.ChannelIdentity) (uint, error) {
	user, _, _, err := a.auth.ProvisionTelegramChannelIdentity(service.TelegramChannelIdentityInput{
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
		return 0, service.ErrNotFound
	}
	return user.ID, nil
}
