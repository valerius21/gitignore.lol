// Package lib provides enhanced Fiber middleware for advanced rate limiting
package lib

import (
	"github.com/gofiber/fiber/v2"
)

// EnhancedRateLimitMiddleware creates a Fiber middleware with advanced scanner protection
func EnhancedRateLimitMiddleware(limiter *EnhancedRateLimiter) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if limiter == nil {
			return c.Next()
		}

		clientIP := c.IP()
		path := c.Path()

		// Pre-check: Block scanner paths immediately
		if limiter.IsScannerPath(path) {
			Logger.Warn("Blocked scanner path", "ip", clientIP, "path", path)
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Not Found",
			})
		}

		// Check rate limit for normal requests first
		if !limiter.IsAllowed(clientIP, path, ResponseTypeNormal) {
			Logger.Warn("Rate limit exceeded", "ip", clientIP, "path", path)
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error":   "Too Many Requests",
				"message": "Rate limit exceeded. Please try again later.",
			})
		}

		// Continue with request processing
		err := c.Next()

		// After request processing, check response status for additional rate limiting
		status := c.Response().StatusCode()
		if status == 404 || status >= 400 {
			responseType := ResponseType404
			if status >= 500 {
				responseType = ResponseTypeError
			}

			// Check error-specific rate limits
			if !limiter.IsAllowed(clientIP, path, responseType) {
				Logger.Warn("Error rate limit exceeded", "ip", clientIP, "status", status)
				// IP is now blocked, but we already processed the request
				// Future requests will be blocked by the pre-check
			}
		}

		return err
	}
}

// EnhancedStatsHandler provides detailed statistics for the enhanced rate limiter
func EnhancedStatsHandler(limiter *EnhancedRateLimiter) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if limiter == nil {
			return c.JSON(fiber.Map{
				"enhanced_rate_limiting": "disabled",
			})
		}

		stats := limiter.GetStats()
		return c.JSON(fiber.Map{
			"enhanced_rate_limiting": "enabled",
			"stats":                  stats,
			"scanner_protection":     "active",
		})
	}
}
