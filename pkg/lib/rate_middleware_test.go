package lib

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"
)

func TestRateLimitMiddleware_AllowsRequests(t *testing.T) {
	// Create a rate limiter: 5 requests per 1 second window
	limiter := NewMovingWindowLimiter(5, 1, 100)
	defer limiter.Stop()

	// Create Fiber app with rate limit middleware
	app := fiber.New()
	app.Use(RateLimitMiddleware(limiter))

	// Add a simple test route
	app.Get("/test", func(c fiber.Ctx) error {
		return c.SendString("OK")
	})

	// First request should be allowed (200)
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Forwarded-For", "192.168.1.1")
	resp, err := app.Test(req)

	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

func TestRateLimitMiddleware_BlocksExcess(t *testing.T) {
	// Create a rate limiter: 2 requests per 1 second window
	limiter := NewMovingWindowLimiter(2, 1, 100)
	defer limiter.Stop()

	// Create Fiber app with rate limit middleware
	app := fiber.New()
	app.Use(RateLimitMiddleware(limiter))

	// Add a simple test route
	app.Get("/test", func(c fiber.Ctx) error {
		return c.SendString("OK")
	})

	testIP := "192.168.1.1"

	// Make 2 allowed requests
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("X-Forwarded-For", testIP)
		resp, err := app.Test(req)

		if err != nil {
			t.Fatalf("Request %d failed: %v", i+1, err)
		}

		if resp.StatusCode != fiber.StatusOK {
			t.Errorf("Request %d: expected status 200, got %d", i+1, resp.StatusCode)
		}
	}

	// 3rd request should be blocked (429)
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Forwarded-For", testIP)
	resp, err := app.Test(req)

	if err != nil {
		t.Fatalf("3rd request failed: %v", err)
	}

	if resp.StatusCode != fiber.StatusTooManyRequests {
		t.Errorf("Expected status 429, got %d", resp.StatusCode)
	}
}

func TestRateLimitMiddleware_NilLimiter(t *testing.T) {
	// Create Fiber app with nil limiter (disabled)
	app := fiber.New()
	app.Use(RateLimitMiddleware(nil))

	// Add a simple test route
	app.Get("/test", func(c fiber.Ctx) error {
		return c.SendString("OK")
	})

	// All requests should pass through (200)
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		resp, err := app.Test(req)

		if err != nil {
			t.Fatalf("Request %d failed: %v", i+1, err)
		}

		if resp.StatusCode != fiber.StatusOK {
			t.Errorf("Request %d: expected status 200, got %d", i+1, resp.StatusCode)
		}
	}
}

func TestRateLimitStatsHandler(t *testing.T) {
	// Create a rate limiter
	limiter := NewMovingWindowLimiter(10, 60, 100)
	defer limiter.Stop()

	// Create Fiber app with stats handler
	app := fiber.New()
	app.Get("/stats", RateLimitStatsHandler(limiter))

	// Make a request to the stats endpoint
	req := httptest.NewRequest("GET", "/stats", nil)
	resp, err := app.Test(req)

	if err != nil {
		t.Fatalf("Stats request failed: %v", err)
	}

	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	// Verify response is JSON with expected fields
	contentType := resp.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "application/json") {
		t.Errorf("Expected Content-Type application/json, got %s", contentType)
	}
}
