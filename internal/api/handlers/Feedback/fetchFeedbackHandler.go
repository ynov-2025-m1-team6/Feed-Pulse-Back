package Feedback

import (
	"encoding/json"
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

// JSONPlaceholder comment structure
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

	// Initialize random seed
	rand.NewSource(time.Now().UnixNano())

	// Social media channels to choose from
	channels := []string{"twitter", "facebook", "instagram", "web"}

	// Board ID from query parameter - default to 1 if not provided
	boardID := c.QueryInt("boardId", 1)

	// Check if the board exists in the database before proceeding
	var board boardModel.Board
	result := database.DB.Where("id = ?", boardID).First(&board)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": fmt.Sprintf("Board with ID %d does not exist", boardID),
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Database error while checking board: " + result.Error.Error(),
		})
	}

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 15 * time.Second,
	}

	// Make the request to the external API
	resp, err := client.Get(url)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to connect to JSONPlaceholder API: " + err.Error(),
		})
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fmt.Sprintf("External API returned error: %d %s", resp.StatusCode, resp.Status),
		})
	}

	// Read response body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to read API response: " + err.Error(),
		})
	}

	// Parse JSON data from response
	var comments []Comment
	err = json.Unmarshal(bodyBytes, &comments)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid JSON format in API response: " + err.Error(),
		})
	}

	// Validate the comments
	if len(comments) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "No comment data found in API response",
		})
	}

	// Convert comments to feedback model
	feedbacks := make([]feedbackModel.Feedback, len(comments))
	now := time.Now()

	for i, comment := range comments {
		// Generate a random date between 1 month ago and now
		randomDuration := time.Duration(-1*rand.Intn(30*24)) * time.Hour
		randomDate := now.Add(randomDuration).UTC()

		// Choose a random channel
		randomChannel := channels[rand.Intn(len(channels))]

		feedbacks[i] = feedbackModel.Feedback{
			Date:    randomDate,
			Channel: randomChannel,
			Text:    comment.Body,
			BoardID: boardID,
		}
	}

	// Validate feedback data (should always be valid in this case, but kept for safety)
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

	// Limit number of feedbacks to process if they're too many (optional)
	limitParam := c.QueryInt("limit", 0)
	if limitParam > 0 && limitParam < len(validFeedbacks) {
		validFeedbacks = validFeedbacks[:limitParam]
	}

	// Save feedbacks to database using the dedicated database function
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
