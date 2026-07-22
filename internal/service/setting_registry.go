package service

import (
	"github.com/dujiao-next/internal/constants"
	"github.com/dujiao-next/internal/models"
	settingsmodule "github.com/dujiao-next/internal/modules/settings"
)

// defaultSettingRegistry 显式登记通用 Update API 支持的归一化策略。
// 每个定义只负责把所属领域的既有解析/默认值/序列化函数接入 Registry，
// 避免 SettingService 再通过全局 switch 了解所有配置类型。
var defaultSettingRegistry = settingsmodule.MustNewRegistry(
	settingsmodule.Definition{
		Key:       constants.SettingKeyDashboardConfig,
		Normalize: settingsmodule.NormalizeDashboardSettingJSON,
	},
	settingsmodule.Definition{
		Key: constants.SettingKeyOrderConfig,
		Normalize: func(value models.JSON) models.JSON {
			return OrderConfigToMap(orderConfigFromJSON(value, DefaultOrderConfig()))
		},
	},
	settingsmodule.Definition{
		Key:       constants.SettingKeySiteConfig,
		Normalize: func(value models.JSON) models.JSON { return normalizeSiteSetting(value) },
		Effects:   []settingsmodule.Effect{settingsmodule.EffectInvalidatePublicConfigCache},
	},
	settingsmodule.Definition{
		Key:       constants.SettingKeyTelegramAuthConfig,
		Normalize: settingsmodule.NormalizeTelegramAuthSettingJSON,
	},
	settingsmodule.Definition{
		Key: constants.SettingKeyNotificationCenterConfig,
		Normalize: func(value models.JSON) models.JSON {
			setting := settingsmodule.DecodeNotificationCenterSetting(value, settingsmodule.NotificationCenterDefaultSetting())
			return models.JSON(settingsmodule.NotificationCenterSettingToMap(setting))
		},
	},
	settingsmodule.Definition{
		Key:       constants.SettingKeyAffiliateConfig,
		Normalize: settingsmodule.NormalizeAffiliateSettingJSON,
	},
	settingsmodule.Definition{
		Key:       constants.SettingKeyTelegramBotConfig,
		Normalize: settingsmodule.NormalizeTelegramBotConfigJSON,
	},
	settingsmodule.Definition{
		Key:       constants.SettingKeyNavConfig,
		Normalize: func(value models.JSON) models.JSON { return normalizeNavConfig(value) },
		Effects:   []settingsmodule.Effect{settingsmodule.EffectInvalidatePublicConfigCache},
	},
	settingsmodule.Definition{
		Key:       constants.SettingKeyRegistrationConfig,
		Normalize: func(value models.JSON) models.JSON { return normalizeRegistrationSetting(value) },
		Effects:   []settingsmodule.Effect{settingsmodule.EffectInvalidatePublicConfigCache},
	},
	settingsmodule.Definition{
		Key:       constants.SettingKeyOrderRiskControlConfig,
		Normalize: settingsmodule.NormalizeOrderRiskControlConfigJSON,
	},
	settingsmodule.Definition{
		Key:       constants.SettingKeyUpstreamSyncConfig,
		Normalize: settingsmodule.NormalizeUpstreamSyncConfigJSON,
	},
	settingsmodule.Definition{
		Key:       constants.SettingKeyCallbackRoutesConfig,
		Normalize: settingsmodule.NormalizeCallbackRoutesSettingJSON,
		Effects:   []settingsmodule.Effect{settingsmodule.EffectInvalidateCallbackRoutesCache},
	},
	settingsmodule.Definition{
		Key:       constants.SettingKeyHomeAnnouncement,
		Normalize: settingsmodule.NormalizeHomeAnnouncementJSON,
		Effects:   []settingsmodule.Effect{settingsmodule.EffectInvalidatePublicConfigCache},
	},
	settingsmodule.Definition{
		Key:     constants.SettingKeyWalletConfig,
		Effects: []settingsmodule.Effect{settingsmodule.EffectInvalidatePublicConfigCache},
	},
)
