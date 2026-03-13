package lib

import (
	"testing"
	"time"
)

func newTestEnhancedRateLimiter(t *testing.T, normalLimit, errorLimit int, windowDuration, blockDuration time.Duration, maxViolations, cleanupIntervalMs int) *EnhancedRateLimiter {
	t.Helper()

	limiter := NewEnhancedRateLimiter(normalLimit, errorLimit, 1, 1, maxViolations, cleanupIntervalMs)
	limiter.windowDuration = windowDuration
	limiter.blockDuration = blockDuration

	t.Cleanup(limiter.Stop)
	return limiter
}

func getIPStateSnapshot(limiter *EnhancedRateLimiter, ip string) (IPState, bool) {
	limiter.mu.RLock()
	defer limiter.mu.RUnlock()

	state, ok := limiter.ipStates[ip]
	if !ok {
		return IPState{}, false
	}

	return *state, true
}

func TestEnhancedRateLimiter_BasicRateLimit(t *testing.T) {
	limiter := newTestEnhancedRateLimiter(t, 3, 2, 100*time.Millisecond, 150*time.Millisecond, 2, 25)
	ip := "192.168.1.10"

	for i := 0; i < 3; i++ {
		if !limiter.IsAllowed(ip, "/api/list", ResponseTypeNormal) {
			t.Fatalf("request %d should be allowed", i+1)
		}
	}

	if limiter.IsAllowed(ip, "/api/list", ResponseTypeNormal) {
		t.Fatal("4th request should be blocked")
	}
}

func TestEnhancedRateLimiter_ScannerPathDetection(t *testing.T) {
	limiter := newTestEnhancedRateLimiter(t, 3, 2, 100*time.Millisecond, 150*time.Millisecond, 2, 25)

	tests := []struct {
		path string
		want bool
	}{
		{path: "/wp-admin", want: true},
		{path: "/.env", want: true},
		{path: "/phpmyadmin", want: true},
		{path: "/api/list", want: false},
		{path: "/health", want: false},
		{path: "/", want: false},
	}

	for _, tt := range tests {
		if got := limiter.IsScannerPath(tt.path); got != tt.want {
			t.Errorf("IsScannerPath(%q) = %v, want %v", tt.path, got, tt.want)
		}
	}
}

func TestEnhancedRateLimiter_IPBlocking(t *testing.T) {
	limiter := newTestEnhancedRateLimiter(t, 1, 1, 100*time.Millisecond, 200*time.Millisecond, 2, 25)
	ip := "192.168.1.11"

	if !limiter.IsAllowed(ip, "/api/list", ResponseTypeNormal) {
		t.Fatal("first request should be allowed")
	}

	if limiter.IsAllowed(ip, "/api/list", ResponseTypeNormal) {
		t.Fatal("second request should be blocked after exceeding limit")
	}

	if limiter.IsAllowed(ip, "/api/list", ResponseTypeNormal) {
		t.Fatal("blocked IP should remain blocked")
	}

	state, ok := getIPStateSnapshot(limiter, ip)
	if !ok {
		t.Fatal("expected IP state to exist")
	}

	if time.Until(state.BlockedUntil) <= 0 {
		t.Fatal("expected blocked IP to have a future unblock time")
	}
	if state.Violations != 1 {
		t.Fatalf("expected 1 violation, got %d", state.Violations)
	}
}

