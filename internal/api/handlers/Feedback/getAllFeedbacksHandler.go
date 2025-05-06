package Feedback

import (
	"github.com/gofiber/fiber/v2"
	"github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/database/Feedback"
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
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch feedbacks",
		})
	}

	return c.Status(fiber.StatusOK).JSON(feedbacks)
}
