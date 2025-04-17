package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/api"
	"log"
)

func main() {
	app := fiber.New(fiber.Config{
		ErrorHandler: customErrorHandler,
	})

	api.SetupRoutes(app)

	log.Printf("Starting server on localhost:3000\n")
	log.Fatal(app.Listen(":3000"))
}

// customErrorHandler provides better error responses
func customErrorHandler(c *fiber.Ctx, err error) error {
	// Default 500 status code
	code := fiber.StatusInternalServerError

	// Retrieve the custom status code if it's a fiber.*Error
	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
	}

	// Set Content-Type: application/json; charset=utf-8
	c.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSONCharsetUTF8)

	// Return status code with error message
	return c.Status(code).JSON(fiber.Map{
		"error": err.Error(),
	})
}
