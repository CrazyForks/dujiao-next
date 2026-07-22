package settingswiring

import (
	"errors"

	settingsapp "github.com/dujiao-next/internal/modules/settings/application"

	"github.com/dujiao-next/internal/config"
	settingsmodule "github.com/dujiao-next/internal/modules/settings"
	"github.com/dujiao-next/internal/service"
	settingstransport "github.com/dujiao-next/internal/transport/http/settings"
)

type settingsSMTPAdapter struct {
	settings *settingsapp.Service
	cfg      *config.Config
	email    *service.EmailService
}

func (a settingsSMTPAdapter) GetSMTPSetting() (settingsmodule.SMTPSetting, error) {
	return a.settings.GetSMTPSetting(a.cfg.Email)
}

func (a settingsSMTPAdapter) PatchSMTPSetting(patch settingsmodule.SMTPSettingPatch) (settingsmodule.SMTPSetting, error) {
	return a.settings.PatchSMTPSetting(a.cfg.Email, patch)
}

func (a settingsSMTPAdapter) ApplyRuntime(setting settingsmodule.SMTPSetting) {
	a.cfg.Email = settingsmodule.SMTPSettingToConfig(setting)
	if a.email != nil {
		a.email.SetConfig(&a.cfg.Email)
	}
}

func (a settingsSMTPAdapter) SendTest(setting settingsmodule.SMTPSetting, toEmail, subject, body string) error {
	configForSend := settingsmodule.SMTPSettingToConfig(setting)
	configForSend.Enabled = true
	err := service.NewEmailService(&configForSend).SendCustomEmail(toEmail, subject, body)
	switch {
	case err == nil:
		return nil
	case errors.Is(err, service.ErrInvalidEmail):
		return settingstransport.ErrSMTPTestInvalidEmail
	case errors.Is(err, service.ErrEmailRecipientRejected):
		return settingstransport.ErrSMTPTestRecipientRejected
	case errors.Is(err, service.ErrEmailServiceDisabled):
		return settingstransport.ErrSMTPTestServiceDisabled
	case errors.Is(err, service.ErrEmailServiceNotConfigured):
		return settingstransport.ErrSMTPTestServiceNotConfigured
	default:
		return err
	}
}
