package Board

import (
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
		mock.ExpectQuery(`SELECT (.+) FROM "boards"`).
			WillReturnRows(boardRows)

		// Setup expectations for validateBoardExists
		mock.ExpectQuery(`SELECT (.+) FROM "boards"`).
			WithArgs(1).
			WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "name"}).
				AddRow(1, testDate, testDate, "Test Board"))

		// Setup expectations for GetBoardsWithFeedbacks
		boardWithFeedbackRows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "name"}).
			AddRow(1, testDate, testDate, "Test Board")
		mock.ExpectQuery(`SELECT (.+) FROM "boards"`).
			WithArgs(1).
			WillReturnRows(boardWithFeedbackRows)

		feedbackRows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "date", "channel", "text", "board_id"}).
			AddRow(1, testDate, testDate, testDate, "test", "Great feedback", 1)
		mock.ExpectQuery(`SELECT (.+) FROM "feedbacks"`).
			WillReturnRows(feedbackRows)

		// Create test request
		app := fiber.New()
		req := app.AcquireCtx(&fasthttp.RequestCtx{})
		req.Locals("user_uuid", "test-user-uuid")

		// Execute handler
		err := BoardMetricsHandler(req)

		// Assert response
		assert.NoError(t, err)
		app.ReleaseCtx(req)
	})

	t.Run("Unauthorized access", func(t *testing.T) {
		app := fiber.New()
		req := app.AcquireCtx(&fasthttp.RequestCtx{})
		// Explicitly not setting user_uuid in locals to simulate unauthorized access

		res := app.AcquireCtx(&fasthttp.RequestCtx{})
		err := BoardMetricsHandler(res)

		// Err should be nil if the handler correctly handles unauthorized access
		if assert.Nil(t, err) {
			assert.Equal(t, fiber.StatusUnauthorized, res.Response().StatusCode())
			assert.Contains(t, string(res.Response().Body()), "unauthorized: user not found in context")
		} else {
			t.Error("Expected no error for unauthorized access")
		}

		app.ReleaseCtx(req)
		app.ReleaseCtx(res)
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
