package userauthwiring

import (
	"github.com/dujiao-next/internal/provider"
	userauthtransport "github.com/dujiao-next/internal/transport/http/userauth"
	captchawiring "github.com/dujiao-next/internal/wiring/captcha"
)

// Handlers contains the complete user authentication HTTP entrypoint set.
type Handlers struct {
	Profile      *userauthtransport.UserProfileHandler
	Email        *userauthtransport.UserEmailHandler
	Password     *userauthtransport.UserPasswordHandler
	Verify       *userauthtransport.UserVerifyHandler
	Login        *userauthtransport.UserLoginHandler
	TwoFA        *userauthtransport.User2FAHandler
	TelegramOIDC *userauthtransport.UserTelegramOIDCHandler
	Telegram     *userauthtransport.UserTelegramHandler
}

// New assembles user authentication transports at the application boundary.
func New(c *provider.Container) Handlers {
	verify := userVerifyTransportAdapter{auth: c.UserAuthService, settings: c.SettingService}
	login := userLoginTransportAdapter{auth: c.UserAuthService, settings: c.SettingService}
	recorder := userLoginRecorderAdapter{logs: c.UserLoginLogService}
	captcha := captchawiring.NewVerifier(c.CaptchaService)

	return Handlers{
		Profile: userauthtransport.NewUserProfileHandler(
			userProfileTransportAdapter{service: c.UserAuthService},
		),
		Email: userauthtransport.NewUserEmailHandler(
			userEmailTransportAdapter{service: c.UserAuthService},
		),
		Password: userauthtransport.NewUserPasswordHandler(userPasswordTransportAdapter{
			auth: c.UserAuthService, settings: c.SettingService,
		}),
		Verify: userauthtransport.NewUserVerifyHandler(
			verify,
			captcha,
			verify,
		),
		Login: userauthtransport.NewUserLoginHandler(
			login,
			login,
			captcha,
			recorder,
		),
		TwoFA: userauthtransport.NewUser2FAHandler(
			user2FATOTPTransportAdapter{totp: c.UserTOTPService},
			user2FAAuthTransportAdapter{auth: c.UserAuthService, users: c.UserRepo},
			user2FAChallengeStoreAdapter{},
			recorder,
		),
		TelegramOIDC: userauthtransport.NewUserTelegramOIDCHandler(
			userTelegramOIDCTransportAdapter{auth: c.UserAuthService},
			recorder,
		),
		Telegram: userauthtransport.NewUserTelegramHandler(
			userTelegramTransportAdapter{auth: c.UserAuthService},
			recorder,
		),
	}
}
