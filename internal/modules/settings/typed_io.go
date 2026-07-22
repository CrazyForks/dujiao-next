package settings

import (
	"time"

	"github.com/dujiao-next/internal/config"
	"github.com/dujiao-next/internal/constants"
)

// GetDashboardSetting 获取仪表盘设置（优先 settings，空时回退默认）。
func (s *Service) GetDashboardSetting() (DashboardSetting, error) {
	fallback := DefaultDashboardSetting()
	if s == nil {
		return fallback, nil
	}
	value, err := s.GetByKey(constants.SettingKeyDashboardConfig)
	if err != nil {
		return fallback, err
	}
	if value == nil {
		return fallback, nil
	}
	return DecodeDashboardSetting(value, fallback), nil
}

// GetDashboardLowStockThreshold 获取低库存阈值（读取失败回退默认值）。
func (s *Service) GetDashboardLowStockThreshold() int {
	defaultThreshold := int(DefaultDashboardSetting().Alert.LowStockThreshold)
	if s == nil {
		return defaultThreshold
	}
	setting, err := s.GetDashboardSetting()
	if err != nil {
		return defaultThreshold
	}
	return int(setting.Alert.LowStockThreshold)
}

// GetAffiliateSetting 获取推广返利设置（优先 settings，空时回退默认）。
func (s *Service) GetAffiliateSetting() (AffiliateSetting, error) {
	fallback := DefaultAffiliateSetting()
	if s == nil {
		return fallback, nil
	}
	value, err := s.GetByKey(constants.SettingKeyAffiliateConfig)
	if err != nil {
		return fallback, err
	}
	if value == nil {
		return fallback, nil
	}
	return DecodeAffiliateSetting(value, fallback), nil
}

// UpdateAffiliateSetting 更新推广返利设置。
func (s *Service) UpdateAffiliateSetting(setting AffiliateSetting) (AffiliateSetting, error) {
	normalized := NormalizeAffiliateSetting(setting)
	if err := ValidateAffiliateSetting(normalized); err != nil {
		return DefaultAffiliateSetting(), err
	}
	if _, err := s.Update(constants.SettingKeyAffiliateConfig, map[string]interface{}(EncodeAffiliateSetting(normalized))); err != nil {
		return DefaultAffiliateSetting(), err
	}
	return normalized, nil
}

// GetUpstreamSyncConfig 获取上游同步配置。
// fallbackInterval 来自 config.yml，仅在数据库没有覆盖值时使用。
func (s *Service) GetUpstreamSyncConfig(fallbackInterval string) (UpstreamSyncConfig, error) {
	fallback := UpstreamSyncFallback(fallbackInterval)
	if s == nil {
		return fallback, nil
	}
	value, err := s.GetByKey(constants.SettingKeyUpstreamSyncConfig)
	if err != nil {
		return fallback, err
	}
	return DecodeUpstreamSyncConfig(value, fallback), nil
}

// GetUpstreamSyncInterval 返回归一化后的同步间隔。
func (s *Service) GetUpstreamSyncInterval(fallbackInterval string) (time.Duration, error) {
	config, err := s.GetUpstreamSyncConfig(fallbackInterval)
	if err != nil {
		return time.Duration(config.IntervalMinutes) * time.Minute, err
	}
	return time.Duration(config.IntervalMinutes) * time.Minute, nil
}

// GetNotificationCenterSetting 获取通知中心配置（优先 settings，空时回退默认）。
func (s *Service) GetNotificationCenterSetting() (NotificationCenterSetting, error) {
	fallback := NotificationCenterDefaultSetting()
	if s == nil {
		return fallback, nil
	}
	value, err := s.GetByKey(constants.SettingKeyNotificationCenterConfig)
	if err != nil {
		return fallback, err
	}
	if value == nil {
		return fallback, nil
	}
	return NormalizeNotificationCenterSetting(DecodeNotificationCenterSetting(value, fallback)), nil
}

// PatchNotificationCenterSetting 基于补丁更新通知中心配置。
func (s *Service) PatchNotificationCenterSetting(patch NotificationCenterSettingPatch) (NotificationCenterSetting, error) {
	current, err := s.GetNotificationCenterSetting()
	if err != nil {
		return NotificationCenterSetting{}, err
	}
	next, err := ApplyNotificationCenterSettingPatch(current, patch)
	if err != nil {
		return NotificationCenterSetting{}, err
	}
	if _, err := s.Update(constants.SettingKeyNotificationCenterConfig, NotificationCenterSettingToMap(next)); err != nil {
		return NotificationCenterSetting{}, err
	}
	return next, nil
}

