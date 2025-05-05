package Feedback

import (
	"errors"
	"fmt"

	"github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/database"
	boardModel "github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/models/Board"
	feedbackModel "github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/models/Feedback"
	"gorm.io/gorm"
)

// UploadFeedbacksFromFile saves multiple feedbacks to the database
// Returns:
//   - int: count of successfully saved feedbacks
//   - []string: list of error messages that occurred during processing
//   - error: any critical error that prevented the overall operation
func UploadFeedbacksFromFile(feedbacks []feedbackModel.Feedback) (int, []string, error) {
	successCount := 0
	errorMessages := make([]string, 0)

	// Early return if no feedbacks to process
	if len(feedbacks) == 0 {
		return 0, errorMessages, nil
	}

	// Use a transaction to ensure data integrity
	tx := database.DB.Begin()
	if tx.Error != nil {
		return 0, nil, fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}

	// Create a map to track which board IDs we've already verified
	verifiedBoards := make(map[int]bool)

	// Process each feedback
	for i, feedback := range feedbacks {
		// Check if we need to verify this board exists
		if _, alreadyVerified := verifiedBoards[feedback.BoardID]; !alreadyVerified {
			// Check if the board exists
			var board boardModel.Board
			result := tx.Where("id = ?", feedback.BoardID).First(&board)
			if result.Error != nil {
				if errors.Is(result.Error, gorm.ErrRecordNotFound) {
					errorMsg := fmt.Sprintf("Feedback #%d: board with ID %d does not exist", i+1, feedback.BoardID)
					errorMessages = append(errorMessages, errorMsg)
					continue
				}
				// Database error, rollback transaction
				tx.Rollback()
				return successCount, errorMessages, fmt.Errorf("database error while checking board: %w", result.Error)
			}
			// Mark this board as verified
			verifiedBoards[feedback.BoardID] = true
		}

		// Create the feedback
		_, err := CreateFeedback(feedback)
		if err != nil {
			errorMsg := fmt.Sprintf("Feedback #%d: %s", i+1, err.Error())
			errorMessages = append(errorMessages, errorMsg)
			continue
		}
		successCount++
	}

	// Commit transaction if there are successful records
	if successCount > 0 {
		if err := tx.Commit().Error; err != nil {
			return successCount, errorMessages, fmt.Errorf("failed to commit transaction: %w", err)
		}
	} else {
		tx.Rollback()
	}

	return successCount, errorMessages, nil
}
