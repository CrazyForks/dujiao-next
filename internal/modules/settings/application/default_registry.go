package settingsapp

import (
	"github.com/dujiao-next/internal/constants"
	settingsmodule "github.com/dujiao-next/internal/modules/settings"
	"github.com/dujiao-next/internal/shared/jsonmap"
)

// defaultSettingRegistry 显式登记通用 Update API 支持的归一化策略。
// 每个定义只负责把所属领域的既有解析/默认值/序列化函数接入 Registry，
// 避免 SettingService 再通过全局 switch 了解所有配置类型。
var defaultSettingRegistry = MustNewRegistry(
	Definition{
		Key:       constants.SettingKeyDashboardConfig,
		Normalize: settingsmodule.NormalizeDashboardSettingJSON,
	},
	Definition{
		Key: constants.SettingKeyOrderConfig,
		Normalize: func(value jsonmap.JSON) jsonmap.JSON {
			return OrderConfigToMap(orderConfigFromJSON(value, DefaultOrderConfig()))
		},
	},
	Definition{
		Key:       constants.SettingKeySiteConfig,
		Normalize: func(value jsonmap.JSON) jsonmap.JSON { return normalizeSiteSetting(value) },
		Effects:   []Effect{EffectInvalidatePublicConfigCache},
	},
	Definition{
		Key:       constants.SettingKeyTelegramAuthConfig,
		Normalize: settingsmodule.NormalizeTelegramAuthSettingJSON,
	},
	Definition{
		Key: constants.SettingKeyNotificationCenterConfig,
		Normalize: func(value jsonmap.JSON) jsonmap.JSON {
			setting := settingsmodule.DecodeNotificationCenterSetting(value, settingsmodule.NotificationCenterDefaultSetting())
			return jsonmap.JSON(settingsmodule.NotificationCenterSettingToMap(setting))
		},
	},
	Definition{
		Key:       constants.SettingKeyAffiliateConfig,
		Normalize: settingsmodule.NormalizeAffiliateSettingJSON,
	},
	Definition{
		Key:       constants.SettingKeyTelegramBotConfig,
		Normalize: settingsmodule.NormalizeTelegramBotConfigJSON,
	},
	Definition{
		Key:       constants.SettingKeyNavConfig,
		Normalize: func(value jsonmap.JSON) jsonmap.JSON { return normalizeNavConfig(value) },
		Effects:   []Effect{EffectInvalidatePublicConfigCache},
	},
	Definition{
		Key:       constants.SettingKeyRegistrationConfig,
		Normalize: func(value jsonmap.JSON) jsonmap.JSON { return normalizeRegistrationSetting(value) },
		Effects:   []Effect{EffectInvalidatePublicConfigCache},
	},
	Definition{
		Key:       constants.SettingKeyOrderRiskControlConfig,
		Normalize: settingsmodule.NormalizeOrderRiskControlConfigJSON,
	},
	Definition{
		Key:       constants.SettingKeyUpstreamSyncConfig,
		Normalize: settingsmodule.NormalizeUpstreamSyncConfigJSON,
	},
	Definition{
		Key:       constants.SettingKeyCallbackRoutesConfig,
		Normalize: settingsmodule.NormalizeCallbackRoutesSettingJSON,
		Effects:   []Effect{EffectInvalidateCallbackRoutesCache},
	},
	Definition{
		Key:       constants.SettingKeyHomeAnnouncement,
		Normalize: settingsmodule.NormalizeHomeAnnouncementJSON,
		Effects:   []Effect{EffectInvalidatePublicConfigCache},
	},
	Definition{
		Key:     constants.SettingKeyWalletConfig,
		Effects: []Effect{EffectInvalidatePublicConfigCache},
	},
)
