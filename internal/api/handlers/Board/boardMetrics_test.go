package Board

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/database"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func setupTest(t *testing.T) sqlmock.Sqlmock {
	// Create a mock database connection
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}

	pgDB := postgres.New(postgres.Config{
		Conn:                 db,
		PreferSimpleProtocol: true,
	})

	gormDB, err := gorm.Open(pgDB, &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open GORM DB: %v", err)
	}

	// Replace the global DB with our mock
	database.DB = gormDB

	return mock
}

func TestBoardMetricsHandler(t *testing.T) {
	t.Run("Success case", func(t *testing.T) {
		mock := setupTest(t)
		testDate := time.Now()

		// Setup expectations for GetBoardsByUserUUID
		boardRows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "name"}).
			AddRow(1, testDate, testDate, "Test Board")
		mock.ExpectQuery(`SELECT (.+) FROM "boards" JOIN user_boards ON boards.id = user_boards.board_id JOIN users ON user_boards.user_id = users.id WHERE users.uuid = \$1`).
			WithArgs("test-user-uuid").
			WillReturnRows(boardRows)

		// Setup expectations for validateBoardExists
		mock.ExpectQuery(`SELECT \* FROM "boards" WHERE id = \$1 ORDER BY "boards"."id" LIMIT \$2`).
			WithArgs(1, 1).
			WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "name"}).
				AddRow(1, testDate, testDate, "Test Board"))
		// Setup expectations for GetBoardsWithFeedbacks - GORM uses Preload which makes separate queries
		// First query: Get the board
		boardQueryRows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "name"}).
			AddRow(1, testDate, testDate, "Test Board")
		mock.ExpectQuery(`SELECT \* FROM "boards" WHERE id = \$1 ORDER BY "boards"."id" LIMIT \$2`).
			WithArgs(1, 1).
			WillReturnRows(boardQueryRows) // Second query: Preload feedbacks
		feedbackQueryRows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "date", "channel", "text", "board_id", "sentiment", "score"}).
			AddRow(1, testDate, testDate, testDate, "test", "Great feedback", 1, "positive", 0.95)
		mock.ExpectQuery(`SELECT \* FROM "feedbacks" WHERE "feedbacks"."board_id" = \$1`).
			WithArgs(1).
			WillReturnRows(feedbackQueryRows) // Third query: GetAnalysisByFeedbackID for metric calculation (called multiple times)
		analysisQueryRows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "feedback_id", "topic"}).
			AddRow(1, testDate, testDate, 1, "user experience")
		// The analysis query is called multiple times during metric calculation
		mock.ExpectQuery(`SELECT \* FROM "analyses" WHERE feedback_id = \$1 ORDER BY "analyses"."id" LIMIT \$2`).
			WithArgs(1, 1).
			WillReturnRows(analysisQueryRows)
		mock.ExpectQuery(`SELECT \* FROM "analyses" WHERE feedback_id = \$1 ORDER BY "analyses"."id" LIMIT \$2`).
			WithArgs(1, 1).
			WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "feedback_id", "topic"}).
				AddRow(1, testDate, testDate, 1, "user experience"))
		mock.ExpectQuery(`SELECT \* FROM "analyses" WHERE feedback_id = \$1 ORDER BY "analyses"."id" LIMIT \$2`).
			WithArgs(1, 1).
			WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "feedback_id", "topic"}).
				AddRow(1, testDate, testDate, 1, "user experience"))
		// Create test app and handler
		app := fiber.New()
		app.Get("/api/board/metrics", func(c *fiber.Ctx) error {
			c.Locals("userUUID", "test-user-uuid")
			return BoardMetricsHandler(c)
		})

		// Create test request
		req := httptest.NewRequest(http.MethodGet, "/api/board/metrics", nil)
		req.Header.Set("Content-Type", "application/json")

		// Execute request
		resp, err := app.Test(req)
		assert.NoError(t, err)

		// Print response body if we get a 500 error for debugging
		if resp.StatusCode == fiber.StatusInternalServerError {
			body, _ := io.ReadAll(resp.Body)
			t.Logf("Response body: %s", string(body))
		}

		// Assert response status
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)

		// Read and verify response body
		body, err := io.ReadAll(resp.Body)
		assert.NoError(t, err)
		assert.NotEmpty(t, body)

		// Verify that all mock expectations were met
		err = mock.ExpectationsWereMet()
		assert.NoError(t, err)
	})
	t.Run("Unauthorized access", func(t *testing.T) {
		app := fiber.New()
		app.Get("/api/board/metrics", BoardMetricsHandler)

		// Create test request without setting userUUID in context
		req := httptest.NewRequest(http.MethodGet, "/api/board/metrics", nil)
		req.Header.Set("Content-Type", "application/json")

		// Execute request
		resp, err := app.Test(req)
		assert.NoError(t, err)

		// Assert response
		assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)

		// Read and verify response body
		body, err := io.ReadAll(resp.Body)
		assert.NoError(t, err)
		assert.Contains(t, string(body), "unauthorized: user not found in context")
	})
	t.Run("GetBoardsByUserUUID returns error", func(t *testing.T) {
		mock := setupTest(t)

		// Setup expectation for GetBoardsByUserUUID to return an error
		mock.ExpectQuery(`SELECT (.+) FROM "boards" JOIN user_boards ON boards.id = user_boards.board_id JOIN users ON user_boards.user_id = users.id WHERE users.uuid = \$1`).
			WithArgs("test-user-uuid").
			WillReturnError(errors.New("database connection failed"))

		// Create test app and handler
		app := fiber.New()
		app.Get("/api/board/metrics", func(c *fiber.Ctx) error {
			c.Locals("userUUID", "test-user-uuid")
			return BoardMetricsHandler(c)
		})

		// Create test request
		req := httptest.NewRequest(http.MethodGet, "/api/board/metrics", nil)
		req.Header.Set("Content-Type", "application/json")

		// Execute request
		resp, err := app.Test(req)
		assert.NoError(t, err)

		// Assert response
		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

		// Read and verify response body
		body, err := io.ReadAll(resp.Body)
		assert.NoError(t, err)
		assert.Contains(t, string(body), "Invalid board")

		// Verify that all mock expectations were met
		err = mock.ExpectationsWereMet()
		assert.NoError(t, err)
	})

	t.Run("No boards found for user", func(t *testing.T) {
		mock := setupTest(t)

		// Setup expectation for GetBoardsByUserUUID to return empty result
		emptyRows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "name"})
		mock.ExpectQuery(`SELECT (.+) FROM "boards" JOIN user_boards ON boards.id = user_boards.board_id JOIN users ON user_boards.user_id = users.id WHERE users.uuid = \$1`).
			WithArgs("test-user-uuid").
			WillReturnRows(emptyRows)

		// Create test app and handler
		app := fiber.New()
		app.Get("/api/board/metrics", func(c *fiber.Ctx) error {
			c.Locals("userUUID", "test-user-uuid")
			return BoardMetricsHandler(c)
		})

		// Create test request
		req := httptest.NewRequest(http.MethodGet, "/api/board/metrics", nil)
		req.Header.Set("Content-Type", "application/json")

		// Execute request
		resp, err := app.Test(req)
		assert.NoError(t, err)

		// Assert response
		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

		// Read and verify response body
		body, err := io.ReadAll(resp.Body)
		assert.NoError(t, err)
		assert.Contains(t, string(body), "No boards found for this user")

		// Verify that all mock expectations were met
		err = mock.ExpectationsWereMet()
		assert.NoError(t, err)
	})

	t.Run("validateBoardExists fails", func(t *testing.T) {
		mock := setupTest(t)
		testDate := time.Now()

		// Setup expectations for GetBoardsByUserUUID
		boardRows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "name"}).
			AddRow(1, testDate, testDate, "Test Board")
		mock.ExpectQuery(`SELECT (.+) FROM "boards" JOIN user_boards ON boards.id = user_boards.board_id JOIN users ON user_boards.user_id = users.id WHERE users.uuid = \$1`).
			WithArgs("test-user-uuid").
			WillReturnRows(boardRows)

		// Setup expectations for validateBoardExists to fail
		mock.ExpectQuery(`SELECT \* FROM "boards" WHERE id = \$1 ORDER BY "boards"."id" LIMIT \$2`).
			WithArgs(1, 1).
			WillReturnError(gorm.ErrRecordNotFound)

		// Create test app and handler
		app := fiber.New()
		app.Get("/api/board/metrics", func(c *fiber.Ctx) error {
			c.Locals("userUUID", "test-user-uuid")
			return BoardMetricsHandler(c)
		})

		// Create test request
		req := httptest.NewRequest(http.MethodGet, "/api/board/metrics", nil)
		req.Header.Set("Content-Type", "application/json")

		// Execute request
		resp, err := app.Test(req)
		assert.NoError(t, err)

		// Assert response
		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

		// Read and verify response body
		body, err := io.ReadAll(resp.Body)
		assert.NoError(t, err)
		assert.Contains(t, string(body), "board with ID 1 does not exist")

		// Verify that all mock expectations were met
		err = mock.ExpectationsWereMet()
		assert.NoError(t, err)
	})

	t.Run("GetBoardsWithFeedbacks fails", func(t *testing.T) {
		mock := setupTest(t)
		testDate := time.Now()

		// Setup expectations for GetBoardsByUserUUID
		boardRows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "name"}).
			AddRow(1, testDate, testDate, "Test Board")
		mock.ExpectQuery(`SELECT (.+) FROM "boards" JOIN user_boards ON boards.id = user_boards.board_id JOIN users ON user_boards.user_id = users.id WHERE users.uuid = \$1`).
			WithArgs("test-user-uuid").
			WillReturnRows(boardRows)

		// Setup expectations for validateBoardExists
		mock.ExpectQuery(`SELECT \* FROM "boards" WHERE id = \$1 ORDER BY "boards"."id" LIMIT \$2`).
			WithArgs(1, 1).
			WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "name"}).
				AddRow(1, testDate, testDate, "Test Board"))

		// Setup expectations for GetBoardsWithFeedbacks to fail
		mock.ExpectQuery(`SELECT \* FROM "boards" WHERE id = \$1 ORDER BY "boards"."id" LIMIT \$2`).
			WithArgs(1, 1).
			WillReturnError(errors.New("database connection failed"))

		// Create test app and handler
		app := fiber.New()
		app.Get("/api/board/metrics", func(c *fiber.Ctx) error {
			c.Locals("userUUID", "test-user-uuid")
			return BoardMetricsHandler(c)
		})

		// Create test request
		req := httptest.NewRequest(http.MethodGet, "/api/board/metrics", nil)
		req.Header.Set("Content-Type", "application/json")

		// Execute request
		resp, err := app.Test(req)
		assert.NoError(t, err)

		// Assert response
		assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)

		// Read and verify response body
		body, err := io.ReadAll(resp.Body)
		assert.NoError(t, err)
		assert.Contains(t, string(body), "Failed to retrieve board feedbacks")

		// Verify that all mock expectations were met
		err = mock.ExpectationsWereMet()
		assert.NoError(t, err)
	})

	t.Run("CalculMetric fails", func(t *testing.T) {
		mock := setupTest(t)
		testDate := time.Now()

		// Setup expectations for GetBoardsByUserUUID
		boardRows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "name"}).
			AddRow(1, testDate, testDate, "Test Board")
		mock.ExpectQuery(`SELECT (.+) FROM "boards" JOIN user_boards ON boards.id = user_boards.board_id JOIN users ON user_boards.user_id = users.id WHERE users.uuid = \$1`).
			WithArgs("test-user-uuid").
			WillReturnRows(boardRows)

		// Setup expectations for validateBoardExists
		mock.ExpectQuery(`SELECT \* FROM "boards" WHERE id = \$1 ORDER BY "boards"."id" LIMIT \$2`).
			WithArgs(1, 1).
			WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "name"}).
				AddRow(1, testDate, testDate, "Test Board"))

		// Setup expectations for GetBoardsWithFeedbacks
		boardQueryRows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "name"}).
			AddRow(1, testDate, testDate, "Test Board")
		mock.ExpectQuery(`SELECT \* FROM "boards" WHERE id = \$1 ORDER BY "boards"."id" LIMIT \$2`).
			WithArgs(1, 1).
			WillReturnRows(boardQueryRows)

		feedbackQueryRows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "date", "channel", "text", "board_id", "sentiment", "score"}).
			AddRow(1, testDate, testDate, testDate, "test", "Great feedback", 1, "positive", 0.95)
		mock.ExpectQuery(`SELECT \* FROM "feedbacks" WHERE "feedbacks"."board_id" = \$1`).
			WithArgs(1).
			WillReturnRows(feedbackQueryRows)

		// Setup expectations for GetAnalysisByFeedbackID to fail (causing CalculMetric to fail)
		mock.ExpectQuery(`SELECT \* FROM "analyses" WHERE feedback_id = \$1 ORDER BY "analyses"."id" LIMIT \$2`).
			WithArgs(1, 1).
			WillReturnError(errors.New("analysis query failed"))

		// Create test app and handler
		app := fiber.New()
		app.Get("/api/board/metrics", func(c *fiber.Ctx) error {
			c.Locals("userUUID", "test-user-uuid")
			return BoardMetricsHandler(c)
		})

		// Create test request
		req := httptest.NewRequest(http.MethodGet, "/api/board/metrics", nil)
		req.Header.Set("Content-Type", "application/json")

		// Execute request
		resp, err := app.Test(req)
		assert.NoError(t, err)

		// Assert response
		assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)

		// Read and verify response body
		body, err := io.ReadAll(resp.Body)
		assert.NoError(t, err)
		assert.Contains(t, string(body), "Failed to calculate metrics")

		// Verify that all mock expectations were met
		err = mock.ExpectationsWereMet()
		assert.NoError(t, err)
	})
}

func TestValidateBoardExists(t *testing.T) {
	t.Run("Board exists", func(t *testing.T) {
		mock := setupTest(t)
		testDate := time.Now()

		// Setup expectations with the correct GORM query pattern
		mock.ExpectQuery(`SELECT \* FROM "boards" WHERE id = \$1 ORDER BY "boards"."id" LIMIT \$2`).
			WithArgs(1, 1). // GORM adds LIMIT 1 for First() queries
			WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "name"}).
				AddRow(1, testDate, testDate, "Test Board"))

		err := validateBoardExists(1)
		assert.NoError(t, err)
	})

	t.Run("Board does not exist", func(t *testing.T) {
		mock := setupTest(t)

		// Setup expectations with the correct GORM query pattern
		mock.ExpectQuery(`SELECT \* FROM "boards" WHERE id = \$1 ORDER BY "boards"."id" LIMIT \$2`).
			WithArgs(999, 1). // GORM adds LIMIT 1 for First() queries
			WillReturnError(gorm.ErrRecordNotFound)

		err := validateBoardExists(999)
		assert.Error(t, err)
	})
}
