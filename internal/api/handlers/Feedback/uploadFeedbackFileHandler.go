package Feedback

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/getsentry/sentry-go"
	"io"

	"github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/database"
	"github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/database/Board"
	"github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/middleware"
	"github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/models/User"

	"github.com/gofiber/fiber/v2"
	feedbackDB "github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/database/Feedback"
	feedbackModel "github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/models/Feedback"
)

// UploadFeedbackFileHandler godoc
// @Summary Upload a file containing feedbacks
// @Description Process a JSON file upload containing feedback data and store it in the database
// @Tags Feedback
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "JSON file containing feedback data"
// @Success 200 {object} map[string]interface{} "JSONPlaceholder data processed successfully"
// @Failure 400 {object} ErrorResponse "Bad request error"
// @Failure 401 {object} ErrorResponse "Unauthorized error"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /api/feedbacks/upload [post]
func UploadFeedbackFileHandler(c *fiber.Ctx) error {
	// Get user UUID from context
	userUUID, check := middleware.GetUserUUID(c)
	if !check {
		sentry.CaptureEvent(&sentry.Event{
			Message: "Unauthorized: user not found in context",
			Level:   sentry.LevelError,
			User: sentry.User{
				ID: userUUID,
			},
			Tags: map[string]string{
				"handler": "UploadFeedbackFileHandler",
				"action":  "upload_feedback_file",
			},
		})
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized: user not found in context",
		})
	}

	// Get user ID from userUUID
	// Get user ID from userUUID
	var userId int
	var userEmail string
	var u User.User
	err := database.DB.Model(&User.User{}).Where("uuid = ?", userUUID).Select("id, email").Scan(&u).Error
	if err != nil {
		sentry.CaptureEvent(&sentry.Event{
			Message: fmt.Sprintf("Failed to retrieve user ID for UUID %s: %v", userUUID, err),
			Level:   sentry.LevelError,
			Extra: map[string]interface{}{
				"error": err.Error(),
			},
			User: sentry.User{
				ID: userUUID,
			},
			Tags: map[string]string{
				"handler": "UploadFeedbackFileHandler",
				"action":  "upload_feedback_file",
			},
		})
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve user ID",
		})
	}
	userId = u.Id
	userEmail = u.Email

	// Get board ID from user ID
	boards, err := Board.GetBoardsByUserID(userId)
	if err != nil {
		sentry.CaptureEvent(&sentry.Event{
			Message: fmt.Sprintf("Failed to retrieve boards for user ID %d: %v", userId, err),
			Level:   sentry.LevelError,
			Extra: map[string]interface{}{
				"error": err.Error(),
			},
			User: sentry.User{
				ID: userUUID,
			},
			Tags: map[string]string{
				"handler": "UploadFeedbackFileHandler",
				"action":  "upload_feedback_file",
			},
		})
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve boards for user",
		})
	}
	if len(boards) == 0 {
		sentry.CaptureEvent(&sentry.Event{
			Message: fmt.Sprintf("No boards found for user ID %d", userId),
			Level:   sentry.LevelWarning,
			User: sentry.User{
				ID: userUUID,
			},
			Tags: map[string]string{
				"handler": "UploadFeedbackFileHandler",
				"action":  "upload_feedback_file",
			},
		})
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "No boards found for this user",
		})
	}
	boardID := boards[0].BaseModel.Id

	// Check if the board exists
	if err := validateBoardExists(boardID); err != nil {
		sentry.CaptureEvent(&sentry.Event{
			Message: fmt.Sprintf("Board with ID %d does not exist: %v", boardID, err),
			Level:   sentry.LevelError,
			Extra: map[string]interface{}{
				"board_id": boardID,
				"error":    err.Error(),
			},
			Tags: map[string]string{
				"handler": "UploadFeedbackFileHandler",
				"action":  "upload_feedback_file",
			},
		})
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Process the uploaded file
	fileBytes, err := getUploadedFileBytes(c)
	if err != nil {
		sentry.CaptureEvent(&sentry.Event{
			Message: fmt.Sprintf("Failed to get uploaded file bytes: %v", err),
			Level:   sentry.LevelError,
			Extra: map[string]interface{}{
				"board_id": boardID,
			},
			User: sentry.User{
				ID:    userUUID,
				Email: userEmail,
			},
			Tags: map[string]string{
				"handler": "UploadFeedbackFileHandler",
				"action":  "upload_feedback_file",
			},
		})
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Parse the JSON data
	feedbacksJson, err := parseFeedbackJson(fileBytes)
	if err != nil {
		sentry.CaptureEvent(&sentry.Event{
			Message: fmt.Sprintf("Failed to parse JSON feedback data: %v", err),
			Level:   sentry.LevelError,
			Extra: map[string]interface{}{
				"board_id": boardID,
			},
			User: sentry.User{
				ID:    userUUID,
				Email: userEmail,
			},
			Attachments: []*sentry.Attachment{
				{
					Filename: "feedback_file.json",
					Payload:  fileBytes,
				},
			},
			Tags: map[string]string{
				"handler": "UploadFeedbackFileHandler",
				"action":  "upload_feedback_file",
			},
		})
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid JSON format: " + err.Error(),
		})
	}

	// Convert to model feedbacks
	feedbacks := convertJsonToFeedbacks(feedbacksJson, boardID)

	// Validate the feedbacks
	validFeedbacks, preValidationErrors := validateFeedbacks(feedbacks)
	if len(validFeedbacks) == 0 {
		sentry.CaptureEvent(&sentry.Event{
			Message: fmt.Sprintf("No valid feedback data found after validation for user ID %d", userId),
			Level:   sentry.LevelWarning,
			Extra: map[string]interface{}{
				"board_id": boardID,
			},
			User: sentry.User{
				ID:    userUUID,
				Email: userEmail,
			},
			Attachments: []*sentry.Attachment{
				{
					Filename: "feedback_file.json",
					Payload:  fileBytes,
				},
			},
			Tags: map[string]string{
				"handler": "UploadFeedbackFileHandler",
				"action":  "upload_feedback_file",
			},
		})
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":             "No valid feedback data found after validation",
			"validation_errors": preValidationErrors,
		})
	}

	// Save feedbacks to database using the dedicated database function
	successCount, dbErrors, err := feedbackDB.UploadFeedbacksFromFile(feedbacks, userEmail)
	if err != nil {
		sentry.CaptureEvent(&sentry.Event{
			Message: fmt.Sprintf("Database error while uploading feedbacks for user ID %d: %v", userId, err),
			Level:   sentry.LevelError,
			Extra: map[string]interface{}{
				"board_id": boardID,
			},
			User: sentry.User{
				ID:    userUUID,
				Email: userEmail,
			},
			Attachments: []*sentry.Attachment{
				{
					Filename: "feedback_file.json",
					Payload:  fileBytes,
				},
			},
			Tags: map[string]string{
				"handler": "UploadFeedbackFileHandler",
				"action":  "upload_feedback_file",
			},
		})
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
	} else if len(feedbacksJson) > 10 {
		return nil, errors.New("too many feedbacks in file, maximum is 10")
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
