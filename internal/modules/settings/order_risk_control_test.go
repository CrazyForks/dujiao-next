package settings

import "testing"

func TestNormalizeOrderRiskControlConfig_Defaults(t *testing.T) {
	cfg := NormalizeOrderRiskControlConfig(OrderRiskControlConfig{})
	if cfg.MaxPendingOrdersPerUser != 0 {
		t.Fatalf("zero should be preserved (means no limit), got %d", cfg.MaxPendingOrdersPerUser)
	}
}

func TestNormalizeOrderRiskControlConfig_ClampValues(t *testing.T) {
	cfg := NormalizeOrderRiskControlConfig(OrderRiskControlConfig{
		MaxPendingOrdersPerUser:       -1,
		MaxPendingOrdersPerIP:         200,
		MaxPendingOrdersPerGuestEmail: 50,
		OrderRateLimit: OrderRateLimitConfig{
			WindowSeconds: 5,    // below min 10
			MaxRequests:   0,    // below min 1
			BlockSeconds:  -100, // below min 0
		},
	})
	if cfg.MaxPendingOrdersPerUser != 3 {
		t.Fatalf("expected clamped to default 3, got %d", cfg.MaxPendingOrdersPerUser)
	}
	if cfg.MaxPendingOrdersPerIP != 5 {
		t.Fatalf("expected clamped to default 5, got %d", cfg.MaxPendingOrdersPerIP)
	}
	if cfg.MaxPendingOrdersPerGuestEmail != 50 {
		t.Fatalf("expected 50 (valid), got %d", cfg.MaxPendingOrdersPerGuestEmail)
	}
	if cfg.OrderRateLimit.WindowSeconds != 60 {
		t.Fatalf("expected clamped window to 60, got %d", cfg.OrderRateLimit.WindowSeconds)
	}
	if cfg.OrderRateLimit.MaxRequests != 5 {
		t.Fatalf("expected clamped max_requests to 5, got %d", cfg.OrderRateLimit.MaxRequests)
	}
	if cfg.OrderRateLimit.BlockSeconds != 120 {
		t.Fatalf("expected clamped block to 120, got %d", cfg.OrderRateLimit.BlockSeconds)
	}
}

func TestNormalizeOrderRiskControlConfig_IPValidation(t *testing.T) {
	cfg := NormalizeOrderRiskControlConfig(OrderRiskControlConfig{
		IPBlacklist: []string{
			"1.2.3.4",         // valid IP
			"10.0.0.0/8",      // valid CIDR
			"invalid_ip",      // invalid - should be removed
			"",                // empty - should be removed
			"  192.168.1.1  ", // valid with whitespace
			"999.999.999.999", // invalid IP
			"abc/24",          // invalid CIDR
		},
	})
	expected := []string{"1.2.3.4", "10.0.0.0/8", "192.168.1.1"}
	if len(cfg.IPBlacklist) != len(expected) {
		t.Fatalf("expected %d IPs, got %d: %v", len(expected), len(cfg.IPBlacklist), cfg.IPBlacklist)
	}
	for i, ip := range expected {
		if cfg.IPBlacklist[i] != ip {
			t.Fatalf("expected IP[%d]=%q, got %q", i, ip, cfg.IPBlacklist[i])
		}
	}
}

func TestNormalizeOrderRiskControlConfig_EmailNormalization(t *testing.T) {
	cfg := NormalizeOrderRiskControlConfig(OrderRiskControlConfig{
		EmailBlacklist: []string{
			"  Spam@Example.COM  ",
			"",
			"test@test.com",
		},
	})
	if len(cfg.EmailBlacklist) != 2 {
		t.Fatalf("expected 2 emails, got %d: %v", len(cfg.EmailBlacklist), cfg.EmailBlacklist)
	}
	if cfg.EmailBlacklist[0] != "spam@example.com" {
		t.Fatalf("expected lowercased email, got %q", cfg.EmailBlacklist[0])
	}
}

func TestIsValidIPOrCIDR(t *testing.T) {
	tests := []struct {
		input string
		valid bool
	}{
		{"1.2.3.4", true},
		{"::1", true},
		{"10.0.0.0/8", true},
		{"192.168.0.0/16", true},
		{"fe80::/10", true},
		{"invalid", false},
		{"999.999.999.999", false},
		{"abc/24", false},
		{"", false},
	}
	for _, tc := range tests {
		if got := isValidIPOrCIDR(tc.input); got != tc.valid {
			t.Errorf("isValidIPOrCIDR(%q) = %v, want %v", tc.input, got, tc.valid)
		}
	}
}
