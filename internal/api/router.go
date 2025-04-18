package api

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/api/handlers"
	"github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/api/handlers/auth"
)

func SetupRoutes(app *fiber.App) {
	app.Use(logger.New(logger.Config{
		Format:     "[${time}] ${status} - ${method} ${path}\n",
		TimeFormat: "02-Jan-2006 15:04:05",
		TimeZone:   "Local",
	}))

	app.Get("/ping", handlers.PingHandler)

	api := app.Group("/api")

	// Authentication routes
	authGrp := api.Group("/auth")
	authGrp.Post("/login", auth.LoginHandler)
	authGrp.Post("/register", auth.RegisterHandler)
	authGrp.Get("/logout", auth.LogoutHandler)
}
