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
	feedbackModel "github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/models/Feedback"
)

func TestGetFeedbacksByUserIdHandler_Success(t *testing.T) {
	// Setup
	mock, cleanup := setupMockDB(t)
	defer cleanup()

	userUUID := "test-user-uuid-123"
	testDate := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)

	// Mock successful database queries
	// 1. Check if user exists
	mock.ExpectQuery(`SELECT count\(\*\) FROM "users" WHERE uuid = \$1`).
		WithArgs(userUUID).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	// 2. Get user ID from UUID
	mock.ExpectQuery(`SELECT "id" FROM "users" WHERE uuid = \$1`).
		WithArgs(userUUID).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	// 3. Get board IDs for user
	mock.ExpectQuery(`SELECT board_id FROM "user_boards" WHERE user_id = \$1`).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"board_id"}).
			AddRow(1).
			AddRow(2))

	// 4. Get feedbacks with analyses for first board
	mockFeedbacks1 := sqlmock.NewRows([]string{
		"feedback_id", "date", "channel", "text", "board_id",
		"sentiment_score", "topic",
	}).
		AddRow(1, testDate, "email", "Great service!", 1, 0.8, "Service").
		AddRow(2, testDate, "web", "Could be better", 1, 0.3, "Quality")

	mock.ExpectQuery(`SELECT feedbacks\.\*, analyses\.\* FROM "feedbacks" LEFT JOIN analyses ON feedbacks.id = analyses.feedback_id WHERE feedbacks.board_id = \$1`).
		WithArgs(1).
		WillReturnRows(mockFeedbacks1)

	// 5. Get feedbacks with analyses for second board
	mockFeedbacks2 := sqlmock.NewRows([]string{
		"feedback_id", "date", "channel", "text", "board_id",
		"sentiment_score", "topic",
	}).
		AddRow(3, testDate, "mobile", "Excellent app!", 2, 0.9, "Product")

	mock.ExpectQuery(`SELECT feedbacks\.\*, analyses\.\* FROM "feedbacks" LEFT JOIN analyses ON feedbacks.id = analyses.feedback_id WHERE feedbacks.board_id = \$1`).
		WithArgs(2).
		WillReturnRows(mockFeedbacks2)
	// Create test request with user context
	req := httptest.NewRequest(http.MethodGet, "/api/feedbacks/analyses", nil)
	req.Header.Set("Content-Type", "application/json")

	// Create a new app instance with middleware simulation
	testApp := fiber.New()
	testApp.Get("/api/feedbacks/analyses", func(c *fiber.Ctx) error {
		// Simulate middleware setting user UUID
		c.Locals("userUUID", userUUID)
		return GetFeedbacksByUserIdHandler(c)
	})

	// Execute request
	resp, err := testApp.Test(req)
	assert.NoError(t, err)

	// Verify response status
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	// Verify response headers
	assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))

	// Read and verify response body
	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)

	var actualFeedbacks []feedbackModel.FeedbackWithAnalysis
	err = json.Unmarshal(body, &actualFeedbacks)
	assert.NoError(t, err)

	// Verify the response contains expected data
	assert.Len(t, actualFeedbacks, 3)

	// Check first feedback
	assert.Equal(t, 1, actualFeedbacks[0].FeedbackID)
	assert.Equal(t, "email", actualFeedbacks[0].Channel)
	assert.Equal(t, "Great service!", actualFeedbacks[0].Text)
	assert.Equal(t, 1, actualFeedbacks[0].BoardID)
	assert.Equal(t, 0.8, actualFeedbacks[0].SentimentScore)
	assert.Equal(t, "Service", actualFeedbacks[0].Topic)

	// Check second feedback
	assert.Equal(t, 2, actualFeedbacks[1].FeedbackID)
	assert.Equal(t, "web", actualFeedbacks[1].Channel)
	assert.Equal(t, "Could be better", actualFeedbacks[1].Text)
	assert.Equal(t, 1, actualFeedbacks[1].BoardID)
	assert.Equal(t, 0.3, actualFeedbacks[1].SentimentScore)
	assert.Equal(t, "Quality", actualFeedbacks[1].Topic)

	// Check third feedback
	assert.Equal(t, 3, actualFeedbacks[2].FeedbackID)
	assert.Equal(t, "mobile", actualFeedbacks[2].Channel)
	assert.Equal(t, "Excellent app!", actualFeedbacks[2].Text)
	assert.Equal(t, 2, actualFeedbacks[2].BoardID)
	assert.Equal(t, 0.9, actualFeedbacks[2].SentimentScore)
	assert.Equal(t, "Product", actualFeedbacks[2].Topic)

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetFeedbacksByUserIdHandler_Unauthorized(t *testing.T) {
	// Setup
	app := fiber.New()
	app.Get("/api/feedbacks/analyses", GetFeedbacksByUserIdHandler)

	// Create test request without user context
	req := httptest.NewRequest(http.MethodGet, "/api/feedbacks/analyses", nil)
	req.Header.Set("Content-Type", "application/json")

	// Execute request without setting user UUID in context
	resp, err := app.Test(req)
	assert.NoError(t, err)

	// Verify response status
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)

	// Verify response headers
	assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))

	// Read and verify response body
	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)

	var errorResponse map[string]interface{}
	err = json.Unmarshal(body, &errorResponse)
	assert.NoError(t, err)

	// Verify error message
	assert.Equal(t, "Unauthorized: user not found in context", errorResponse["error"])
}

