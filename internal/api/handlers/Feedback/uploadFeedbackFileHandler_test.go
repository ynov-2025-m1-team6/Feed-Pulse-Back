package Feedback

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http/httptest"
	"net/textproto"
	"regexp"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/middleware"
	feedbackModel "github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/models/Feedback"
	"github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/utils/sentimentAnalysis"
)

// setupTestAppUpload creates a test Fiber app with the upload handler
func setupTestAppUpload() *fiber.App {
	app := fiber.New()
	// Handler will be registered in individual tests after middleware setup
	return app
}

// Helper function to create multipart form data with file
func createMultipartForm(fieldName, fileName, fileContent, contentType string) (*bytes.Buffer, string, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	if fileName != "" {
		// Create a custom part with the proper content type
		h := make(textproto.MIMEHeader)
		h.Set("Content-Disposition", fmt.Sprintf(`form-data; name="%s"; filename="%s"`, fieldName, fileName))
		if contentType != "" {
			h.Set("Content-Type", contentType)
		}

		part, err := writer.CreatePart(h)
		if err != nil {
			return nil, "", err
		}

		// Write content to the part
		_, err = io.WriteString(part, fileContent)
		if err != nil {
			return nil, "", err
		}
	}

	writer.Close()
	return body, writer.FormDataContentType(), nil
}

func TestUploadFeedbackFileHandler_Unauthorized(t *testing.T) {
	// Setup
	app := setupTestAppUpload()
	_, cleanup := setupMockDB(t)
	defer cleanup()

	// Register the handler (no middleware, so no user context)
	app.Post("/api/feedbacks/upload", UploadFeedbackFileHandler)

	// Create multipart form
	jsonContent := `[{"date": "2024-01-01T00:00:00Z", "channel": "test", "text": "test"}]`
	body, contentType, err := createMultipartForm("file", "test.json", jsonContent, "application/json")
	assert.NoError(t, err)

	// Create request without user context
	req := httptest.NewRequest("POST", "/api/feedbacks/upload", body)
	req.Header.Set("Content-Type", contentType)

	// Execute request (no user context set)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestUploadFeedbackFileHandler_UserNotFound(t *testing.T) {
	// Initialize sentiment analysis with dummy values for testing
	sentimentAnalysis.InitSentimentAnalysis("dummy-key", "dummy-password")

	// Setup
	app := setupTestAppUpload()
	mock, cleanup := setupMockDB(t)
	defer cleanup()

	userUUID := "550e8400-e29b-41d4-a716-446655440000"

	// Set up middleware to inject user context BEFORE the handler registration
	app.Use(func(c *fiber.Ctx) error {
		c.Locals(middleware.UserContextKey, userUUID)
		return c.Next()
	})
	// Re-register the handler after middleware
	app.Post("/api/feedbacks/upload", UploadFeedbackFileHandler)

	// Mock user lookup - return no rows (user not found)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, email FROM "users" WHERE uuid = $1`)).
		WithArgs(userUUID).
		WillReturnError(fmt.Errorf("record not found"))

	// Create multipart form
	jsonContent := `[{"date": "2024-01-01T00:00:00Z", "channel": "test", "text": "test"}]`
	body, contentType, err := createMultipartForm("file", "test.json", jsonContent, "application/json")
	assert.NoError(t, err)

	// Create request
	req := httptest.NewRequest("POST", "/api/feedbacks/upload", body)
	req.Header.Set("Content-Type", contentType)

	// Execute request
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
}

// Test helper functions

func TestParseFeedbackJson_ValidData(t *testing.T) {
	jsonData := `[{
		"date": "2024-01-01T00:00:00Z",
		"channel": "test",
		"text": "feedback"
	}]`

	feedbacks, err := parseFeedbackJson([]byte(jsonData))
	assert.NoError(t, err)
	assert.Len(t, feedbacks, 1)
	assert.Equal(t, "test", feedbacks[0].Channel)
	assert.Equal(t, "feedback", feedbacks[0].Text)
}

func TestParseFeedbackJson_InvalidJSON(t *testing.T) {
	jsonData := `invalid json`

	_, err := parseFeedbackJson([]byte(jsonData))
	assert.Error(t, err)
}

func TestParseFeedbackJson_EmptyArray(t *testing.T) {
	jsonData := `[]`

	_, err := parseFeedbackJson([]byte(jsonData))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no feedback data found")
}

func TestParseFeedbackJson_TooManyFeedbacks(t *testing.T) {
	// Create JSON with 11 feedbacks
	jsonData := "["
	for i := 0; i < 11; i++ {
		if i > 0 {
			jsonData += ","
		}
		jsonData += `{"date": "2024-01-01T00:00:00Z", "channel": "test", "text": "feedback"}`
	}
	jsonData += "]"

	_, err := parseFeedbackJson([]byte(jsonData))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "too many feedbacks")
}

func TestConvertJsonToFeedbacks(t *testing.T) {
	date, _ := time.Parse(time.RFC3339, "2024-01-01T00:00:00Z")
	feedbacksJson := []feedbackModel.FeedbackJson{
		{
			Date:    date,
			Channel: "test",
			Text:    "feedback",
		},
	}

	feedbacks := convertJsonToFeedbacks(feedbacksJson, 1)
	assert.Len(t, feedbacks, 1)
	assert.Equal(t, 1, feedbacks[0].BoardID)
	assert.Equal(t, "test", feedbacks[0].Channel)
	assert.Equal(t, "feedback", feedbacks[0].Text)
}

func TestValidateFeedbacks(t *testing.T) {
	feedbacks := []feedbackModel.Feedback{
		{
			Channel: "valid",
			Text:    "valid feedback",
		},
		{
			Channel: "",
			Text:    "invalid feedback",
		},
		{
			Channel: "another valid",
			Text:    "",
		},
	}

	validFeedbacks, errors := validateFeedbacks(feedbacks)
	assert.Len(t, validFeedbacks, 1)
	assert.Len(t, errors, 2)
	assert.Equal(t, "valid", validFeedbacks[0].Channel)
}
