package web

import (
	"time"

	"github.com/gofiber/contrib/fiberzerolog"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cache"
	"github.com/gofiber/fiber/v2/middleware/healthcheck"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	gfUtils "github.com/gofiber/utils/v2"
	"github.com/valerius21/gitignore.lol/pkg/repository"
	"github.com/valerius21/gitignore.lol/pkg/utils"
)

type WebServer struct {
	App *fiber.App
}

var DefaultWebServer WebServer

func init() {
	// Intianciate Fiber
	app := fiber.New()

	// init repo
	repo := repository.DefaultRepository

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

	// client-side caching
	app.Use(cache.New())

	// Healthcheck endpoint
	app.Use(healthcheck.New())

	// static files sharing
	app.Static("/", utils.DefaultRepoInfo.LocalPath)

	// Handlers
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello there")
	})

	app.Get("/:template", func(c *fiber.Ctx) error {
		result := gfUtils.CopyString(c.Params("template"))

		logger.Debug().Str("template", result).Msg("got template: " + result)
		repo.UpdateRepo()
		fileName, err := repo.GetMappedFileName(result)
		if err != nil {
			logger.Error().Err(err).Msg("failed to get mapped file name")
			return c.SendStatus(500)
		}

		// read file contents from fileName
		fileContents, err := repo.GetFileContent(fileName, result)

		if err != nil {
			logger.Error().Err(err).Msg("failed to get file contents")
			return c.SendStatus(500)
		}

		content := string(fileContents)

		if content == "" {
			return c.SendStatus(404)
		}

		return c.SendString(content)
	})

	DefaultWebServer = WebServer{
		App: app,
	}
}
