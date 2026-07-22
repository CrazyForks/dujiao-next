package settingswiring

import (
	"github.com/dujiao-next/internal/config"
	"github.com/dujiao-next/internal/modules/captcha"
	settingsmodule "github.com/dujiao-next/internal/modules/settings"
	settingsapp "github.com/dujiao-next/internal/modules/settings/application"
)

type settingsCaptchaAdapter struct {
	settings *settingsapp.Service
	cfg      *config.Config
	captcha  *captcha.Service
}

func (a settingsCaptchaAdapter) GetCaptchaSetting() (settingsmodule.CaptchaSetting, error) {
	return a.settings.GetCaptchaSetting(a.cfg.Captcha)
}

func (a settingsCaptchaAdapter) PatchCaptchaSetting(patch settingsmodule.CaptchaSettingPatch) (settingsmodule.CaptchaSetting, error) {
	return a.settings.PatchCaptchaSetting(a.cfg.Captcha, patch)
}

func (a settingsCaptchaAdapter) ApplyRuntime(setting settingsmodule.CaptchaSetting) {
	a.cfg.Captcha = settingsmodule.CaptchaSettingToConfig(setting)
	if a.captcha != nil {
		a.captcha.SetDefaultConfig(a.cfg.Captcha)
		a.captcha.InvalidateCache()
	}
}
