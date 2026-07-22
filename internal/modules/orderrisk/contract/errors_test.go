package contract

import "testing"

func TestRateLimitedErrorContract(t *testing.T) {
	err := &RateLimitedError{RetryAfter: 60}
	if err.Error() != ErrOrderRateLimited.Error() {
		t.Fatalf("expected error message match")
	}
	if !err.Is(ErrOrderRateLimited) {
		t.Fatal("expected rate-limited error identity")
	}
	if retryAfter := GetRetryAfter(err); retryAfter != 60 {
		t.Fatalf("expected retry-after 60, got %d", retryAfter)
	}
	if retryAfter := GetRetryAfter(ErrIPBlacklisted); retryAfter != 0 {
		t.Fatalf("expected zero retry-after for unrelated error, got %d", retryAfter)
	}
}
