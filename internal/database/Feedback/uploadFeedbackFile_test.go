package Feedback

import (
	"github.com/tot0p/env"
	"github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/utils/sentimentAnalysis"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/database"
	"github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/models/Feedback"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestUploadFeedbacksFromFile(t *testing.T) {

	//init the mixtral client
	_ = env.LoadPath("../../../.env")
	sentimentAnalysis.InitSentimentAnalysis(env.Get("MISTRAL_API_KEY"))

	// Test successful upload of feedbacks
	t.Run("Successfully upload and save multiple feedbacks", func(t *testing.T) {
		// Setup test DB with regexp matcher for more flexible matching
		mockDB, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
		if err != nil {
			t.Fatalf("Error creating mock database: %v", err)
		}
		defer mockDB.Close()

		dialector := postgres.New(postgres.Config{
			Conn:                 mockDB,
			PreferSimpleProtocol: true,
		})

		db, err := gorm.Open(dialector, &gorm.Config{})
		if err != nil {
			t.Fatalf("Error opening gorm DB: %v", err)
		}

		database.DB = db

		// Define test data
		testDate := time.Now()
		testFeedbacks := []Feedback.Feedback{
			{
				Date:    testDate,
				Channel: "email",
				Text:    "Test feedback 1",
				BoardID: 1,
			},
			{
				Date:    testDate,
				Channel: "chat",
				Text:    "Test feedback 2",
				BoardID: 1,
			},
		}

		// Mock expectations
		mock.ExpectBegin()

		// Expect board validation query in the transaction
		boardRows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "name"}).
			AddRow(1, time.Now(), time.Now(), "Test Board")
		mock.ExpectQuery(`SELECT \* FROM "boards"`).
			WithArgs(1, 1).
			WillReturnRows(boardRows)

		// For the first feedback, expect validation check inside CreateFeedback
		boardRowsAgain := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "name"}).
			AddRow(1, time.Now(), time.Now(), "Test Board")
		mock.ExpectQuery(`SELECT \* FROM "boards"`).
			WithArgs(1, 1).
			WillReturnRows(boardRowsAgain)

		// Then expect the insert for the first feedback
		mock.ExpectBegin()
		mock.ExpectQuery(`INSERT INTO "feedbacks"`).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
		mock.ExpectQuery(`SELECT (.+) FROM "feedbacks" WHERE (.+)`).
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "date", "channel", "text", "board_id"}).
				AddRow(1, time.Now(), time.Now(), testDate, "email", "The application is great!", 1))
		mock.ExpectQuery(`INSERT INTO "analyses"`).
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
		mock.ExpectCommit()

		// For the second feedback, expect validation check inside CreateFeedback
		boardRowsAgain2 := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "name"}).
			AddRow(1, time.Now(), time.Now(), "Test Board")
		mock.ExpectQuery(`SELECT \* FROM "boards"`).
			WithArgs(1, 1).
			WillReturnRows(boardRowsAgain2)

		// Then expect the insert for the second feedback
		mock.ExpectBegin()
		mock.ExpectQuery(`INSERT INTO "feedbacks"`).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(2))
		mock.ExpectQuery(`SELECT (.+) FROM "feedbacks" WHERE (.+)`).
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "date", "channel", "text", "board_id"}).
				AddRow(1, time.Now(), time.Now(), testDate, "email", "The application is great!", 1))
		mock.ExpectQuery(`INSERT INTO "analyses"`).
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
		mock.ExpectCommit()

		mock.ExpectCommit()

		// Execute the function being tested
		successCount, errs, err := UploadFeedbacksFromFile(testFeedbacks)

		// Assert results
		assert.Nil(t, err)
		assert.Equal(t, len(testFeedbacks), successCount)
		assert.Empty(t, errs)

		// Verify all expectations were met
		err = mock.ExpectationsWereMet()
		assert.NoError(t, err)
	})

	// Test non-existent board
	t.Run("Board does not exist", func(t *testing.T) {
		// Setup test DB with regexp matcher for more flexible matching
		mockDB, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
		if err != nil {
			t.Fatalf("Error creating mock database: %v", err)
		}
		defer mockDB.Close()

		dialector := postgres.New(postgres.Config{
			Conn:                 mockDB,
			PreferSimpleProtocol: true,
		})

		db, err := gorm.Open(dialector, &gorm.Config{})
		if err != nil {
			t.Fatalf("Error opening gorm DB: %v", err)
		}

		database.DB = db

		// Define test data with non-existent board
		testFeedbacks := []Feedback.Feedback{
			{
				Date:    time.Now(),
				Channel: "email",
				Text:    "Test feedback",
				BoardID: 999, // Board doesn't exist
			},
		}

		// Mock expectations
		mock.ExpectBegin()

		// Board doesn't exist
		mock.ExpectQuery(`SELECT \* FROM "boards"`).
			WithArgs(999, 1).
			WillReturnError(gorm.ErrRecordNotFound)

		mock.ExpectRollback()

		// Execute the function being tested
		successCount, dbErrors, err := UploadFeedbacksFromFile(testFeedbacks)

		// Assert results
		assert.Nil(t, err)
		assert.Equal(t, 0, successCount)
		assert.Len(t, dbErrors, 1)
		assert.Contains(t, dbErrors[0], "board with ID 999 does not exist")

		// Verify all expectations were met
		err = mock.ExpectationsWereMet()
		assert.NoError(t, err)
	})

	// Test with empty feedbacks list
	t.Run("Empty feedbacks list", func(t *testing.T) {
		// Define empty test data
		var testFeedbacks []Feedback.Feedback

		// Execute the function being tested
		successCount, dbErrors, err := UploadFeedbacksFromFile(testFeedbacks)

		// Assert results
		assert.Nil(t, err)
		assert.Equal(t, 0, successCount)
		assert.Empty(t, dbErrors)
		// No DB operations should occur for empty input
	})
}
