package settingsbootstrap

import (
	"github.com/dujiao-next/internal/config"
	settingstransport "github.com/dujiao-next/internal/modules/settings/transport/http"
	"github.com/dujiao-next/internal/provider"
)

func NewSMTPHandler(c *provider.Container, cfg *config.Config) *settingstransport.SMTPHandler {
	return settingstransport.NewSMTPHandler(settingsSMTPAdapter{
		settings: c.SettingService, cfg: cfg, email: c.EmailService,
	})
}

func NewCaptchaHandler(c *provider.Container, cfg *config.Config) *settingstransport.CaptchaHandler {
	return settingstransport.NewCaptchaHandler(settingsCaptchaAdapter{
		settings: c.SettingService, cfg: cfg, captcha: c.CaptchaService,
	})
}

func NewTelegramAuthHandler(c *provider.Container, cfg *config.Config) *settingstransport.TelegramAuthHandler {
	return settingstransport.NewTelegramAuthHandler(settingsTelegramAuthAdapter{
		settings: c.SettingService, cfg: cfg, telegramAuth: c.TelegramAuthService,
	})
}
