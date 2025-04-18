package Feedback

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
	feedbackDB "github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/database/Feedback"
	feedbackModel "github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/models/Feedback"
)

// FetchFeedbackHandler connects to an external API to fetch feedback data
// and stores it in the database
func FetchFeedbackHandler(c *fiber.Ctx) error {
	// Extract endpoint URL from request
	url := c.Query("url")
	if url == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "API endpoint URL is required",
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
			"error": "Failed to connect to external API: " + err.Error(),
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
	var feedbacksJson []feedbackModel.FeedbackJson
	err = json.Unmarshal(bodyBytes, &feedbacksJson)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid JSON format in API response: " + err.Error(),
		})
	}

	// Validate the feedbacks
	if len(feedbacksJson) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "No feedback data found in API response",
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
		"message":       "API data processed successfully",
		"total":         len(feedbacks),
		"success_count": successCount,
		"error_count":   len(feedbacks) - successCount,
		"errors":        allErrors,
	})
}
