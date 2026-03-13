// Package lib provides Fiber middleware for rate limiting
package lib

import (
	"github.com/gofiber/fiber/v2"
)

// RateLimitMiddleware creates a Fiber middleware that applies rate limiting
func RateLimitMiddleware(limiter *MovingWindowLimiter) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Skip rate limiting if disabled
		if limiter == nil {
			return c.Next()
		}

		// Get client IP
		clientIP := c.IP()

		// Check if request is allowed
		if !limiter.IsAllowed(clientIP) {
			Logger.Warn("Rate limit exceeded", "ip", clientIP)
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error":   "Too Many Requests",
				"message": "Rate limit exceeded. Please try again later.",
			})
		}

		// Request is allowed, continue
		return c.Next()
	}
}

// RateLimitStatsHandler provides an endpoint to view rate limiter statistics
func RateLimitStatsHandler(limiter *MovingWindowLimiter) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if limiter == nil {
			return c.JSON(fiber.Map{
				"rate_limiting": "disabled",
			})
		}

		stats := limiter.GetStats()
		return c.JSON(fiber.Map{
			"rate_limiting": "enabled",
			"stats":         stats,
		})
	}
}
