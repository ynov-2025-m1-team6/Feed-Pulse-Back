package api

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/api/handlers"
	"github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/api/handlers/Feedback"
)

func SetupRoutes(app *fiber.App) {
	app.Use(logger.New(logger.Config{
		Format:     "[${time}] ${status} - ${method} ${path}\n",
		TimeFormat: "02-Jan-2006 15:04:05",
		TimeZone:   "Local",
	}))

	app.Get("/ping", handlers.PingHandler)

	// Feedback routes
	feedbacks := app.Group("/feedbacks")
	feedbacks.Get("/", Feedback.GetAllFeedbacksHandler)
	feedbacks.Get("/:id", Feedback.GetFeedbackByIDHandler)
	feedbacks.Post("/", Feedback.CreateFeedbackHandler)
	feedbacks.Put("/:id", Feedback.UpdateFeedbackHandler)
	feedbacks.Delete("/:id", Feedback.DeleteFeedbackHandler)
	feedbacks.Get("/channel/:channel", Feedback.GetFeedbacksByChannelHandler)
}
