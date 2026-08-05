package application

import (
	"errors"
	"reflect"
	"testing"

	orderriskcontract "github.com/dujiao-next/internal/modules/orderrisk/contract"
	settingssecurity "github.com/dujiao-next/internal/modules/settings/schema/security"
)

type settingReaderStub struct {
	config settingssecurity.OrderRiskControlConfig
	err    error
}

func (s settingReaderStub) GetOrderRiskControlConfig() (settingssecurity.OrderRiskControlConfig, error) {
	return s.config, s.err
}

type rateLimiterStub struct {
	calls  int
	input  orderriskcontract.CheckInput
	config settingssecurity.OrderRateLimitConfig
	err    error
}

func (s *rateLimiterStub) Check(input orderriskcontract.CheckInput, config settingssecurity.OrderRateLimitConfig) error {
	s.calls++
	s.input = input
	s.config = config
	return s.err
}

type pendingGateStub struct {
	lockedKeys       []string
	pendingByUser    int64
	pendingGuestIP   int64
	pendingMemberIP  int64
	pendingByProduct map[uint]int64
	err              error
}

func (s *pendingGateStub) LockRiskKeys(keys []string) error {
	s.lockedKeys = append([]string(nil), keys...)
	return s.err
}
func (s *pendingGateStub) CountPendingByUserID(uint) (int64, error) {
	return s.pendingByUser, s.err
}
func (s *pendingGateStub) CountPendingGuestByRiskIP(string) (int64, error) {
	return s.pendingGuestIP, s.err
}
func (s *pendingGateStub) CountPendingMemberByRiskIP(string) (int64, error) {
	return s.pendingMemberIP, s.err
}
func (s *pendingGateStub) SumPendingGuestQuantityByRiskIP(string, []uint) (map[uint]int64, error) {
	return s.pendingByProduct, s.err
}

func testConfig() settingssecurity.OrderRiskControlConfig {
	cfg := settingssecurity.DefaultOrderRiskControlConfig()
	cfg.Enabled = true
	return cfg
}

func TestCheckOrderAllowed_DisabledByDefault(t *testing.T) {
	svc := NewService(Options{Settings: settingReaderStub{config: settingssecurity.DefaultOrderRiskControlConfig()}})
	result, err := svc.CheckOrderAllowed(orderriskcontract.CheckInput{
		IsGuest:  true,
		ClientIP: "2001:db8:abcd:12::1234",
		Items:    []orderriskcontract.OrderItem{{ProductID: 1, Quantity: 999}},
	})
	if err != nil {
		t.Fatalf("expected nil when globally disabled, got %v", err)
	}
	if result.RiskIP != "2001:db8:abcd:12::/64" {
		t.Fatalf("expected normalized risk IP even while disabled, got %q", result.RiskIP)
	}
}

func TestCheckOrderAllowed_IPBlacklistCanonicalizesAddress(t *testing.T) {
	cfg := testConfig()
	cfg.Common.IPBlacklist = []string{"1.2.3.4", "10.0.0.0/8"}
	svc := NewService(Options{Settings: settingReaderStub{config: cfg}})

	for _, ip := range []string{"1.2.3.4", "::ffff:1.2.3.4", "10.2.3.4"} {
		if _, err := svc.CheckOrderAllowed(orderriskcontract.CheckInput{IsGuest: true, ClientIP: ip}); !errors.Is(err, orderriskcontract.ErrIPBlacklisted) {
			t.Fatalf("expected %q to be blacklisted, got %v", ip, err)
		}
	}
}

func TestCheckOrderAllowed_GuestQuantityAggregatesProductAcrossSKUs(t *testing.T) {
	cfg := testConfig()
	cfg.Guest.MaxQuantityPerProductPerOrder = 2
	svc := NewService(Options{Settings: settingReaderStub{config: cfg}})

	_, err := svc.CheckOrderAllowed(orderriskcontract.CheckInput{
		IsGuest:  true,
		ClientIP: "1.2.3.4",
		Items: []orderriskcontract.OrderItem{
			{ProductID: 7, Quantity: 1},
			{ProductID: 7, Quantity: 2},
		},
	})
	if !errors.Is(err, orderriskcontract.ErrProductQuantityLimit) {
		t.Fatalf("expected aggregate guest quantity to be rejected, got %v", err)
	}
}

func TestCheckOrderAllowed_MemberUsesIndependentQuantityPolicy(t *testing.T) {
	cfg := testConfig()
	cfg.Guest.MaxQuantityPerProductPerOrder = 1
	cfg.Member.MaxQuantityPerProductPerOrder = 5
	svc := NewService(Options{Settings: settingReaderStub{config: cfg}})

	if _, err := svc.CheckOrderAllowed(orderriskcontract.CheckInput{
		UserID: 3,
		Items:  []orderriskcontract.OrderItem{{ProductID: 1, Quantity: 5}},
	}); err != nil {
		t.Fatalf("member quantity must not use guest limit: %v", err)
	}
	if _, err := svc.CheckOrderAllowed(orderriskcontract.CheckInput{
		UserID: 3,
		Items:  []orderriskcontract.OrderItem{{ProductID: 1, Quantity: 6}},
	}); !errors.Is(err, orderriskcontract.ErrProductQuantityLimit) {
		t.Fatalf("expected member quantity limit, got %v", err)
	}
}