// GetSMTPSetting 获取 SMTP 设置（优先 settings，空时回退默认配置）。
func (s *Service) GetSMTPSetting(defaultCfg config.EmailConfig) (SMTPSetting, error) {
	fallback := DefaultSMTPSetting(defaultCfg)
	if s == nil {
		return fallback, nil
	}
	value, err := s.GetByKey(constants.SettingKeySMTPConfig)
	if err != nil {
		return fallback, err
	}
	if value == nil {
		return fallback, nil
	}
	return NormalizeSMTPSetting(DecodeSMTPSetting(value, fallback)), nil
}

// PatchSMTPSetting 基于补丁更新 SMTP 设置。
func (s *Service) PatchSMTPSetting(defaultCfg config.EmailConfig, patch SMTPSettingPatch) (SMTPSetting, error) {
	current, err := s.GetSMTPSetting(defaultCfg)
	if err != nil {
		return SMTPSetting{}, err
	}
	next, err := ApplySMTPSettingPatch(current, patch)
	if err != nil {
		return SMTPSetting{}, err
	}
	if _, err := s.Update(constants.SettingKeySMTPConfig, map[string]interface{}(EncodeSMTPSetting(next))); err != nil {
		return SMTPSetting{}, err
	}
	return next, nil
}

// GetCaptchaSetting 获取验证码设置（优先 settings，空时回退 config.yml）。
func (s *Service) GetCaptchaSetting(defaultCfg config.CaptchaConfig) (CaptchaSetting, error) {
	fallback := DefaultCaptchaSetting(defaultCfg)
	if s == nil {
		return fallback, nil
	}
	value, err := s.GetByKey(constants.SettingKeyCaptchaConfig)
	if err != nil {
		return fallback, err
	}
	if value == nil {
		return fallback, nil
	}
	return NormalizeCaptchaSetting(DecodeCaptchaSetting(value, fallback)), nil
}

// PatchCaptchaSetting 基于补丁更新验证码设置。
func (s *Service) PatchCaptchaSetting(defaultCfg config.CaptchaConfig, patch CaptchaSettingPatch) (CaptchaSetting, error) {
	current, err := s.GetCaptchaSetting(defaultCfg)
	if err != nil {
		return CaptchaSetting{}, err
	}
	next, err := ApplyCaptchaSettingPatch(current, patch)
	if err != nil {
		return CaptchaSetting{}, err
	}
	if _, err := s.Update(constants.SettingKeyCaptchaConfig, map[string]interface{}(EncodeCaptchaSetting(next))); err != nil {
		return CaptchaSetting{}, err
	}
	return next, nil
}

// GetTelegramAuthSetting 获取 Telegram 登录配置。
func (s *Service) GetTelegramAuthSetting(defaultCfg config.TelegramAuthConfig) (TelegramAuthSetting, error) {
	fallback := DefaultTelegramAuthSetting(defaultCfg)
	if s == nil {
		return fallback, nil
	}
	value, err := s.GetByKey(constants.SettingKeyTelegramAuthConfig)
	if err != nil {
		return fallback, err
	}
	if value == nil {
		return fallback, nil
	}
	return NormalizeTelegramAuthSetting(DecodeTelegramAuthSetting(value, fallback)), nil
}

// PatchTelegramAuthSetting 基于补丁更新 Telegram 登录配置。
func (s *Service) PatchTelegramAuthSetting(defaultCfg config.TelegramAuthConfig, patch TelegramAuthSettingPatch) (TelegramAuthSetting, error) {
	current, err := s.GetTelegramAuthSetting(defaultCfg)
	if err != nil {
		return TelegramAuthSetting{}, err
	}
	next := NormalizeTelegramAuthSetting(ApplyTelegramAuthSettingPatch(current, patch))
	if err := ValidateTelegramAuthSetting(next); err != nil {
		return TelegramAuthSetting{}, err
	}
	if _, err := s.Update(constants.SettingKeyTelegramAuthConfig, map[string]interface{}(EncodeTelegramAuthSetting(next))); err != nil {
		return TelegramAuthSetting{}, err
	}
	return next, nil
}

// GetOrderEmailTemplateSetting 获取订单邮件模板配置（优先 settings，空时回退默认）。
func (s *Service) GetOrderEmailTemplateSetting() (OrderEmailTemplateSetting, error) {
	fallback := DefaultOrderEmailTemplateSetting()
	if s == nil {
		return fallback, nil
	}
	value, err := s.GetByKey(constants.SettingKeyOrderEmailTemplateConfig)
	if err != nil {
		return fallback, err
	}
	if value == nil {
		return fallback, nil
	}
	return NormalizeOrderEmailTemplateSetting(DecodeOrderEmailTemplateSetting(value, fallback)), nil
}

