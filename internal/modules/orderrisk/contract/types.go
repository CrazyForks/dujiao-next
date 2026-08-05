package contract

import (
	"net"
	"strings"
)

// OrderItem 是风控所需的最小订单项快照。
type OrderItem struct {
	ProductID uint
	Quantity  int
}

// CheckInput 是订单风控检查所需的调用上下文。
type CheckInput struct {
	UserID           uint
	ClientIP         string
	RiskIP           string
	IsGuest          bool
	SkipIPCheck      bool
	ConsumeRateLimit bool
	Items            []OrderItem
}

// CheckResult 返回创建订单后续步骤需要复用的规范化风控上下文。
type CheckResult struct {
	RiskIP               string
	PaymentExpireMinutes int
}

// NormalizeRiskIP 生成游客风控键：IPv4 使用完整地址，IPv6 按 /64 前缀聚合。
func NormalizeRiskIP(raw string) string {
	ip := net.ParseIP(strings.TrimSpace(raw))
	if ip == nil {
		return ""
	}
	if ipv4 := ip.To4(); ipv4 != nil {
		return ipv4.String()
	}
	network := ip.Mask(net.CIDRMask(64, 128))
	if network == nil {
		return ""
	}
	return network.String() + "/64"
}
