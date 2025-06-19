package Feedback

import (
	"github.com/getsentry/sentry-go"
	"github.com/gofiber/fiber/v2"
	"github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/database/Feedback"
	feedbackModel "github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/models/Feedback"
)

// GetAllFeedbacksHandler godoc
// @Summary Get all feedbacks
// @Description Fetch all feedbacks from the database
// @Tags Feedback
// @Produce json
// @Success 200 {object} Feedback.Feedback "List of feedbacks"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /api/feedbacks [get]
func GetAllFeedbacksHandler(c *fiber.Ctx) error {
	feedbacks, err := Feedback.GetAllFeedbacks()
	if err != nil {
		sentry.CaptureEvent(&sentry.Event{
			Message: "Failed to fetch feedbacks",
			Level:   sentry.LevelError,
			Extra: map[string]interface{}{
				"error": err.Error(),
			},
			Tags: map[string]string{
				"handler": "GetAllFeedbacksHandler",
				"action":  "fetch_feedbacks",
			},
		})
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch feedbacks",
		})
	}

	var result []feedbackModel.Feedback
	for _, fb := range feedbacks {
		cached, err := Feedback.GetFeedbackFromCache(fb.Id)
		if err == nil {
			result = append(result, cached)
		} else {
			// Non trouvé en cache, on ajoute depuis la base et on le met en cache
			result = append(result, fb)
			_ = Feedback.SetFeedbackToCache(fb)
		}
	}

	return c.Status(fiber.StatusOK).JSON(result)
}
