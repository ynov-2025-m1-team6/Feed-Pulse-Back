package Feedback

import (
	"encoding/json"
	"fmt"
	"github.com/gofiber/fiber/v2"
	feedbackDB "github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/database/Feedback"
	feedbackModel "github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/models/Feedback"
	"io"
)

// UploadFeedbackFileHandler processes a JSON file upload and stores the data in the database
func UploadFeedbackFileHandler(c *fiber.Ctx) error {
	// Get uploaded file
	file, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "No file uploaded or invalid form field",
		})
	}

	// Check if it's a JSON file
	if file.Header.Get("Content-Type") != "application/json" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "File must be JSON format",
		})
	}

	// Open and read the file
	uploadedFile, err := file.Open()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to open uploaded file",
		})
	}
	defer uploadedFile.Close()

	fileBytes, err := io.ReadAll(uploadedFile)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to read uploaded file",
		})
	}

	// Parse the JSON data - try both formats
	var feedbacksJson []feedbackModel.FeedbackJson

	// First try direct array format
	err = json.Unmarshal(fileBytes, &feedbacksJson)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid JSON format: " + err.Error(),
		})
	}

	// Validate the feedbacks
	if len(feedbacksJson) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "No feedback data found in file",
		})
	}

	// Cast to Model Feedback
	feedbacks := make([]feedbackModel.Feedback, len(feedbacksJson))
	for i, feedback := range feedbacksJson {
		feedbacks[i] = feedbackModel.Feedback{
			Date:    feedback.Date,
			Channel: feedback.Channel,
			Text:    feedback.Text,
		}
	}

	// Validate and prepare feedback data
	validFeedbacks := make([]feedbackModel.Feedback, 0)
	preValidationErrors := make([]string, 0)

	for i, feedbackData := range feedbacks {
		// Validate required fields
		if feedbackData.Channel == "" || feedbackData.Text == "" {
			errorMsg := fmt.Sprintf("Feedback #%d missing required fields", i+1)
			preValidationErrors = append(preValidationErrors, errorMsg)
			continue
		}

		validFeedbacks = append(validFeedbacks, feedbackData)
	}

	// If no valid feedbacks after validation, return error
	if len(validFeedbacks) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":             "No valid feedback data found after validation",
			"validation_errors": preValidationErrors,
		})
	}

	// Save feedbacks to database using the dedicated database function
	successCount, dbErrors, err := feedbackDB.UploadFeedbacksFromFile(feedbacks)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Database error: " + err.Error(),
		})
	}

	// Combine all errors
	allErrors := append(preValidationErrors, dbErrors...)

	// Return summary of the operation
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message":       "File processed successfully",
		"total":         len(feedbacks),
		"success_count": successCount,
		"error_count":   len(feedbacks) - successCount,
		"errors":        allErrors,
	})
}
