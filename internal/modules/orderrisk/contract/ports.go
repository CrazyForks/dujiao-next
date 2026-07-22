package contract

import settingssecurity "github.com/dujiao-next/internal/modules/settings/schema/security"

type SettingReader interface {
	GetOrderRiskControlConfig() (settingssecurity.OrderRiskControlConfig, error)
}

type PendingOrderCounter interface {
	CountPendingByUserID(userID uint) (int64, error)
	CountPendingByClientIP(clientIP string) (int64, error)
	CountPendingByGuestEmail(email string) (int64, error)
}

// RateLimiter 执行具有外部状态的下单频率限制。
type RateLimiter interface {
	Check(input CheckInput, config settingssecurity.OrderRateLimitConfig) error
}

// Controller 是订单上下文调用风控所需的用例端口。
type Controller interface {
	CheckOrderAllowed(input CheckInput) error
}
