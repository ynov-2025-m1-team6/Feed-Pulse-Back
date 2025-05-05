package Feedback

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/gofiber/fiber/v2"
	feedbackDB "github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/database/Feedback"
	feedbackModel "github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/models/Feedback"
)

// UploadFeedbackFileHandler processes a JSON file upload and stores the data in the database
func UploadFeedbackFileHandler(c *fiber.Ctx) error {
	// Get board ID from params - default to 1 if not provided
	boardID, err := c.ParamsInt("board_id", 1)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid board ID",
		})
	}

	// Check if the board exists
	if err := validateBoardExists(boardID); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Process the uploaded file
	fileBytes, err := getUploadedFileBytes(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Parse the JSON data
	feedbacksJson, err := parseFeedbackJson(fileBytes)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid JSON format: " + err.Error(),
		})
	}

	// Convert to model feedbacks
	feedbacks := convertJsonToFeedbacks(feedbacksJson, boardID)

	// Validate the feedbacks
	validFeedbacks, preValidationErrors := validateFeedbacks(feedbacks)
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

// getUploadedFileBytes retrieves and validates the uploaded file, returning its contents
func getUploadedFileBytes(c *fiber.Ctx) ([]byte, error) {
	// Get uploaded file
	file, err := c.FormFile("file")
	if err != nil {
		return nil, errors.New("no file uploaded or invalid form field")
	}

	// Check if it's a JSON file
	if file.Header.Get("Content-Type") != "application/json" {
		return nil, errors.New("file must be JSON format")
	}

	// Open and read the file
	uploadedFile, err := file.Open()
	if err != nil {
		return nil, errors.New("failed to open uploaded file")
	}
	defer uploadedFile.Close()

	return io.ReadAll(uploadedFile)
}

// parseFeedbackJson parses the JSON data into the feedback structure
func parseFeedbackJson(fileBytes []byte) ([]feedbackModel.FeedbackJson, error) {
	var feedbacksJson []feedbackModel.FeedbackJson

	err := json.Unmarshal(fileBytes, &feedbacksJson)
	if err != nil {
		return nil, err
	}

	// Validate the feedbacks
	if len(feedbacksJson) == 0 {
		return nil, errors.New("no feedback data found in file")
	}

	return feedbacksJson, nil
}

// convertJsonToFeedbacks converts JSON feedback data to model instances
func convertJsonToFeedbacks(feedbacksJson []feedbackModel.FeedbackJson, boardID int) []feedbackModel.Feedback {
	feedbacks := make([]feedbackModel.Feedback, len(feedbacksJson))

	for i, feedback := range feedbacksJson {
		feedbacks[i] = feedbackModel.Feedback{
			Date:    feedback.Date,
			Channel: feedback.Channel,
			Text:    feedback.Text,
			BoardID: boardID,
		}
	}

	return feedbacks
}

// validateFeedbacks validates feedback data and returns valid feedbacks and errors
func validateFeedbacks(feedbacks []feedbackModel.Feedback) ([]feedbackModel.Feedback, []string) {
	validFeedbacks := make([]feedbackModel.Feedback, 0)
	validationErrors := make([]string, 0)

	for i, feedbackData := range feedbacks {
		// Validate required fields
		if feedbackData.Channel == "" || feedbackData.Text == "" {
			errorMsg := fmt.Sprintf("Feedback #%d missing required fields", i+1)
			validationErrors = append(validationErrors, errorMsg)
			continue
		}

		validFeedbacks = append(validFeedbacks, feedbackData)
	}

	return validFeedbacks, validationErrors
}
