package lib

import (
	"testing"
	"time"
)

func TestMovingWindowLimiter_BasicRateLimit(t *testing.T) {
	// Create a rate limiter: 3 requests per 1 second window
	limiter := NewMovingWindowLimiter(3, 1, 100)
	defer limiter.Stop()

	testIP := "192.168.1.1"

	// First 3 requests should be allowed
	for i := 0; i < 3; i++ {
		if !limiter.IsAllowed(testIP) {
			t.Errorf("Request %d should be allowed", i+1)
		}
	}

	// 4th request should be blocked
	if limiter.IsAllowed(testIP) {
		t.Error("4th request should be blocked")
	}

	// Wait for window to pass and try again
	time.Sleep(1100 * time.Millisecond)

	// Should be allowed again after window passes
	if !limiter.IsAllowed(testIP) {
		t.Error("Request should be allowed after window expiration")
	}
}

func TestMovingWindowLimiter_DifferentIPs(t *testing.T) {
	// Create a rate limiter: 2 requests per 1 second window
	limiter := NewMovingWindowLimiter(2, 1, 100)
	defer limiter.Stop()

	ip1 := "192.168.1.1"
	ip2 := "192.168.1.2"

	// Each IP should have independent limits
	for i := 0; i < 2; i++ {
		if !limiter.IsAllowed(ip1) {
			t.Errorf("IP1 request %d should be allowed", i+1)
		}
		if !limiter.IsAllowed(ip2) {
			t.Errorf("IP2 request %d should be allowed", i+1)
		}
	}

	// 3rd request from each IP should be blocked
	if limiter.IsAllowed(ip1) {
		t.Error("IP1 3rd request should be blocked")
	}
	if limiter.IsAllowed(ip2) {
		t.Error("IP2 3rd request should be blocked")
	}
}

func TestMovingWindowLimiter_SlidingWindow(t *testing.T) {
	// Create a rate limiter: 2 requests per 2 second window
	limiter := NewMovingWindowLimiter(2, 2, 100)
	defer limiter.Stop()

	testIP := "192.168.1.1"

	// Make 2 requests immediately
	if !limiter.IsAllowed(testIP) {
		t.Error("First request should be allowed")
	}
	if !limiter.IsAllowed(testIP) {
		t.Error("Second request should be allowed")
	}

	// 3rd request should be blocked
	if limiter.IsAllowed(testIP) {
		t.Error("Third request should be blocked")
	}

	// Wait 1 second (half the window)
	time.Sleep(1100 * time.Millisecond)

	// Should still be blocked (only 1 second passed out of 2 second window)
	if limiter.IsAllowed(testIP) {
		t.Error("Request should still be blocked")
	}

	// Wait another second (total 2+ seconds)
	time.Sleep(1100 * time.Millisecond)

	// Now should be allowed again
	if !limiter.IsAllowed(testIP) {
		t.Error("Request should be allowed after full window expiration")
	}
}

func TestMovingWindowLimiter_Stats(t *testing.T) {
	limiter := NewMovingWindowLimiter(5, 60, 100)
	defer limiter.Stop()

	// Make some requests
	limiter.IsAllowed("192.168.1.1")
	limiter.IsAllowed("192.168.1.1")
	limiter.IsAllowed("192.168.1.2")

	stats := limiter.GetStats()

	if stats["active_ips"] != 2 {
		t.Errorf("Expected 2 active IPs, got %v", stats["active_ips"])
	}

	if stats["total_requests"] != 3 {
		t.Errorf("Expected 3 total requests, got %v", stats["total_requests"])
	}

	if stats["max_requests"] != 5 {
		t.Errorf("Expected max_requests to be 5, got %v", stats["max_requests"])
	}

	if stats["window_seconds"] != 60 {
		t.Errorf("Expected window_seconds to be 60, got %v", stats["window_seconds"])
	}
}

func TestMovingWindowLimiter_Cleanup(t *testing.T) {
	// Create a rate limiter with fast cleanup for testing
	limiter := NewMovingWindowLimiter(3, 1, 50) // 50ms cleanup interval
	defer limiter.Stop()

	testIP := "192.168.1.1"

	// Make a request
	limiter.IsAllowed(testIP)

	// Verify IP is tracked
	stats := limiter.GetStats()
	if stats["active_ips"] != 1 {
		t.Errorf("Expected 1 active IP, got %v", stats["active_ips"])
	}

	// Wait for cleanup to run multiple times
	time.Sleep(1200 * time.Millisecond)

	// Verify expired entries are cleaned up
	stats = limiter.GetStats()
	if stats["active_ips"] != 0 {
		t.Errorf("Expected 0 active IPs after cleanup, got %v", stats["active_ips"])
	}
}
