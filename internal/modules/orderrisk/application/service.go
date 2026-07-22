package application

import (
	"crypto/sha256"
	"fmt"
	"net"
	"strings"
	"sync"

	"github.com/dujiao-next/internal/logger"
	orderriskcontract "github.com/dujiao-next/internal/modules/orderrisk/contract"
	settingssecurity "github.com/dujiao-next/internal/modules/settings/schema/security"
)

type parsedIPBlacklist struct {
	exactIPs map[string]struct{}
	cidrs    []*net.IPNet
	hash     string
}

// Options 声明订单风控应用服务的全部端口。
type Options struct {
	Settings    orderriskcontract.SettingReader
	Orders      orderriskcontract.PendingOrderCounter
	RateLimiter orderriskcontract.RateLimiter
}

// Service 编排订单黑名单、待支付数量和下单频率检查。
type Service struct {
	settings    orderriskcontract.SettingReader
	orders      orderriskcontract.PendingOrderCounter
	rateLimiter orderriskcontract.RateLimiter

	mu              sync.RWMutex
	cachedBlacklist *parsedIPBlacklist
}

var _ orderriskcontract.Controller = (*Service)(nil)

func NewService(options Options) *Service {
	return &Service{
		settings:    options.Settings,
		orders:      options.Orders,
		rateLimiter: options.RateLimiter,
	}
}

func (s *Service) CheckOrderAllowed(input orderriskcontract.CheckInput) error {
	if s == nil || s.settings == nil {
		return nil
	}
	cfg, err := s.settings.GetOrderRiskControlConfig()
	if err != nil {
		logger.Warnw("risk_control_get_config_error", "error", err)
		return nil
	}
	if !cfg.Enabled {
		return nil
	}

	if !input.SkipIPCheck && input.ClientIP != "" && len(cfg.IPBlacklist) > 0 && s.isIPInBlacklist(input.ClientIP, cfg.IPBlacklist) {
		return orderriskcontract.ErrIPBlacklisted
	}
	if input.IsGuest && input.GuestEmail != "" && len(cfg.EmailBlacklist) > 0 {
		normalizedEmail := strings.ToLower(strings.TrimSpace(input.GuestEmail))
		for _, blocked := range cfg.EmailBlacklist {
			if normalizedEmail == blocked {
				return orderriskcontract.ErrEmailBlacklisted
			}
		}
	}
	if err := s.checkPendingOrderLimits(input, cfg); err != nil {
		return err
	}
	if cfg.OrderRateLimit.Enabled && s.rateLimiter != nil {
		return s.rateLimiter.Check(input, cfg.OrderRateLimit)
	}
	return nil
}

func (s *Service) checkPendingOrderLimits(input orderriskcontract.CheckInput, cfg settingssecurity.OrderRiskControlConfig) error {
	if s.orders == nil {
		return nil
	}
	if input.UserID > 0 && cfg.MaxPendingOrdersPerUser > 0 {
		count, err := s.orders.CountPendingByUserID(input.UserID)
		if err != nil {
			logger.Warnw("risk_control_count_pending_by_user_error", "user_id", input.UserID, "error", err)
		} else if count >= int64(cfg.MaxPendingOrdersPerUser) {
			return orderriskcontract.ErrTooManyPendingOrders
		}
	}
	if !input.SkipIPCheck && input.ClientIP != "" && cfg.MaxPendingOrdersPerIP > 0 {
		count, err := s.orders.CountPendingByClientIP(input.ClientIP)
		if err != nil {
			logger.Warnw("risk_control_count_pending_by_ip_error", "ip", input.ClientIP, "error", err)
		} else if count >= int64(cfg.MaxPendingOrdersPerIP) {
			return orderriskcontract.ErrTooManyPendingOrders
		}
	}
	if input.IsGuest && input.GuestEmail != "" && cfg.MaxPendingOrdersPerGuestEmail > 0 {
		count, err := s.orders.CountPendingByGuestEmail(input.GuestEmail)
		if err != nil {
			logger.Warnw("risk_control_count_pending_by_email_error", "email", input.GuestEmail, "error", err)
		} else if count >= int64(cfg.MaxPendingOrdersPerGuestEmail) {
			return orderriskcontract.ErrTooManyPendingOrders
		}
	}
	return nil
}

func (s *Service) getOrBuildBlacklist(blacklist []string) *parsedIPBlacklist {
	hash := hashBlacklist(blacklist)
	s.mu.RLock()
	if s.cachedBlacklist != nil && s.cachedBlacklist.hash == hash {
		cached := s.cachedBlacklist
		s.mu.RUnlock()
		return cached
	}
	s.mu.RUnlock()

	parsed := &parsedIPBlacklist{exactIPs: make(map[string]struct{}, len(blacklist)), hash: hash}
	for _, entry := range blacklist {
		if strings.Contains(entry, "/") {
			_, cidr, err := net.ParseCIDR(entry)
			if err == nil {
				parsed.cidrs = append(parsed.cidrs, cidr)
			}
		} else {
			parsed.exactIPs[entry] = struct{}{}
		}
	}

	s.mu.Lock()
	if s.cachedBlacklist == nil || s.cachedBlacklist.hash != hash {
		s.cachedBlacklist = parsed
	}
	s.mu.Unlock()
	return parsed
}

func hashBlacklist(list []string) string {
	hash := sha256.New()
	for _, value := range list {
		hash.Write([]byte(value))
		hash.Write([]byte{0})
	}
	return fmt.Sprintf("%x", hash.Sum(nil))
}

func (s *Service) isIPInBlacklist(clientIP string, blacklist []string) bool {
	ip := net.ParseIP(clientIP)
	if ip == nil {
		return false
	}
	parsed := s.getOrBuildBlacklist(blacklist)
	if _, exists := parsed.exactIPs[clientIP]; exists {
		return true
	}
	for _, cidr := range parsed.cidrs {
		if cidr.Contains(ip) {
			return true
		}
	}
	return false
}
