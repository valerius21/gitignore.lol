// Package server provides the HTTP server implementation
package server

import (
	"fmt"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/gofiber/swagger"

	_ "me.valerius/gitignore-lol/docs"
	lib "me.valerius/gitignore-lol/pkg/lib"
	"me.valerius/gitignore-lol/web"
)

// TemplateResponse represents the response for the list endpoint
type TemplateResponse struct {
	// List of available gitignore templates
	Files []string `json:"files" example:"[\"go\",\"node\",\"python\"]"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	// Error message
	Error string `json:"error" example:"Template not found"`
}

// HealthResponse represents a health check response
type HealthResponse struct {
	// Service status
	Status string `json:"status" example:"OK"`
}

// Run starts the HTTP server and returns the app for graceful shutdown
func Run(port int, gitRunner *lib.GitRunner, rateLimiter *lib.MovingWindowLimiter, enhancedLimiter *lib.EnhancedRateLimiter) (*fiber.App, error) {
	app := fiber.New()

	// Apply rate limiting middleware to API routes only
	apiGroup := app.Group("/api")
	if enhancedLimiter != nil {
		lib.Logger.Info("Enhanced rate limiting enabled with scanner protection")
		apiGroup.Use(lib.EnhancedRateLimitMiddleware(enhancedLimiter))
	} else if rateLimiter != nil {
		lib.Logger.Info("Basic rate limiting enabled", "max_requests", rateLimiter.GetStats()["max_requests"], "window_seconds", rateLimiter.GetStats()["window_seconds"])
		apiGroup.Use(lib.RateLimitMiddleware(rateLimiter))
	} else {
		lib.Logger.Info("Rate limiting disabled")
	}

	// Serve Swagger UI (no rate limiting)
	app.Get("/swagger/*", swagger.New(swagger.Config{
		URL:         "/swagger/doc.json",
		DeepLinking: true,
		Title:       "gitignore.lol API Documentation",
	}))

	// Serve static files (no rate limiting)
	app.Use("/", filesystem.New(filesystem.Config{
		Root:       http.FS(web.LandingPageFiles),
		PathPrefix: "landing-page/dist",
		Browse:     true,
	}))

	app.Get("/documentation", func(c *fiber.Ctx) error {
		return c.Redirect("/")
	})

	// API routes with rate limiting
	apiGroup.Get("/list", func(c *fiber.Ctx) error {
		return listTemplates(c, gitRunner)
	})
	apiGroup.Get("/*", func(c *fiber.Ctx) error {
		return getTemplates(c, gitRunner)
	})

	// Rate limiter stats endpoint (no rate limiting)
	if enhancedLimiter != nil {
		app.Get("/stats", lib.EnhancedStatsHandler(enhancedLimiter))
	} else {
		app.Get("/stats", lib.RateLimitStatsHandler(rateLimiter))
	}

	// Start listening in a goroutine
	go func() {
		if err := app.Listen(fmt.Sprintf(":%d", port)); err != nil {
			lib.Logger.Error("Server error", "error", err)
		}
	}()

	return app, nil
}
