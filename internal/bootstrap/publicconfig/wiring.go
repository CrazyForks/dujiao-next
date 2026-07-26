package publicconfigwiring

import (
	"github.com/dujiao-next/internal/provider"
	publicconfigtransport "github.com/dujiao-next/internal/modules/settings/transport/http/public"
)

func NewHandler(c *provider.Container) *publicconfigtransport.Handler {
	var captcha publicconfigtransport.CaptchaPublic
	if c.CaptchaService != nil {
		captcha = publicConfigCaptchaAdapter{svc: c.CaptchaService}
	}
	var telegram publicconfigtransport.TelegramAuthPublic
	if c.TelegramAuthService != nil {
		telegram = publicConfigTelegramAdapter{svc: c.TelegramAuthService}
	}
	var overlay publicconfigtransport.ResellerOverlay
	if c.ResellerSiteConfigService != nil {
		overlay = publicConfigResellerOverlayAdapter{svc: c.ResellerSiteConfigService}
	}
	fallback := publicconfigtransport.TelegramAuthFallback{}
	if c.Config != nil {
		fallback = publicconfigtransport.TelegramAuthFallback{
			Enabled:     c.Config.TelegramAuth.Enabled,
			BotUsername: c.Config.TelegramAuth.BotUsername,
			MiniAppURL:  c.Config.TelegramAuth.MiniAppURL,
		}
	}
	return publicconfigtransport.NewHandler(
		publicConfigCacheAdapter{},
		publicConfigSettingsAdapter{settings: c.SettingService, cfg: c.Config},
		publicConfigPaymentAdapter{payments: c.PaymentService},
		captcha,
		telegram,
		fallback,
		overlay,
	)
}
