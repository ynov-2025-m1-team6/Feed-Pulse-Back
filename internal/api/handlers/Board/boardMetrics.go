package Board

import (
	"errors"
	"fmt"
	"github.com/getsentry/sentry-go"
	"github.com/gofiber/fiber/v2"
	"github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/database"
	"github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/database/Board"
	"github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/middleware"
	boardModel "github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/models/Board"
	calculmetric "github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/utils/calculMetric"
	"github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/utils/httpUtils"
	"gorm.io/gorm"
)

// ErrorResponse represents the error response structure
type ErrorResponse struct {
	Error string `json:"error"`
}

// BoardMetricsHandler godoc
// @Summary Get board metrics
// @Description Get metrics for a specific board based on feedback data
// @Tags Board
// @Accept json
// @Produce json
// @Success 200 {object} metric.Metric "Metrics data"
// @Failure 400 {object} ErrorResponse "Bad request error"
// @Failure 401 {object} ErrorResponse "Unauthorized error"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /api/board/metrics [get]
func BoardMetricsHandler(c *fiber.Ctx) error {
	// Get board ID from query parameter - default to 1 if not provided
	userUUID, ok := middleware.GetUserUUID(c)
	if !ok {
		sentry.CaptureEvent(&sentry.Event{
			Message: "Unauthorized access: user UUID not found in context",
			Level:   sentry.LevelError,
			Tags: map[string]string{
				"handler": "BoardMetricsHandler",
				"action":  "get_board_metrics",
			},
		})
		return httpUtils.NewError(c, fiber.StatusUnauthorized, errors.New("unauthorized: user not found in context"))
	}

	// Get user information from database
	userBoard, err := Board.GetBoardsByUserUUID(userUUID)
	if err != nil {
		sentry.CaptureEvent(&sentry.Event{
			Message: fmt.Sprintf("Failed to retrieve boards for user %s: %v", userUUID, err),
			Level:   sentry.LevelError,
			User: sentry.User{
				ID: userUUID,
			},
			Tags: map[string]string{
				"handler": "BoardMetricsHandler",
				"action":  "get_board_metrics",
			},
		})
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid board",
		})
	}
	if len(userBoard) == 0 {
		sentry.CaptureEvent(&sentry.Event{
			Message: fmt.Sprintf("No boards found for user %s", userUUID),
			Level:   sentry.LevelWarning,
			User: sentry.User{
				ID: userUUID,
			},
			Tags: map[string]string{
				"handler": "BoardMetricsHandler",
				"action":  "get_board_metrics",
			},
		})
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "No boards found for this user",
		})
	}
	boardID := userBoard[0].Id

	// Check if the board exists in the database before proceeding
	if err := validateBoardExists(boardID); err != nil {
		sentry.CaptureEvent(&sentry.Event{
			Message: fmt.Sprintf("Board validation failed for ID %d: %v", boardID, err),
			Level:   sentry.LevelError,
			Extra: map[string]interface{}{
				"boardID": boardID,
			},
			User: sentry.User{
				ID: userUUID,
			},
			Tags: map[string]string{
				"handler": "BoardMetricsHandler",
				"action":  "get_board_metrics",
			},
		})
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	board, err := Board.GetBoardsWithFeedbacks(boardID)
	if err != nil {
		sentry.CaptureEvent(&sentry.Event{
			Message: fmt.Sprintf("Failed to retrieve board feedbacks for board ID %d: %v", boardID, err),
			Level:   sentry.LevelError,
			Extra: map[string]interface{}{
				"boardID": boardID,
			},
			User: sentry.User{
				ID: userUUID,
			},
			Tags: map[string]string{
				"handler": "BoardMetricsHandler",
				"action":  "get_board_metrics",
			},
		})
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve board feedbacks: " + err.Error(),
		})
	}

	metrics, err := calculmetric.CalculMetric(board.Feedbacks)
	if err != nil {
		sentry.CaptureEvent(&sentry.Event{
			Message: fmt.Sprintf("Failed to calculate metrics for board ID %d: %v", boardID, err),
			Level:   sentry.LevelError,
			Extra: map[string]interface{}{
				"boardID": boardID,
			},
			User: sentry.User{
				ID: userUUID,
			},
			Tags: map[string]string{
				"handler": "BoardMetricsHandler",
				"action":  "get_board_metrics",
			},
		})
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to calculate metrics: " + err.Error(),
		})
	}
	return c.JSON(metrics)
}

// validateBoardExists checks if a board with the given ID exists in the database
func validateBoardExists(boardID int) error {
	var board boardModel.Board
	result := database.DB.Where("id = ?", boardID).First(&board)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return fmt.Errorf("board with ID %d does not exist", boardID)
		}
		return fmt.Errorf("database error while checking board: %v", result.Error)
	}
	return nil
}