func TestCheckOrderAllowed_GuestRateLimitUsesRiskIPOnlyWhenConsuming(t *testing.T) {
	cfg := testConfig()
	limited := &orderriskcontract.RateLimitedError{RetryAfter: 42}
	limiter := &rateLimiterStub{err: limited}
	svc := NewService(Options{Settings: settingReaderStub{config: cfg}, RateLimiter: limiter})
	input := orderriskcontract.CheckInput{IsGuest: true, ClientIP: "2001:db8:1:2::abcd"}

	if _, err := svc.CheckOrderAllowed(input); err != nil {
		t.Fatalf("preview-style check must not consume rate limit: %v", err)
	}
	if limiter.calls != 0 {
		t.Fatalf("expected no rate limiter call, got %d", limiter.calls)
	}
	input.ConsumeRateLimit = true
	if _, err := svc.CheckOrderAllowed(input); err != limited {
		t.Fatalf("expected limiter error, got %v", err)
	}
	if limiter.calls != 1 || limiter.input.RiskIP != "2001:db8:1:2::/64" || limiter.input.UserID != 0 {
		t.Fatalf("unexpected guest limiter input: calls=%d input=%+v", limiter.calls, limiter.input)
	}
}

func TestCheckPendingOrderAllowed_GuestLocksIPAndCountsOrders(t *testing.T) {
	cfg := testConfig()
	cfg.Guest.MaxPendingOrdersPerIP = 2
	gate := &pendingGateStub{pendingGuestIP: 2}
	svc := NewService(Options{Settings: settingReaderStub{config: cfg}})

	err := svc.CheckPendingOrderAllowed(orderriskcontract.CheckInput{
		IsGuest: true,
		RiskIP:  "1.2.3.4",
		Items:   []orderriskcontract.OrderItem{{ProductID: 1, Quantity: 1}},
	}, gate)
	if !errors.Is(err, orderriskcontract.ErrTooManyPendingOrders) {
		t.Fatalf("expected guest pending order limit, got %v", err)
	}
	if !reflect.DeepEqual(gate.lockedKeys, []string{"guest:ip:1.2.3.4"}) {
		t.Fatalf("unexpected lock keys: %#v", gate.lockedKeys)
	}
}

func TestCheckPendingOrderAllowed_GuestCountsPendingProductQuantity(t *testing.T) {
	cfg := testConfig()
	cfg.Guest.MaxPendingOrdersPerIP = 10
	cfg.Guest.MaxPendingQuantityPerIPProduct = 2
	gate := &pendingGateStub{pendingByProduct: map[uint]int64{7: 1}}
	svc := NewService(Options{Settings: settingReaderStub{config: cfg}})

	err := svc.CheckPendingOrderAllowed(orderriskcontract.CheckInput{
		IsGuest: true,
		RiskIP:  "1.2.3.4",
		Items: []orderriskcontract.OrderItem{
			{ProductID: 7, Quantity: 1},
			{ProductID: 7, Quantity: 1},
		},
	}, gate)
	if !errors.Is(err, orderriskcontract.ErrPendingProductQuantityLimit) {
		t.Fatalf("expected pending product quantity limit, got %v", err)
	}
}

func TestCheckPendingOrderAllowed_MemberUsesUserAndOptionalIP(t *testing.T) {
	cfg := testConfig()
	cfg.Member.MaxPendingOrdersPerUser = 5
	cfg.Member.MaxPendingOrdersPerIP = 3
	gate := &pendingGateStub{pendingByUser: 4, pendingMemberIP: 3}
	svc := NewService(Options{Settings: settingReaderStub{config: cfg}})

	err := svc.CheckPendingOrderAllowed(orderriskcontract.CheckInput{UserID: 9, RiskIP: "1.2.3.4"}, gate)
	if !errors.Is(err, orderriskcontract.ErrTooManyPendingOrders) {
		t.Fatalf("expected member IP pending limit, got %v", err)
	}
	if !reflect.DeepEqual(gate.lockedKeys, []string{"member:user:9", "member:ip:1.2.3.4"}) {
		t.Fatalf("unexpected member lock keys: %#v", gate.lockedKeys)
	}
}

func TestCheckOrderAllowed_GuestReturnsExpiryOverride(t *testing.T) {
	cfg := testConfig()
	cfg.Guest.PaymentExpireMinutes = 8
	svc := NewService(Options{Settings: settingReaderStub{config: cfg}})

	result, err := svc.CheckOrderAllowed(orderriskcontract.CheckInput{IsGuest: true, ClientIP: "1.2.3.4"})
	if err != nil {
		t.Fatal(err)
	}
	if result.PaymentExpireMinutes != 8 {
		t.Fatalf("expected guest expiry 8, got %d", result.PaymentExpireMinutes)
	}
}
