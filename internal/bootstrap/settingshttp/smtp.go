package settingsbootstrap

import (
	"errors"

	settingsapp "github.com/dujiao-next/internal/modules/settings/application"
	settingsmessaging "github.com/dujiao-next/internal/modules/settings/schema/messaging"

	"github.com/dujiao-next/internal/config"
	settingstransport "github.com/dujiao-next/internal/modules/settings/transport/http"
	"github.com/dujiao-next/internal/service"
)

type settingsSMTPAdapter struct {
	settings *settingsapp.Service
	cfg      *config.Config
	email    *service.EmailService
}

func (a settingsSMTPAdapter) GetSMTPSetting() (settingsmessaging.SMTPSetting, error) {
	return a.settings.GetSMTPSetting(a.cfg.Email)
}

func (a settingsSMTPAdapter) PatchSMTPSetting(patch settingsmessaging.SMTPSettingPatch) (settingsmessaging.SMTPSetting, error) {
	return a.settings.PatchSMTPSetting(a.cfg.Email, patch)
}

func (a settingsSMTPAdapter) ApplyRuntime(setting settingsmessaging.SMTPSetting) {
	a.cfg.Email = settingsmessaging.SMTPSettingToConfig(setting)
	if a.email != nil {
		a.email.SetConfig(&a.cfg.Email)
	}
}

func (a settingsSMTPAdapter) SendTest(setting settingsmessaging.SMTPSetting, toEmail, subject, body string) error {
	configForSend := settingsmessaging.SMTPSettingToConfig(setting)
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
