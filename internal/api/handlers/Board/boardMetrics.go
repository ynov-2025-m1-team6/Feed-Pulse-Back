package Board

import (
	"errors"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/database"
	"github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/database/Board"
	boardModel "github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/models/Board"
	calculmetric "github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/utils/calculMetric"
	"gorm.io/gorm"
)

func BoardMetricsHandler(c *fiber.Ctx) error {
	// Get board ID from query parameter - default to 1 if not provided
	boardID, err := c.ParamsInt("board_id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid board ID",
		})
	}

	// Check if the board exists in the database before proceeding
	if err := validateBoardExists(boardID); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	board, err := Board.GetBoardsWithFeedbacks(boardID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve board feedbacks: " + err.Error(),
		})
	}

	metrics, err := calculmetric.CalculMetric(board.Feedbacks)
	if err != nil {
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
