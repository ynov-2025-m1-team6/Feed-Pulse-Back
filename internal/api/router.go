package api

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/api/handlers"
	"github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/api/handlers/Board"
	"github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/api/handlers/Feedback"
	"github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/api/handlers/auth"
	"github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/middleware"
)

func SetupRoutes(app *fiber.App) {
	// use logger middleware
	app.Use(logger.New(logger.Config{
		Format:     "[${time}] ${status} - ${method} ${path}\n",
		TimeFormat: "02-Jan-2006 15:04:05",
		TimeZone:   "Local",
	}))

	// use CORS middleware
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
		AllowMethods: "GET, POST, PUT, DELETE, OPTIONS",
	}))

	app.Get("/ping", handlers.PingHandler)

	api := app.Group("/api")

	// Authentication routes
	authGrp := api.Group("/auth")
	authGrp.Post("/login", auth.LoginHandler)
	authGrp.Post("/register", auth.RegisterHandler)
	authGrp.Get("/user", middleware.AuthRequired(), auth.UserInfoHandler)

	// Protected routes (require authentication)
	authGrp.Get("/logout", middleware.AuthRequired(), auth.LogoutHandler)

	// Feedback routes
	feedbackGrp := api.Group("/feedbacks")
	feedbackGrp.Post("/upload", middleware.AuthRequired(), Feedback.UploadFeedbackFileHandler)
	feedbackGrp.Post("/fetch", middleware.AuthRequired(), Feedback.FetchFeedbackHandler)
	feedbackGrp.Get("/analyses", middleware.AuthRequired(), Feedback.GetFeedbacksByUserIdHandler)

	boardGrp := api.Group("/board")
	boardGrp.Get("/metrics", middleware.AuthRequired(), Board.BoardMetricsHandler)

}
