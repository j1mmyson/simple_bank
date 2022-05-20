package api

import (
	"github.com/gofiber/fiber/v2"
)

func main() {

	api := fiber.New()
	api.Get("/ping")
}

func pingpong(c *fiber.Ctx) error {

	return c.SendString("pong")
}
