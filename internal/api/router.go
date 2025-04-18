package api

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/api/handlers"
	"github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/api/handlers/Analysis"
	"github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/api/handlers/Feedback"
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

	// Feedback routes
	feedbacks := api.Group("/feedbacks")
	feedbacks.Get("/", Feedback.GetAllFeedbacksHandler)
	feedbacks.Get("/:id", Feedback.GetFeedbackByIDHandler)
	feedbacks.Post("/", Feedback.CreateFeedbackHandler)
	feedbacks.Put("/:id", Feedback.UpdateFeedbackHandler)
	feedbacks.Delete("/:id", Feedback.DeleteFeedbackHandler)
	feedbacks.Get("/channel/:channel", Feedback.GetFeedbacksByChannelHandler)

	// Analyses routes
	analyses := api.Group("/analyses")
	analyses.Get("/", Analysis.GetAllAnalysesHandler)
	analyses.Get("/:id", Analysis.GetAnalysisByIDHandler)
	analyses.Post("/", Analysis.AddAnalysisHandler)
	analyses.Put("/:id", Analysis.UpdateAnalysisHandler)
	analyses.Delete("/:id", Analysis.DeleteAnalysisHandler)
	analyses.Get("/feedback/:feedback_id", Analysis.GetAnalysisByFeedbackIDHandler)
	analyses.Get("/topic/:topic", Analysis.GetAnalysesByTopicHandler)
	analyses.Get("/sentiment", Analysis.GetAnalysesBySentimentRangeHandler)
}
