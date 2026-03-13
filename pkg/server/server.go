// Package server provides the HTTP server implementation
package server

import (
	"fmt"
	"io/fs"
	"strings"

	swaggo "github.com/gofiber/contrib/v3/swaggo"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	recovermw "github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/gofiber/fiber/v3/middleware/static"

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

func newApp(gitRunner *lib.GitRunner, rateLimiter *lib.MovingWindowLimiter, enhancedLimiter *lib.EnhancedRateLimiter, enableStats bool, corsOrigins string) (*fiber.App, error) {
	app := fiber.New()
	applyGlobalMiddleware(app, corsOrigins)

	landingPageFS, err := fs.Sub(web.LandingPageFiles, "landing-page/dist")
	if err != nil {
		return nil, fmt.Errorf("prepare static filesystem: %w", err)
	}

	registerHealthRoute(app)

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
	app.Get("/swagger/*", swaggo.HandlerDefault)

	// Serve static files (no rate limiting)
	app.Use("/", static.New("", static.Config{
		FS:     landingPageFS,
		Browse: true,
	}))

	// API routes with rate limiting
	apiGroup.Get("/list", func(c fiber.Ctx) error {
		return listTemplates(c, gitRunner)
	})
	apiGroup.Get("/*", func(c fiber.Ctx) error {
		return getTemplates(c, gitRunner)
	})

	// Rate limiter stats endpoint (no rate limiting, only if enabled)
	if enableStats {
		if enhancedLimiter != nil {
			app.Get("/stats", lib.EnhancedStatsHandler(enhancedLimiter))
		} else {
			app.Get("/stats", lib.RateLimitStatsHandler(rateLimiter))
		}
	}

	return app, nil
}

func applyGlobalMiddleware(app *fiber.App, corsOrigins string) {
	app.Use(recovermw.New())
	app.Use(cors.New(cors.Config{AllowOrigins: parseCorsOrigins(corsOrigins)}))
}

func parseCorsOrigins(corsOrigins string) []string {
	if corsOrigins == "" {
		return []string{"*"}
	}

	origins := strings.Split(corsOrigins, ",")
	for i := range origins {
		origins[i] = strings.TrimSpace(origins[i])
	}

	return origins
}

func registerHealthRoute(app *fiber.App) {
	app.Get("/health", func(c fiber.Ctx) error {
		return c.Status(fiber.StatusOK).JSON(HealthResponse{Status: "ok"})
	})
}

// Run starts the HTTP server and returns the app for graceful shutdown
func Run(port int, gitRunner *lib.GitRunner, rateLimiter *lib.MovingWindowLimiter, enhancedLimiter *lib.EnhancedRateLimiter, enableStats bool, corsOrigins string) (*fiber.App, error) {
	app, err := newApp(gitRunner, rateLimiter, enhancedLimiter, enableStats, corsOrigins)
	if err != nil {
		return nil, err
	}

	// Start listening in a goroutine
	go func() {
		if err := app.Listen(fmt.Sprintf(":%d", port)); err != nil {
			lib.Logger.Error("Server error", "error", err)
		}
	}()

	return app, nil
}
