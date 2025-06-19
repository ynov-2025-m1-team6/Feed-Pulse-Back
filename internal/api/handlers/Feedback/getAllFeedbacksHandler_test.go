package Feedback

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/models/Feedback"
)

// setupTestApp creates a test Fiber app with the handler
func setupTestApp() *fiber.App {
	app := fiber.New()
	app.Get("/api/feedbacks", GetAllFeedbacksHandler)
	return app
}

func TestGetAllFeedbacksHandler_Success(t *testing.T) {
	// Setup
	app := setupTestApp()
	mock, cleanup := setupMockDB(t)
	defer cleanup()

	// Create mock data
	now := time.Now()
	feedbackDate := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)

	// Mock database query with expected data
	rows := sqlmock.NewRows([]string{
		"id", "created_at", "updated_at", "date", "channel", "text", "board_id",
	}).
		AddRow(1, now, now, feedbackDate, "email", "Great product!", 1).
		AddRow(2, now, now, feedbackDate.Add(24*time.Hour), "web", "Could be improved", 1)

	mock.ExpectQuery(`SELECT \* FROM "feedbacks"`).WillReturnRows(rows)

	// Create test request
	req := httptest.NewRequest(http.MethodGet, "/api/feedbacks", nil)
	req.Header.Set("Content-Type", "application/json")

	// Execute request
	resp, err := app.Test(req)
	assert.NoError(t, err)

	// Verify response status
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	// Verify response headers
	assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))

	// Read and verify response body
	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)

	var actualFeedbacks []Feedback.Feedback
	err = json.Unmarshal(body, &actualFeedbacks)
	assert.NoError(t, err)

	// Verify the response contains expected data
	assert.Len(t, actualFeedbacks, 2)

	// Check first feedback
	assert.Equal(t, 1, actualFeedbacks[0].BaseModel.Id)
	assert.Equal(t, "email", actualFeedbacks[0].Channel)
	assert.Equal(t, "Great product!", actualFeedbacks[0].Text)
	assert.Equal(t, 1, actualFeedbacks[0].BoardID)

	// Check second feedback
	assert.Equal(t, 2, actualFeedbacks[1].BaseModel.Id)
	assert.Equal(t, "web", actualFeedbacks[1].Channel)
	assert.Equal(t, "Could be improved", actualFeedbacks[1].Text)
	assert.Equal(t, 1, actualFeedbacks[1].BoardID)

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetAllFeedbacksHandler_DatabaseError(t *testing.T) {
	// Setup
	app := setupTestApp()
	mock, cleanup := setupMockDB(t)
	defer cleanup()

	// Mock database error
	expectedError := errors.New("database connection failed")
	mock.ExpectQuery(`SELECT \* FROM "feedbacks"`).WillReturnError(expectedError)

	// Create test request
	req := httptest.NewRequest(http.MethodGet, "/api/feedbacks", nil)
	req.Header.Set("Content-Type", "application/json")

	// Execute request
	resp, err := app.Test(req)
	assert.NoError(t, err)

	// Verify response status
	assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)

	// Verify response headers
	assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))

	// Read and verify response body
	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)

	var errorResponse map[string]interface{}
	err = json.Unmarshal(body, &errorResponse)
	assert.NoError(t, err)

	// Verify error message
	assert.Equal(t, "Failed to fetch feedbacks", errorResponse["error"])

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}
