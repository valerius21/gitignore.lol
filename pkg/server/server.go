package server

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	slogfiber "github.com/samber/slog-fiber"

	lib "me.valerius/gitignore-lol/pkg/lib"
)

func Run(port int, gitRunner *lib.GitRunner) error {
	app := fiber.New()
	app.Use(slogfiber.New(lib.Logger))

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, world")
	})

	app.Get("/api/list", func(c *fiber.Ctx) error {
		fileNames, err := gitRunner.ListFiles()
		if err != nil {
			lib.Logger.Error("List Files", "error", err)
			return c.SendStatus(500)
		}
		return c.JSON(&fiber.Map{
			"files": fileNames,
		})
	})

	return app.Listen(fmt.Sprintf(":%d", port))
}