func TestGetFeedbacksByUserIdHandler_UserNotFound(t *testing.T) {
	// Setup
	mock, cleanup := setupMockDB(t)
	defer cleanup()

	userUUID := "nonexistent-user-uuid"

	// Mock user not found
	mock.ExpectQuery(`SELECT count\(\*\) FROM "users" WHERE uuid = \$1`).
		WithArgs(userUUID).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	// Create test app with middleware simulation
	testApp := fiber.New()
	testApp.Get("/api/feedbacks/analyses", func(c *fiber.Ctx) error {
		// Simulate middleware setting user UUID
		c.Locals("userUUID", userUUID)
		return GetFeedbacksByUserIdHandler(c)
	})

	// Create test request
	req := httptest.NewRequest(http.MethodGet, "/api/feedbacks/analyses", nil)
	req.Header.Set("Content-Type", "application/json")

	// Execute request
	resp, err := testApp.Test(req)
	assert.NoError(t, err)

	// Verify response status
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

	// Read and verify response body
	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)

	var errorResponse map[string]interface{}
	err = json.Unmarshal(body, &errorResponse)
	assert.NoError(t, err)

	// Verify error message
	assert.Equal(t, "User not found", errorResponse["error"])

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetFeedbacksByUserIdHandler_BoardNotFound(t *testing.T) {
	// Setup
	mock, cleanup := setupMockDB(t)
	defer cleanup()

	userUUID := "user-with-no-boards"

	// Mock user exists but has no boards
	mock.ExpectQuery(`SELECT count\(\*\) FROM "users" WHERE uuid = \$1`).
		WithArgs(userUUID).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	mock.ExpectQuery(`SELECT "id" FROM "users" WHERE uuid = \$1`).
		WithArgs(userUUID).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	// Mock empty board list (no boards for user)
	mock.ExpectQuery(`SELECT board_id FROM "user_boards" WHERE user_id = \$1`).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"board_id"}))

	// Create test app with middleware simulation
	testApp := fiber.New()
	testApp.Get("/api/feedbacks/analyses", func(c *fiber.Ctx) error {
		// Simulate middleware setting user UUID
		c.Locals("userUUID", userUUID)
		return GetFeedbacksByUserIdHandler(c)
	})

	// Create test request
	req := httptest.NewRequest(http.MethodGet, "/api/feedbacks/analyses", nil)
	req.Header.Set("Content-Type", "application/json")

	// Execute request
	resp, err := testApp.Test(req)
	assert.NoError(t, err)

	// Verify response status
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

	// Read and verify response body
	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)

	var errorResponse map[string]interface{}
	err = json.Unmarshal(body, &errorResponse)
	assert.NoError(t, err)

	// Verify error message
	assert.Equal(t, "No boards found for this user", errorResponse["error"])

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetFeedbacksByUserIdHandler_DatabaseError(t *testing.T) {
	// Setup
	mock, cleanup := setupMockDB(t)
	defer cleanup()

	userUUID := "test-user-uuid"

	// Mock user exists
	mock.ExpectQuery(`SELECT count\(\*\) FROM "users" WHERE uuid = \$1`).
		WithArgs(userUUID).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	mock.ExpectQuery(`SELECT "id" FROM "users" WHERE uuid = \$1`).
		WithArgs(userUUID).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	mock.ExpectQuery(`SELECT board_id FROM "user_boards" WHERE user_id = \$1`).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"board_id"}).AddRow(1))

	// Mock database error when fetching feedbacks
	mock.ExpectQuery(`SELECT feedbacks\.\*, analyses\.\* FROM "feedbacks" LEFT JOIN analyses ON feedbacks.id = analyses.feedback_id WHERE feedbacks.board_id = \$1`).
		WithArgs(1).
		WillReturnError(errors.New("database connection failed"))

	// Create test app with middleware simulation
	testApp := fiber.New()
	testApp.Get("/api/feedbacks/analyses", func(c *fiber.Ctx) error {
		// Simulate middleware setting user UUID
		c.Locals("userUUID", userUUID)
		return GetFeedbacksByUserIdHandler(c)
	})

	// Create test request
	req := httptest.NewRequest(http.MethodGet, "/api/feedbacks/analyses", nil)
	req.Header.Set("Content-Type", "application/json")

	// Execute request
	resp, err := testApp.Test(req)
	assert.NoError(t, err)
	// Verify response status
	assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)

	// Read and verify response body
	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)

	var errorResponse map[string]interface{}
	err = json.Unmarshal(body, &errorResponse)
	assert.NoError(t, err)

	// Verify error message contains the database error
	assert.Contains(t, errorResponse["error"], "database connection failed")

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}
