package Feedback

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"time"

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

// FetchFeedbackHandler connects to JSONPlaceholder API to fetch comment data
// and converts them to feedback data with random dates and social media channels
func FetchFeedbackHandler(c *fiber.Ctx) error {
	// Default URL is JSONPlaceholder comments API
	url := "https://jsonplaceholder.typicode.com/comments"

	// Initialize random number generator properly
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	// Social media channels to choose from
	channels := []string{"twitter", "facebook", "instagram", "web"}

	// Get board ID from query parameter - default to 1 if not provided
	boardID, err := c.ParamsInt("board_id", 1)
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

	// Fetch comments from external API
	comments, err := fetchCommentsFromAPI(url)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Convert comments to feedback models
	feedbacks := convertCommentsToFeedback(comments, boardID, channels, rng)

	// Validate feedback data
	validFeedbacks, preValidationErrors := validateFeedbacks(feedbacks)
	if len(validFeedbacks) == 0 {
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
	successCount, dbErrors, err := feedbackDB.FetchAndSaveFeedbacks(validFeedbacks)
	if err != nil {
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
