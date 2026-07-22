package orderrisk

import (
	"testing"

	settingssecurity "github.com/dujiao-next/internal/modules/settings/schema/security"
	"github.com/dujiao-next/internal/shared/jsonmap"
)

// --- isIPInBlacklist 测试 ---

func TestIsIPInBlacklist(t *testing.T) {
	svc := &Service{}

	blacklist := []string{"1.2.3.4", "10.0.0.0/8", "192.168.1.0/24"}

	tests := []struct {
		ip      string
		blocked bool
	}{
		{"1.2.3.4", true},        // exact match
		{"1.2.3.5", false},       // not in list
		{"10.0.0.1", true},       // CIDR /8 match
		{"10.255.255.255", true}, // CIDR /8 match
		{"11.0.0.1", false},      // outside CIDR
		{"192.168.1.100", true},  // CIDR /24 match
		{"192.168.2.1", false},   // outside CIDR /24
		{"", false},              // empty IP
		{"invalid", false},       // invalid IP
	}

	for _, tc := range tests {
		if got := svc.isIPInBlacklist(tc.ip, blacklist); got != tc.blocked {
			t.Errorf("isIPInBlacklist(%q) = %v, want %v", tc.ip, got, tc.blocked)
		}
	}
}

func TestIsIPInBlacklist_CacheReuse(t *testing.T) {
	svc := &Service{}
	blacklist := []string{"1.2.3.4", "10.0.0.0/8"}

	// First call builds cache
	svc.isIPInBlacklist("1.2.3.4", blacklist)
	if svc.cachedBlacklist == nil {
		t.Fatal("expected cache to be built")
	}
	hash1 := svc.cachedBlacklist.hash

	// Same list should reuse cache
	svc.isIPInBlacklist("10.0.0.1", blacklist)
	if svc.cachedBlacklist.hash != hash1 {
		t.Fatal("expected cache to be reused")
	}

	// Different list should rebuild cache
	svc.isIPInBlacklist("1.2.3.4", []string{"5.6.7.8"})
	if svc.cachedBlacklist.hash == hash1 {
		t.Fatal("expected cache to be rebuilt for different list")
	}
}

// --- RiskRateLimitedError 测试 ---

func TestRiskRateLimitedError_Is(t *testing.T) {
	err := &RiskRateLimitedError{RetryAfter: 60}
	if err.Error() != ErrRiskOrderRateLimited.Error() {
		t.Fatalf("expected error message match")
	}
	if !err.Is(ErrRiskOrderRateLimited) {
		t.Fatal("expected Is(ErrRiskOrderRateLimited) to be true")
	}
}

func TestGetRetryAfter(t *testing.T) {
	if ra := GetRetryAfter(ErrRiskIPBlacklisted); ra != 0 {
		t.Fatalf("expected 0 for non-rate-limit error, got %d", ra)
	}
	if ra := GetRetryAfter(&RiskRateLimitedError{RetryAfter: 42}); ra != 42 {
		t.Fatalf("expected 42, got %d", ra)
	}
}

// --- CheckOrderAllowed 集成测试（使用 mock） ---

type mockOrderRepoForRisk struct {
	pendingByUser  int64
	pendingByIP    int64
	pendingByEmail int64
}

func (m *mockOrderRepoForRisk) CountPendingByUserID(_ uint) (int64, error) {
	return m.pendingByUser, nil
}
func (m *mockOrderRepoForRisk) CountPendingByClientIP(_ string) (int64, error) {
	return m.pendingByIP, nil
}
func (m *mockOrderRepoForRisk) CountPendingByGuestEmail(_ string) (int64, error) {
	return m.pendingByEmail, nil
}

type settingReaderStub struct {
	config settingssecurity.OrderRiskControlConfig
}

func (s settingReaderStub) GetOrderRiskControlConfig() (settingssecurity.OrderRiskControlConfig, error) {
	return s.config, nil
}

func newTestRiskControlService(pendingByUser, pendingByIP, pendingByEmail int64, cfgJSON jsonmap.JSON) *Service {
	config := settingssecurity.DecodeOrderRiskControlConfig(cfgJSON, settingssecurity.DefaultOrderRiskControlConfig())
	return NewService(settingReaderStub{config: config}, &mockOrderRepoForRisk{
		pendingByUser:  pendingByUser,
		pendingByIP:    pendingByIP,
		pendingByEmail: pendingByEmail,
	})
}

func TestCheckOrderAllowed_DisabledByDefault(t *testing.T) {
	svc := newTestRiskControlService(100, 100, 100, nil)
	err := svc.CheckOrderAllowed(CheckInput{
		UserID:   1,
		ClientIP: "1.2.3.4",
	})
	if err != nil {
		t.Fatalf("expected nil when disabled, got %v", err)
	}
}

func TestCheckOrderAllowed_IPBlacklist(t *testing.T) {
	svc := newTestRiskControlService(0, 0, 0, jsonmap.JSON{
		"enabled":      true,
		"ip_blacklist": []interface{}{"1.2.3.4", "10.0.0.0/8"},
	})

	if err := svc.CheckOrderAllowed(CheckInput{ClientIP: "1.2.3.4"}); err != ErrRiskIPBlacklisted {
		t.Fatalf("expected ErrRiskIPBlacklisted, got %v", err)
	}
	if err := svc.CheckOrderAllowed(CheckInput{ClientIP: "10.0.0.1"}); err != ErrRiskIPBlacklisted {
		t.Fatalf("expected ErrRiskIPBlacklisted for CIDR, got %v", err)
	}
	if err := svc.CheckOrderAllowed(CheckInput{ClientIP: "2.3.4.5"}); err != nil {
		t.Fatalf("expected nil for non-blocked IP, got %v", err)
	}
}

