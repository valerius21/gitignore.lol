package web

import (
	"time"

	"github.com/gofiber/contrib/fiberzerolog"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cache"
	"github.com/gofiber/fiber/v2/middleware/healthcheck"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/valerius21/gitignore.lol/pkg/utils"
)

type WebServer struct {
	App *fiber.App
}

func (ws *WebServer) Init() {
	// Intianciate Fiber
	app := fiber.New()

	// init logging
	logger := utils.InitLogger()

	app.Use(fiberzerolog.New(fiberzerolog.Config{
		Logger: &logger,
	}))

	// Set up Rate Limiter
	app.Use(limiter.New(limiter.Config{
		Max:               20,
		Expiration:        1 * time.Second,
		LimiterMiddleware: limiter.SlidingWindow{},
	}))

	// TODO: Set up cache
	// server-side caching
	// client-side caching
	app.Use(cache.New())

	// Healthcheck endpoint
	app.Use(healthcheck.New())

	// Download Repository

	// static files sharing
	app.Static("/", "/tmp/gitignore")

	// Handlers
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("")
	})

	ws.App = app
}
