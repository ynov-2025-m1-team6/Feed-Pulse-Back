package handlers

import (
	"github.com/gofiber/fiber/v2"
)

func PingHandler(c *fiber.Ctx) error {
	// Respond with a 200 OK status and a simple message
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Pong",
	})
}
