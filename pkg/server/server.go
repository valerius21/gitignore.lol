package server

import (
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v3"

	lib "me.valerius/gitignore-lol/pkg/lib"
)

// //go:embed
// var landingPageFiles embed.FS

func Run(port int, gitRunner *lib.GitRunner) error {
	app := fiber.New()
	// app.Use(slogfiber.New(lib.Logger))

	app.Get("/", func(c fiber.Ctx) error {
		return c.SendString("hello world")
	})

	app.Get("/api/healthz", func(c fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	app.Get("/api/list", func(c fiber.Ctx) error {
		fileNames, err := gitRunner.ListFiles()
		if err != nil {
			lib.Logger.Error("List Files", "error", err)
			return c.SendStatus(fiber.StatusInternalServerError)
		}
		return c.JSON(&fiber.Map{
			"files": fileNames,
		})
	})

	app.Get("/api/*", func(c fiber.Ctx) error {
		params := c.Params("*")
		lib.Logger.Info(params)

		var res strings.Builder

		for i, name := range strings.Split(params, ",") {
			lib.Logger.Info("Request", "n", i, "param", name)
			content, err := gitRunner.GetFileContents(name)
			if err != nil {
				lib.Logger.Error("Param has no match", "param", name, "error", err)
				return c.SendStatus(fiber.StatusBadRequest)
			}
			_, err = res.WriteString(content)
			if err != nil {
				lib.Logger.Error("Could not write to string builder", "param", name, "error", err)
				return c.SendStatus(fiber.StatusInternalServerError)
			}
			res.WriteString("\n")
		}

		return c.SendString(res.String())
	})

	return app.Listen(fmt.Sprintf(":%d", port))
}
