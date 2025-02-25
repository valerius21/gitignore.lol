package server

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/gofiber/fiber/v2"
	lib "me.valerius/gitignore-lol/pkg/lib"
)

// @title           gitignore.lol API
// @version         1.0
// @description     A service to generate .gitignore files for your projects. An implementation inspired by the previously known gitignore.io.

// @license.name  MIT
// @license.url   https://github.com/valerius21/gitignore.lol/blob/main/LICENSE

// @host      localhost:4444
// @BasePath  /
// @schemes   http

// ListTemplates godoc
// @Summary      List available templates
// @Description  Returns a list of all available .gitignore templates
// @Tags         templates
// @Produce      json
// @Success      200  {object}  server.TemplateResponse  "List of available templates"
// @Failure      500  {object}  server.ErrorResponse     "Internal server error"
// @Router       /api/list [get]
func listTemplates(c *fiber.Ctx, gitRunner *lib.GitRunner) error {
	fileNames, err := gitRunner.ListFiles()
	if err != nil {
		lib.Logger.Error("List Files", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error: "Internal Server Error",
		})
	}
	return c.Status(fiber.StatusOK).JSON(TemplateResponse{
		Files: fileNames,
	})
}

// GetTemplates godoc
// @Summary      Get gitignore templates
// @Description  Returns combined .gitignore file for specified templates
// @Tags         templates
// @Param        templateList  path      string  true  "Comma-separated list of templates (e.g., go,node,python)"
// @Produce      text/plain
// @Success      200  {string}  string                "Combined .gitignore file content"
// @Failure      400  {object}  server.ErrorResponse  "Template not found"
// @Failure      500  {object}  server.ErrorResponse  "Internal server error"
// @Router       /api/{templateList} [get]
func getTemplates(c *fiber.Ctx, gitRunner *lib.GitRunner) error {
	params := c.Params("*")
	decodedParams, err := url.QueryUnescape(params)
	if err != nil {
		lib.Logger.Error("Failed to decode URL parameters", "error", err)
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error: "Invalid URL encoding in parameters",
		})
	}
	lib.Logger.Info(decodedParams)

	var res strings.Builder

	for i, name := range strings.Split(decodedParams, ",") {
		lib.Logger.Info("Request", "n", i, "param", name)
		content, err := gitRunner.GetFileContents(name)
		if err != nil {
			lib.Logger.Error("Param has no match", "param", name, "error", err)
			return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
				Error: fmt.Sprintf("Template '%s' not found", name),
			})
		}
		_, err = res.WriteString(content)
		if err != nil {
			lib.Logger.Error("Could not write to string builder", "param", name, "error", err)
			return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
				Error: "Failed to process template content",
			})
		}
		res.WriteString("\n")
	}

	return c.Status(fiber.StatusOK).SendString(res.String())
}
