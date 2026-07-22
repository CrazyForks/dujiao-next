package settingswiring

import (
	"github.com/dujiao-next/internal/config"
	settingsmodule "github.com/dujiao-next/internal/modules/settings"
	"github.com/dujiao-next/internal/service"
)

type settingsTelegramAuthAdapter struct {
	settings     *service.SettingService
	cfg          *config.Config
	telegramAuth *service.TelegramAuthService
}

func (a settingsTelegramAuthAdapter) GetTelegramAuthSetting() (settingsmodule.TelegramAuthSetting, error) {
	return a.settings.GetTelegramAuthSetting(a.cfg.TelegramAuth)
}

func (a settingsTelegramAuthAdapter) PatchTelegramAuthSetting(patch settingsmodule.TelegramAuthSettingPatch) (settingsmodule.TelegramAuthSetting, error) {
	return a.settings.PatchTelegramAuthSetting(a.cfg.TelegramAuth, patch)
}

func (a settingsTelegramAuthAdapter) ApplyRuntime(setting settingsmodule.TelegramAuthSetting) {
	a.cfg.TelegramAuth = settingsmodule.TelegramAuthSettingToConfig(setting)
	if a.telegramAuth != nil {
		a.telegramAuth.SetConfig(a.cfg.TelegramAuth)
	}
}
