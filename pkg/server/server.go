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

// Run starts the HTTP server
func Run(port int, gitRunner *lib.GitRunner) error {
	app := fiber.New()

	// Serve Swagger UI
	app.Get("/swagger/*", swagger.New(swagger.Config{
		URL:         "/swagger/doc.json",
		DeepLinking: true,
		Title:       "gitignore.lol API Documentation",
	}))

	app.Use("/", filesystem.New(filesystem.Config{
		Root:       http.FS(web.LandingPageFiles),
		PathPrefix: "landing-page/dist",
		Browse:     true,
	}))

	app.Get("/documentation", func(c *fiber.Ctx) error {
		return c.Redirect("/")
	})

	// app.Get("/api/healthz", healthCheck)
	app.Get("/api/list", func(c *fiber.Ctx) error {
		return listTemplates(c, gitRunner)
	})
	app.Get("/api/*", func(c *fiber.Ctx) error {
		return getTemplates(c, gitRunner)
	})

	return app.Listen(fmt.Sprintf(":%d", port))
}
