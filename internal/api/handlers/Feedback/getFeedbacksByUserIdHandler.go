package Feedback

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/database/Feedback"
	"github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/middleware"
)

// ErrorResponse represents the error response structure
type ErrorResponse struct {
	Error string `json:"error"`
}

// GetFeedbacksByUserIdHandler godoc
// @Summary Get feedbacks for the authenticated user
// @Description Retrieves feedbacks with their analyses for the authenticated user
// @Tags Feedback
// @Accept json
// @Produce json
// @Param channel query string false "Filter feedbacks by channel"
// @Success 200 {array} Feedback.FeedbackWithAnalysis "List of feedbacks with analyses"
// @Failure 400 {object} ErrorResponse "Bad request error"
// @Failure 401 {object} ErrorResponse "Unauthorized error"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /api/feedbacks/analyses [get]
// GetFeedbacksByUserIdHandler retrieves feedbacks with their analyses for a given user ID.
func GetFeedbacksByUserIdHandler(c *fiber.Ctx) error {
	// Get user UUID from context
	userUUID, check := middleware.GetUserUUID(c)
	if !check {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized: user not found in context",
		})
	}

	// Get channel if specified
	channel := c.Query("channel", "")

	feedbacks, err := Feedback.GetFeedbacksWithAnalysesByUserId(userUUID, channel)
	if err != nil {
		if errors.Is(err, Feedback.ErrUserNotFound) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "User not found",
			})
		}
		if errors.Is(err, Feedback.ErrBoardNotFound) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "No boards found for this user",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(feedbacks)
}
