package Feedback

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/getsentry/sentry-go"
	"io"
	"math/rand"
	"net/http"
	"time"

	"github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/database/Board"
	"github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/middleware"
	"github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/models/User"

	"github.com/gofiber/fiber/v2"
	"github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/database"
	feedbackDB "github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/database/Feedback"
	boardModel "github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/models/Board"
	feedbackModel "github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/models/Feedback"
	"gorm.io/gorm"
)

// Comment represents the JSONPlaceholder comment structure
type Comment struct {
	PostID int    `json:"postId"`
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Email  string `json:"email"`
	Body   string `json:"body"`
}

// FetchFeedbackHandler godoc
// @Summary Fetch feedback data from external API
// @Description Fetches comment data from JSONPlaceholder API and converts them to feedback data
// @Tags Feedback
// @Accept json
// @Produce json
// @Param limit query int false "Limit the number of feedbacks to fetch"
// @Success 200 {object} map[string]interface{} "JSONPlaceholder data processed successfully"
// @Failure 400 {object} ErrorResponse "Bad request error"
// @Failure 401 {object} ErrorResponse "Unauthorized error"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /api/feedbacks/fetch [post]
func FetchFeedbackHandler(c *fiber.Ctx) error {
	// Default URL is JSONPlaceholder comments API
	url := "https://jsonplaceholder.typicode.com/comments"

	// Initialize random number generator properly
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	// Social media channels to choose from
	channels := []string{"twitter", "facebook", "instagram", "web"}

	// Get user UUID from context
	userUUID, check := middleware.GetUserUUID(c)
	if !check {
		sentry.CaptureEvent(&sentry.Event{
			Message: "Unauthorized access: user not found in context",
			Level:   sentry.LevelError,
			Tags: map[string]string{
				"handler": "FetchFeedbackHandler",
				"action":  "fetch_feedback",
			},
		})
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized: user not found in context",
		})
	}

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
				"user_uuid": userUUID,
				"error":     err.Error(),
			},
			Tags: map[string]string{
				"handler": "FetchFeedbackHandler",
				"action":  "fetch_feedback",
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
				"user_id": userId,
				"error":   err.Error(),
			},
			Tags: map[string]string{
				"handler": "FetchFeedbackHandler",
				"action":  "fetch_feedback",
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
			Extra: map[string]interface{}{
				"user_id": userId,
			},
			Tags: map[string]string{
				"handler": "FetchFeedbackHandler",
				"action":  "fetch_feedback",
			},
		})
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "No boards found for this user",
		})
	}
	boardID := boards[0].BaseModel.Id

	// Check if the board exists in the database before proceeding
	if err := validateBoardExists(boardID); err != nil {
		sentry.CaptureEvent(&sentry.Event{
			Message: fmt.Sprintf("Board validation failed for ID %d: %v", boardID, err),
			Level:   sentry.LevelError,
			Extra: map[string]interface{}{
				"board_id": boardID,
				"error":    err.Error(),
			},
			Tags: map[string]string{
				"handler": "FetchFeedbackHandler",
				"action":  "fetch_feedback",
			},
		})
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Fetch comments from external API
	comments, err := fetchCommentsFromAPI(url)
	if err != nil {
		sentry.CaptureEvent(&sentry.Event{
			Message: fmt.Sprintf("Failed to fetch comments from API: %v", err),
			Level:   sentry.LevelError,
			Extra: map[string]interface{}{
				"url":   url,
				"error": err.Error(),
			},
			Tags: map[string]string{
				"handler": "FetchFeedbackHandler",
				"action":  "fetch_feedback",
			},
		})
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Convert comments to feedback models
	feedbacks := convertCommentsToFeedback(comments, boardID, channels, rng)

	// Validate feedback data
	validFeedbacks, preValidationErrors := validateFeedbacks(feedbacks)
	if len(validFeedbacks) == 0 {
		sentry.CaptureEvent(&sentry.Event{
			Message: "No valid feedback data found after validation",
			Level:   sentry.LevelWarning,
			Extra: map[string]interface{}{
				"feedbacks": feedbacks,
			},
			Tags: map[string]string{
				"handler": "FetchFeedbackHandler",
				"action":  "fetch_feedback",
			},
		})
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":             "No valid feedback data found after validation",
			"validation_errors": preValidationErrors,
		})
	}

	// Apply limit if requested
	limitParam := c.QueryInt("limit", 0)
	if limitParam > 0 && limitParam < len(validFeedbacks) {
		validFeedbacks = validFeedbacks[:limitParam]
	}

	// Save feedbacks to database
	successCount, dbErrors, err := feedbackDB.FetchAndSaveFeedbacks(validFeedbacks, userEmail)
	if err != nil {
		sentry.CaptureEvent(&sentry.Event{
			Message: fmt.Sprintf("Database error while saving feedbacks: %v", err),
			Level:   sentry.LevelError,
			Extra: map[string]interface{}{
				"user_email": userEmail,
				"feedbacks":  validFeedbacks,
				"error":      err.Error(),
			},
			Tags: map[string]string{
				"handler": "FetchFeedbackHandler",
				"action":  "fetch_feedback",
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
		"message":       "JSONPlaceholder data processed successfully",
		"total":         len(feedbacks),
		"success_count": successCount,
		"error_count":   len(feedbacks) - successCount,
		"errors":        allErrors,
	})
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

// fetchCommentsFromAPI fetches comments from the JSONPlaceholder API
func fetchCommentsFromAPI(url string) ([]Comment, error) {
	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 15 * time.Second,
	}

	// Make the request to the external API
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to JSONPlaceholder API: %v", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("external API returned error: %d %s", resp.StatusCode, resp.Status)
	}

	// Read response body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read API response: %v", err)
	}

	// Parse JSON data from response
	var comments []Comment
	err = json.Unmarshal(bodyBytes, &comments)
	if err != nil {
		return nil, fmt.Errorf("invalid JSON format in API response: %v", err)
	}

	// Validate the comments
	if len(comments) == 0 {
		return nil, fmt.Errorf("no comment data found in API response")
	}

	return comments, nil
}

// convertCommentsToFeedback converts comments to feedback models with random dates and channels
func convertCommentsToFeedback(comments []Comment, boardID int, channels []string, rng *rand.Rand) []feedbackModel.Feedback {
	feedbacks := make([]feedbackModel.Feedback, len(comments))
	now := time.Now()

	for i, comment := range comments {
		// Generate a random date between 1 month ago and now
		randomDuration := time.Duration(-1*rng.Intn(30*24)) * time.Hour
		randomDate := now.Add(randomDuration).UTC()

		// Choose a random channel
		randomChannel := channels[rng.Intn(len(channels))]

		feedbacks[i] = feedbackModel.Feedback{
			Date:    randomDate,
			Channel: randomChannel,
			Text:    comment.Body,
			BoardID: boardID,
		}
	}

	return feedbacks
}
