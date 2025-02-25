//go:build !test
// +build !test

// @title           gitignore.lol API
// @version         1.0
// @description     A service to generate .gitignore files for your projects. An implementation insprired by the previously known gitignore.io.

// @contact.name   Project URL
// @contact.url    https://github.com/valerius21/gitignore.lol

// @license.name  MIT
// @license.url   https://github.com/valerius21/gitignore.lol/blob/main/LICENSE

// @host      gitignore.lol
// @BasePath  /
// @schemes   https http
package server

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/gofiber/swagger" // swagger handler

	_ "me.valerius/gitignore-lol/docs"
	lib "me.valerius/gitignore-lol/pkg/lib"
	"me.valerius/gitignore-lol/web"
)

func Run(port int, gitRunner *lib.GitRunner) error {
	app := fiber.New()
	// app.Use(slogfiber.New(lib.Logger))

	app.Use("/", filesystem.New(filesystem.Config{
		Root:       http.FS(web.LandingPageFiles),
		PathPrefix: "landing-page/dist",
		Browse:     true,
	}))

	app.Get("/documentation", func(c *fiber.Ctx) error {
		return c.Redirect("/")
	})

	app.Get("/swagger/*", swagger.New(swagger.Config{
		URL: "/swagger/doc.json",
	}))

	// @Summary Check if the serivce is healthy
	// @Description returns 200, if the server is available
	// @Tags healthcheck
	// @Success 200 Service healthy
	// @Router /api/healthz [get]
	app.Get("/api/healthz", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	// @Summary Get available templates
	// @Description Returns a list of all available .gitignore templates
	// @Tags templates
	// @Accept json
	// @Produce json
	// @Success 200 {object} map[string]interface{} "List of available templates"
	// @Failure 500 {object} string "Internal Server Error"
	// @Router /api/list [get]
	// List handles the root endpoint request.
	// It returns a JSON response containing available .gitignore templates.
	// The response is cached to improve performance.
	// Responds:
	//   - 200: JSON response with available templates
	//   - 500: Internal server error if cache or repository operations fail
	app.Get("/api/list", func(c *fiber.Ctx) error {
		fileNames, err := gitRunner.ListFiles()
		if err != nil {
			lib.Logger.Error("List Files", "error", err)
			return c.SendStatus(fiber.StatusInternalServerError)
		}
		return c.JSON(&fiber.Map{
			"files": fileNames,
		})
	})

	// @Summary Get gitignore templates
	// @Description Returns combined .gitignore file for specified templates
	// @Tags templates
	// @Accept json
	// @Produce text/plain
	// @Param templateList path string true "Comma-separated list of templates (e.g., go,node,python)"
	// @Success 200 {string} string "Combined .gitignore file content"
	// @Failure 404 {string} string "Template not found"
	// @Failure 500 {string} string "Internal Server Error"
	// @Router /api/{templateList} [get]
	// Templates handles the request for one or more .gitignore templates.
	// It accepts a comma-separated list of template names in the URL parameter "templateList".
	// The templates are fetched from the repository, combined, and returned as a single string.
	// Example: /api/go,node,python will return a combined .gitignore file for Go, Node.js, and Python.
	app.Get("/api/*", func(c *fiber.Ctx) error {
		params := c.Params("*")
		lib.Logger.Info(params)

		var res strings.Builder

		for i, name := range strings.Split(params, ",") {
			lib.Logger.Info("Request", "n", i, "param", name)
			content, err := gitRunner.GetFileContents(name)
			if err != nil {
				lib.Logger.Error("Param has no match", "param", name, "error", err)
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error": fmt.Sprintf("Template '%s' not found", name),
				})
			}
			_, err = res.WriteString(content)
			if err != nil {
				lib.Logger.Error("Could not write to string builder", "param", name, "error", err)
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error": "Failed to process template content",
				})
			}
			res.WriteString("\n")
		}

		return c.SendString(res.String())
	})

	return app.Listen(fmt.Sprintf(":%d", port))
}