func TestEnhancedRateLimiter_ExponentialBackoff(t *testing.T) {
	limiter := newTestEnhancedRateLimiter(t, 1, 1, 30*time.Millisecond, 80*time.Millisecond, 2, 500)
	ip := "192.168.1.12"

	if !limiter.IsAllowed(ip, "/api/list", ResponseTypeNormal) {
		t.Fatal("first request should be allowed")
	}
	if limiter.IsAllowed(ip, "/api/list", ResponseTypeNormal) {
		t.Fatal("second request should trigger first block")
	}

	state, ok := getIPStateSnapshot(limiter, ip)
	if !ok {
		t.Fatal("expected IP state after first violation")
	}
	firstRemaining := time.Until(state.BlockedUntil)

	time.Sleep(110 * time.Millisecond)

	if !limiter.IsAllowed(ip, "/api/list", ResponseTypeNormal) {
		t.Fatal("request after first block should be allowed")
	}
	if limiter.IsAllowed(ip, "/api/list", ResponseTypeNormal) {
		t.Fatal("next request should trigger second block")
	}

	state, ok = getIPStateSnapshot(limiter, ip)
	if !ok {
		t.Fatal("expected IP state after second violation")
	}
	secondRemaining := time.Until(state.BlockedUntil)

	if state.Violations != 2 {
		t.Fatalf("expected 2 violations, got %d", state.Violations)
	}
	if secondRemaining <= firstRemaining+40*time.Millisecond {
		t.Fatalf("expected second block to be longer, got first=%v second=%v", firstRemaining, secondRemaining)
	}
}

func TestEnhancedRateLimiter_Cleanup(t *testing.T) {
	limiter := newTestEnhancedRateLimiter(t, 2, 2, 100*time.Millisecond, 100*time.Millisecond, 2, 25)
	ip := "192.168.1.13"

	if !limiter.IsAllowed(ip, "/api/list", ResponseTypeNormal) {
		t.Fatal("request should be allowed")
	}

	stats := limiter.GetStats()
	if stats["active_ips"] != 1 {
		t.Fatalf("expected 1 active IP before cleanup, got %v", stats["active_ips"])
	}

	time.Sleep(250 * time.Millisecond)

	stats = limiter.GetStats()
	if stats["active_ips"] != 0 {
		t.Fatalf("expected expired IP to be cleaned up, got %v active IPs", stats["active_ips"])
	}
}

func TestEnhancedRateLimiter_Stats(t *testing.T) {
	limiter := NewEnhancedRateLimiter(3, 2, 1, 1, 2, 25)
	t.Cleanup(limiter.Stop)

	if !limiter.IsAllowed("192.168.1.20", "/api/list", ResponseTypeNormal) {
		t.Fatal("first normal request should be allowed")
	}
	if !limiter.IsAllowed("192.168.1.20", "/api/list", ResponseTypeNormal) {
		t.Fatal("second normal request should be allowed")
	}
	if !limiter.IsAllowed("192.168.1.21", "/missing", ResponseType404) {
		t.Fatal("first error request should be allowed")
	}
	if !limiter.IsAllowed("192.168.1.21", "/missing", ResponseType404) {
		t.Fatal("second error request should be allowed")
	}
	if limiter.IsAllowed("192.168.1.21", "/missing", ResponseType404) {
		t.Fatal("third error request should be blocked")
	}

	stats := limiter.GetStats()

	if stats["active_ips"] != 2 {
		t.Fatalf("expected 2 active IPs, got %v", stats["active_ips"])
	}
	if stats["blocked_ips"] != 1 {
		t.Fatalf("expected 1 blocked IP, got %v", stats["blocked_ips"])
	}
	if stats["normal_requests"] != 2 {
		t.Fatalf("expected 2 normal requests, got %v", stats["normal_requests"])
	}
	if stats["error_requests"] != 2 {
		t.Fatalf("expected 2 error requests, got %v", stats["error_requests"])
	}
	if stats["normal_limit"] != 3 {
		t.Fatalf("expected normal_limit 3, got %v", stats["normal_limit"])
	}
	if stats["error_limit"] != 2 {
		t.Fatalf("expected error_limit 2, got %v", stats["error_limit"])
	}
	if stats["window_seconds"] != 1 {
		t.Fatalf("expected window_seconds 1, got %v", stats["window_seconds"])
	}
	if stats["block_minutes"] != 1 {
		t.Fatalf("expected block_minutes 1, got %v", stats["block_minutes"])
	}
}
