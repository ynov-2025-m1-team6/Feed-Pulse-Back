package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/tot0p/env"
	"github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/api"
	"github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/database"
	"github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/models"
	"github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/sessionManager"
	"log"
	"os"
	"time"
)

func main() {
	app := fiber.New(fiber.Config{
		ErrorHandler: customErrorHandler,
	})
	err := env.Load()
	if err != nil {
		if !os.IsNotExist(err) {
			log.Fatalf("Failed to load environment variables: %v", err)
		}
	}

	// Initialize the database connection
	err = database.InitDatabase(env.Get("DB_USER"), env.Get("DB_PASSWORD"), env.Get("DB_NAME"), env.Get("DB_HOST"), env.Get("DB_PORT"), env.Get("DB_SSLMODE"))
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Initialize and automigrate the models
	err = models.InitModels()
	if err != nil {
		log.Fatalf("Failed to initialize models: %v", err)
	}

	// Initialize the SessionManager
	if env.Get("SECRET_KEY") == "" {
		log.Fatal("SECRET_KEY environment variable is not set")
	}
	sessionManager.InitSessionManager(env.Get("SECRET_KEY"), 3*time.Hour)

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
