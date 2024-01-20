package web

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/contrib/fiberzerolog"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cache"
	"github.com/gofiber/fiber/v2/middleware/healthcheck"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/storage/memory/v2"
	gfUtils "github.com/gofiber/utils/v2"
	"github.com/valerius21/gitignore.lol/pkg/repository"
	"github.com/valerius21/gitignore.lol/pkg/utils"
)

type WebServer struct {
	App *fiber.App
}

var DefaultWebServer WebServer

const availableFilesKey = "availableFiles"

func init() {
	// Intianciate Fiber
	app := fiber.New()

	// init repo
	repo := repository.DefaultRepository

	// init logging
	logger := utils.InitLogger()

	// init store for caching
	store := memory.New()

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
		resp, err := store.Get(availableFilesKey)
		if err != nil {
			logger.Error().Err(err).Msg("failed to get cache")
			return c.SendStatus(500)
		}

		if resp != nil {
			logger.Debug().Msg("cache hit")
			return c.JSON(resp)
		}

		logger.Debug().Msg("cache miss")

		files, err := repo.GetAvailabeIgnoreFiles()
		if err != nil {
			logger.Error().Err(err).Msg("failed to get available files")
			return c.SendStatus(500)
		}

		message := make(map[string]interface{})

		message["message"] = "Welcome to gitignore.lol. Here are the available templates. " +
			"This endpoint may change in the future." +
			"You can use them by appending them to the URL. Example: https://gitignore.lol/go,node,python"

		// rename the files to lowercase
		renamed := make([]string, 0)
		for _, file := range files {
			renamed = append(renamed, strings.ToLower(file))
		}

		message["templates"] = renamed

		// cache the response
		respBytes, err := json.Marshal(message)

		if err != nil {
			logger.Error().Err(err).Msg("failed to marshal response")
			return c.SendStatus(500)
		}

		err = store.Set(availableFilesKey, respBytes, 0)

		if err != nil {
			logger.Error().Err(err).Msg("failed to set cache")
			return c.SendStatus(500)
		}

		return c.JSON(message)
	})

	app.Get("/:template", func(c *fiber.Ctx) error {
		result := gfUtils.CopyString(c.Params("template"))
		logger.Debug().Str("template", result).Msg("got template: " + result)

		// sanitize the result
		result = strings.ReplaceAll(result, " ", ",")

		// split the result by comma
		templates := make([]string, 0)
		if strings.Contains(result, ",") {
			templates = strings.Split(result, ",")
		} else {
			templates = append(templates, result)
		}

		repo.UpdateRepo()

		ignores := make(map[string]string)

		for _, result := range templates {
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

			ignores[result] = content
		}

		// combine the contents of the files
		content := fmt.Sprintf("# ++++ gitignore.io/%s ++++\n", result)
		for key, value := range ignores {
			content += fmt.Sprintf("# ==== %s ====\n", strings.ReplaceAll(key, "# ", ""))
			content += value
		}

		return c.SendString(content)
	})

	DefaultWebServer = WebServer{
		App: app,
	}
}
