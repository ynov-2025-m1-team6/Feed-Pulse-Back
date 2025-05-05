package Feedback

import (
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/database"
)

func setupTestGetFeedback() (sqlmock.Sqlmock, error) {
	var db *sql.DB
	var mock sqlmock.Sqlmock
	var err error

	// Create a new mock database connection
	db, mock, err = sqlmock.New()
	if err != nil {
		return nil, err
	}

	// Configure GORM to use our mock database
	dialector := postgres.New(postgres.Config{
		DSN:                  "sqlmock_db_0",
		DriverName:           "postgres",
		Conn:                 db,
		PreferSimpleProtocol: true,
	})

	gormDB, err := gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// Replace the global database connection with our mock
	database.DB = gormDB

	return mock, nil
}

func TestGetFeedbacksWithAnalysesByUserId(t *testing.T) {
	mock, err := setupTestGetFeedback()
	if err != nil {
		t.Fatalf("Error setting up test: %v", err)
	}

	// Test case 1: User not found
	userUUID := "d4c8169e-8016-496d-af6f-af4a85323028"
	mock.ExpectQuery(`SELECT count\(\*\) FROM "users" WHERE uuid = \$1`).
		WithArgs(userUUID).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	feedbacks, err := GetFeedbacksWithAnalysesByUserId(userUUID, "")
	assert.Nil(t, feedbacks)
	assert.Equal(t, ErrUserNotFound, err)

	// Test case 3: User has boards with feedbacks and analyses
	userUUID = "d4c8169e-8016-496d-af6f-af4a85323028"
	mock.ExpectQuery(`SELECT count\(\*\) FROM "users" WHERE uuid = \$1`).
		WithArgs(userUUID).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	// Mock query to get user ID from UUID
	mock.ExpectQuery(`SELECT "id" FROM "users" WHERE uuid = \$1`).
		WithArgs(userUUID).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	userID := 1
	mock.ExpectQuery(`SELECT board_id FROM "user_boards" WHERE user_id = \$1`).
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"board_id"}).
			AddRow(1).
			AddRow(2))

	// Prepare example feedback data for first board
	testDate := time.Now()
	mockFeedbacks1 := sqlmock.NewRows([]string{
		"feedback_id", "date", "channel", "text", "board_id",
		"sentiment_score", "topic"}).
		AddRow(1, testDate, "email", "Great service", 1, 0.8, "Service").
		AddRow(2, testDate, "web", "Could be better", 1, 0.3, "Quality")

	mock.ExpectQuery(`SELECT feedbacks\.\*, analyses\.\* FROM "feedbacks" LEFT JOIN analyses ON feedbacks.id = analyses.feedback_id WHERE feedbacks.board_id = \$1`).
		WithArgs(1).
		WillReturnRows(mockFeedbacks1)

	// Prepare example feedback data for second board
	mockFeedbacks2 := sqlmock.NewRows([]string{
		"feedback_id", "date", "channel", "text", "board_id",
		"sentiment_score", "topic"}).
		AddRow(3, testDate, "email", "Excellent", 2, 0.9, "Service")

	mock.ExpectQuery(`SELECT feedbacks\.\*, analyses\.\* FROM "feedbacks" LEFT JOIN analyses ON feedbacks.id = analyses.feedback_id WHERE feedbacks.board_id = \$1`).
		WithArgs(2).
		WillReturnRows(mockFeedbacks2)

	feedbacks, err = GetFeedbacksWithAnalysesByUserId(userUUID, "")
	assert.Nil(t, err)
	assert.Len(t, feedbacks, 3)
	if len(feedbacks) >= 3 {
		assert.Equal(t, 1, feedbacks[0].FeedbackID)
		assert.Equal(t, "email", feedbacks[0].Channel)
		assert.Equal(t, "Great service", feedbacks[0].Text)
		assert.Equal(t, 0.8, feedbacks[0].SentimentScore)
		assert.Equal(t, "Service", feedbacks[0].Topic)

		assert.Equal(t, 3, feedbacks[2].FeedbackID)
		assert.Equal(t, "email", feedbacks[2].Channel)
		assert.Equal(t, "Excellent", feedbacks[2].Text)
		assert.Equal(t, 2, feedbacks[2].BoardID)
	}

	// Test case 4: Database error when checking if user exists
	userUUID = "d4c8169e-8016-496d-af6f-af4a85323028"
	mock.ExpectQuery(`SELECT count\(\*\) FROM "users" WHERE uuid = \$1`).
		WithArgs(userUUID).
		WillReturnError(errors.New("database error"))

	feedbacks, err = GetFeedbacksWithAnalysesByUserId(userUUID, "")
	assert.Nil(t, feedbacks)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "database error")

	// Test case 5: Database error when retrieving boards
	userUUID = "d4c8169e-8016-496d-af6f-af4a85323028"
	mock.ExpectQuery(`SELECT count\(\*\) FROM "users" WHERE uuid = \$1`).
		WithArgs(userUUID).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	// Mock query to get user ID from UUID
	mock.ExpectQuery(`SELECT "id" FROM "users" WHERE uuid = \$1`).
		WithArgs(userUUID).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	mock.ExpectQuery(`SELECT board_id FROM "user_boards" WHERE user_id = \$1`).
		WithArgs(1).
		WillReturnError(errors.New("database error"))

	feedbacks, err = GetFeedbacksWithAnalysesByUserId(userUUID, "")
	assert.Nil(t, feedbacks)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "database error")

	// Test case 6: Database error when retrieving feedbacks
	userUUID = "d4c8169e-8016-496d-af6f-af4a85323028"
	mock.ExpectQuery(`SELECT count\(\*\) FROM "users" WHERE uuid = \$1`).
		WithArgs(userUUID).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	// Mock query to get user ID from UUID
	mock.ExpectQuery(`SELECT "id" FROM "users" WHERE uuid = \$1`).
		WithArgs(userUUID).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	mock.ExpectQuery(`SELECT board_id FROM "user_boards" WHERE user_id = \$1`).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"board_id"}).AddRow(3))

	mock.ExpectQuery(`SELECT feedbacks\.\*, analyses\.\* FROM "feedbacks" LEFT JOIN analyses ON feedbacks.id = analyses.feedback_id WHERE feedbacks.board_id = \$1`).
		WithArgs(3).
		WillReturnError(errors.New("database error"))

	feedbacks, err = GetFeedbacksWithAnalysesByUserId(userUUID, "")
	assert.Nil(t, feedbacks)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "database error")
}