// PatchOrderEmailTemplateSetting 基于补丁更新订单邮件模板配置。
func (s *Service) PatchOrderEmailTemplateSetting(patch OrderEmailTemplateSettingPatch) (OrderEmailTemplateSetting, error) {
	current, err := s.GetOrderEmailTemplateSetting()
	if err != nil {
		return OrderEmailTemplateSetting{}, err
	}
	next, err := ApplyOrderEmailTemplateSettingPatch(current, patch)
	if err != nil {
		return OrderEmailTemplateSetting{}, err
	}
	if _, err := s.Update(constants.SettingKeyOrderEmailTemplateConfig, map[string]interface{}(EncodeOrderEmailTemplateSetting(next))); err != nil {
		return OrderEmailTemplateSetting{}, err
	}
	return next, nil
}

// ResetOrderEmailTemplateSetting 重置订单邮件模板为默认。
func (s *Service) ResetOrderEmailTemplateSetting() (OrderEmailTemplateSetting, error) {
	defaultSetting := DefaultOrderEmailTemplateSetting()
	if s == nil {
		return defaultSetting, nil
	}
	if _, err := s.Update(constants.SettingKeyOrderEmailTemplateConfig, map[string]interface{}(EncodeOrderEmailTemplateSetting(defaultSetting))); err != nil {
		return OrderEmailTemplateSetting{}, err
	}
	return defaultSetting, nil
}

// GetTelegramBotConfig 获取 Telegram Bot 配置。
func (s *Service) GetTelegramBotConfig() (TelegramBotConfigSetting, error) {
	fallback := DefaultTelegramBotConfig()
	if s == nil {
		return fallback, nil
	}
	value, err := s.GetByKey(constants.SettingKeyTelegramBotConfig)
	if err != nil {
		return fallback, err
	}
	if value == nil {
		return fallback, nil
	}
	parsed := DecodeTelegramBotConfig(value, fallback)
	parsed.Menu.Items = ensureBuiltinMenuItems(parsed.Menu.Items)
	return parsed, nil
}

// UpdateTelegramBotConfig 整对象覆盖更新 Telegram Bot 配置，自动递增 config_version。
func (s *Service) UpdateTelegramBotConfig(cfg TelegramBotConfigSetting) (TelegramBotConfigSetting, error) {
	current, err := s.GetTelegramBotConfig()
	if err != nil {
		return TelegramBotConfigSetting{}, err
	}

	cfg.ConfigVersion = current.ConfigVersion + 1
	cfg.Basic.Description = normalizeLocalizedText(cfg.Basic.Description)
	cfg.Welcome.Message = normalizeLocalizedText(cfg.Welcome.Message)
	cfg.Help.Title = normalizeLocalizedText(cfg.Help.Title)
	cfg.Help.Intro = normalizeLocalizedText(cfg.Help.Intro)
	cfg.Help.CenterHint = normalizeLocalizedText(cfg.Help.CenterHint)
	cfg.Help.SupportHint = normalizeLocalizedText(cfg.Help.SupportHint)
	cfg.Help.Items = normalizeHelpItems(cfg.Help.Items)
	cfg.Menu.Items = NormalizeTelegramBotMenuItems(cfg.Menu.Items)

	if _, err := s.Update(constants.SettingKeyTelegramBotConfig, EncodeTelegramBotConfig(cfg)); err != nil {
		return TelegramBotConfigSetting{}, err
	}

	runtimeStatus, _ := s.GetTelegramBotRuntimeStatus()
	runtimeStatus.ConfigVersion = cfg.ConfigVersion
	_ = s.UpdateTelegramBotRuntimeStatus(runtimeStatus)

	return cfg, nil
}

// GetTelegramBotRuntimeStatus 获取 Telegram Bot 运行时状态。
func (s *Service) GetTelegramBotRuntimeStatus() (TelegramBotRuntimeStatusSetting, error) {
	fallback := DefaultTelegramBotRuntimeStatus()
	if s == nil {
		return fallback, nil
	}
	value, err := s.GetByKey(constants.SettingKeyTelegramBotRuntimeStatus)
	if err != nil {
		return fallback, err
	}
	if value == nil {
		return fallback, nil
	}
	return DecodeTelegramBotRuntimeStatus(value, fallback), nil
}

// UpdateTelegramBotRuntimeStatus 更新 Telegram Bot 运行时状态。
func (s *Service) UpdateTelegramBotRuntimeStatus(status TelegramBotRuntimeStatusSetting) error {
	if s == nil {
		return nil
	}
	_, err := s.Update(constants.SettingKeyTelegramBotRuntimeStatus, EncodeTelegramBotRuntimeStatus(status))
	return err
}
