package Feedback

import (
	"github.com/gofiber/fiber/v2"
	"github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/database/Feedback"
)

// GetFeedbacksByUserIdHandler retrieves feedbacks with their analyses for a given user ID.
func GetFeedbacksByUserIdHandler(c *fiber.Ctx) error {
	userId, err := c.ParamsInt("user_id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user ID",
		})
	}
	if userId <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "User ID must be a positive integer",
		})
	}

	feedbacks, err := Feedback.GetFeedbacksWithAnalysesByUserId(userId)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(feedbacks)
}