func TestCheckOrderAllowed_EmailBlacklist(t *testing.T) {
	svc := newTestRiskControlService(0, 0, 0, jsonmap.JSON{
		"enabled":         true,
		"email_blacklist": []interface{}{"spam@example.com"},
	})

	if err := svc.CheckOrderAllowed(CheckInput{
		IsGuest:    true,
		GuestEmail: "SPAM@example.com",
		ClientIP:   "2.3.4.5",
	}); err != ErrRiskEmailBlacklisted {
		t.Fatalf("expected ErrRiskEmailBlacklisted, got %v", err)
	}

	// Non-guest should not be blocked by email blacklist
	if err := svc.CheckOrderAllowed(CheckInput{
		UserID:   1,
		ClientIP: "2.3.4.5",
	}); err != nil {
		t.Fatalf("expected nil for non-guest, got %v", err)
	}
}

func TestCheckOrderAllowed_PendingOrderLimits(t *testing.T) {
	cfg := jsonmap.JSON{
		"enabled":                            true,
		"max_pending_orders_per_user":        float64(2),
		"max_pending_orders_per_ip":          float64(3),
		"max_pending_orders_per_guest_email": float64(1),
	}

	// User at limit
	svc := newTestRiskControlService(2, 0, 0, cfg)
	if err := svc.CheckOrderAllowed(CheckInput{UserID: 1, ClientIP: "5.6.7.8"}); err != ErrRiskTooManyPendingOrders {
		t.Fatalf("expected ErrRiskTooManyPendingOrders for user, got %v", err)
	}

	// User under limit
	svc = newTestRiskControlService(1, 0, 0, cfg)
	if err := svc.CheckOrderAllowed(CheckInput{UserID: 1, ClientIP: "5.6.7.8"}); err != nil {
		t.Fatalf("expected nil for user under limit, got %v", err)
	}

	// IP at limit
	svc = newTestRiskControlService(0, 3, 0, cfg)
	if err := svc.CheckOrderAllowed(CheckInput{ClientIP: "5.6.7.8"}); err != ErrRiskTooManyPendingOrders {
		t.Fatalf("expected ErrRiskTooManyPendingOrders for IP, got %v", err)
	}

	// Guest email at limit
	svc = newTestRiskControlService(0, 0, 1, cfg)
	if err := svc.CheckOrderAllowed(CheckInput{
		IsGuest:    true,
		GuestEmail: "test@example.com",
		ClientIP:   "5.6.7.8",
	}); err != ErrRiskTooManyPendingOrders {
		t.Fatalf("expected ErrRiskTooManyPendingOrders for guest email, got %v", err)
	}
}

func TestCheckOrderAllowed_SkipIPCheck(t *testing.T) {
	cfg := jsonmap.JSON{
		"enabled":                     true,
		"max_pending_orders_per_user": float64(5),
		"max_pending_orders_per_ip":   float64(1),
		"ip_blacklist":                []interface{}{"1.2.3.4"},
	}

	// IP 在黑名单中，但 SkipIPCheck=true 应放行
	svc := newTestRiskControlService(0, 0, 0, cfg)
	if err := svc.CheckOrderAllowed(CheckInput{
		UserID:      1,
		ClientIP:    "1.2.3.4",
		SkipIPCheck: true,
	}); err != nil {
		t.Fatalf("expected nil with SkipIPCheck, got %v", err)
	}

	// IP 并发超限，但 SkipIPCheck=true 应放行
	svc = newTestRiskControlService(0, 999, 0, cfg)
	if err := svc.CheckOrderAllowed(CheckInput{
		UserID:      1,
		ClientIP:    "5.6.7.8",
		SkipIPCheck: true,
	}); err != nil {
		t.Fatalf("expected nil with SkipIPCheck for IP pending, got %v", err)
	}

	// 用户维度超限，SkipIPCheck=true 不影响，仍应拦截
	svc = newTestRiskControlService(5, 0, 0, cfg)
	if err := svc.CheckOrderAllowed(CheckInput{
		UserID:      1,
		ClientIP:    "5.6.7.8",
		SkipIPCheck: true,
	}); err != ErrRiskTooManyPendingOrders {
		t.Fatalf("expected ErrRiskTooManyPendingOrders for user limit with SkipIPCheck, got %v", err)
	}
}

func TestCheckOrderAllowed_ZeroLimitMeansNoLimit(t *testing.T) {
	svc := newTestRiskControlService(999, 999, 999, jsonmap.JSON{
		"enabled":                            true,
		"max_pending_orders_per_user":        float64(0),
		"max_pending_orders_per_ip":          float64(0),
		"max_pending_orders_per_guest_email": float64(0),
	})
	if err := svc.CheckOrderAllowed(CheckInput{
		UserID:     1,
		ClientIP:   "1.2.3.4",
		IsGuest:    true,
		GuestEmail: "test@example.com",
	}); err != nil {
		t.Fatalf("expected nil when limits are 0 (disabled), got %v", err)
	}
}
