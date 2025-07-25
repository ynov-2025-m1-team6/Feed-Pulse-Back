package api

import (
	"github.com/getsentry/sentry-go"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/tot0p/env"
	"github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/api/handlers"
	"github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/api/handlers/Board"
	"github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/api/handlers/Feedback"
	"github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/api/handlers/auth"
	"github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/middleware"
)

func SetupRoutes(app *fiber.App) error {
	// setup sentry error tracking
	err := sentry.Init(sentry.ClientOptions{
		Dsn:              env.Get("SENTRY_DSN"),
		Environment:      env.Get("ENV"),
		EnableTracing:    true,
		TracesSampleRate: 1.0, // Adjust the sample rate as needed
		SendDefaultPII:   true,
	})
	if err != nil {
		return err
	}

	// Initialize Sentry error handling middleware
	sentryHandler := middleware.SentryFiber(middleware.Options{
		Repanic:         false,
		WaitForDelivery: true,
	})
	app.Use(sentryHandler)

	// use logger middleware
	app.Use(logger.New(logger.Config{
		Format:     "[${time}] ${status} - ${method} ${path}\n",
		TimeFormat: "02-Jan-2006 15:04:05",
		TimeZone:   "Local",
	}))
	allowedOrigins := env.Get("ALLOWED_ORIGINS")
	if allowedOrigins == "" {
		allowedOrigins = "*"
	}

	// use CORS middleware
	app.Use(cors.New(cors.Config{
		AllowOrigins: allowedOrigins,
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
		AllowMethods: "GET, POST, PUT, DELETE, OPTIONS",
	}))

	app.All("/foo", func(c *fiber.Ctx) error {
		panic("y tho")
	})

	app.Get("/ping", handlers.PingHandler) // done

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
	feedbackGrp.Get("/", Feedback.GetAllFeedbacksHandler)
	feedbackGrp.Post("/upload", middleware.AuthRequired(), Feedback.UploadFeedbackFileHandler)
	feedbackGrp.Post("/fetch", middleware.AuthRequired(), Feedback.FetchFeedbackHandler)
	feedbackGrp.Get("/analyses", middleware.AuthRequired(), Feedback.GetFeedbacksByUserIdHandler)

	boardGrp := api.Group("/board")
	boardGrp.Get("/metrics", middleware.AuthRequired(), Board.BoardMetricsHandler)
	return nil
}
