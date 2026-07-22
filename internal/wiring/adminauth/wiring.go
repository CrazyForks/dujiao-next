package adminauthwiring

import (
	"github.com/dujiao-next/internal/provider"
	adminauthtransport "github.com/dujiao-next/internal/transport/http/adminauth"
	captchawiring "github.com/dujiao-next/internal/wiring/captcha"
)

type Handlers struct {
	Login     *adminauthtransport.AdminLoginHandler
	TwoFA     *adminauthtransport.Admin2FAHandler
	UserTwoFA *adminauthtransport.AdminUser2FAHandler
}

func New(c *provider.Container) Handlers {
	recorder := adminLoginRecorderAdapter{logs: c.AdminLoginLogRepo}
	return Handlers{
		Login: adminauthtransport.NewAdminLoginHandler(
			adminLoginAuthTransportAdapter{auth: c.AuthService},
			captchawiring.NewVerifier(c.CaptchaService),
			recorder,
		),
		TwoFA: adminauthtransport.NewAdmin2FAHandler(
			admin2FATOTPTransportAdapter{totp: c.TOTPService},
			admin2FAAuthTransportAdapter{auth: c.AuthService},
			admin2FAChallengeStoreAdapter{},
			recorder,
		),
		UserTwoFA: adminauthtransport.NewAdminUser2FAHandler(
			adminUser2FATransportAdapter{totp: c.UserTOTPService},
		),
	}
}
