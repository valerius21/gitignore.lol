package server

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
)

func Run(port int) error {

	app := fiber.New()

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, world")
	})

	return app.Listen(fmt.Sprintf(":%d", port))
}
