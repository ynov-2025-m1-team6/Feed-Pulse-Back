package Feedback

import (
	"github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/database"
	feedbackModel "github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/models/Feedback"
)

// FetchAndSaveFeedbacks saves multiple feedbacks to the database from any source
// Returns the count of successfully saved feedbacks and any errors that might have occurred
func FetchAndSaveFeedbacks(feedbacks []feedbackModel.Feedback) (int, []string, error) {
	successCount := 0
	errors := make([]string, 0)

	// Early return if no feedbacks to process
	if len(feedbacks) == 0 {
		return 0, errors, nil
	}

	// Use a transaction to ensure data integrity
	tx := database.DB.Begin()
	if tx.Error != nil {
		return 0, nil, tx.Error
	}

	// Process each feedback
	for _, feedback := range feedbacks {
		result := tx.Create(&feedback)
		if result.Error != nil {
			errors = append(errors, result.Error.Error())
			continue // Continue processing other feedbacks instead of returning immediately
		}
		successCount++
	}

	// Commit transaction if there are successful records
	if successCount > 0 {
		if err := tx.Commit().Error; err != nil {
			// Don't add commit errors to the errors slice, as it would be returned separately as the function error
			return successCount, errors, err
		}
	} else {
		tx.Rollback()
	}

	return successCount, errors, nil
}
