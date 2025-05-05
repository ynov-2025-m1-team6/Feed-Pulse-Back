package Feedback

import (
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/database"
	"github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/models/Feedback"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestFetchAndSaveFeedbacks(t *testing.T) {
	// Create fresh mock for each test
	t.Run("Successfully fetch and save multiple feedbacks", func(t *testing.T) {
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

		// Expect board validation query - only needs to happen once since both feedbacks use the same board
		boardRows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "name"}).
			AddRow(1, time.Now(), time.Now(), "Test Board")
		mock.ExpectQuery(`SELECT \* FROM "boards"`).
			WithArgs(1, 1).
			WillReturnRows(boardRows)

		// For the first feedback, expect the validation check inside CreateFeedback
		boardRows1 := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "name"}).
			AddRow(1, time.Now(), time.Now(), "Test Board")
		mock.ExpectQuery(`SELECT \* FROM "boards"`).
			WithArgs(1, 1).
			WillReturnRows(boardRows1)

		// Then expect the insert for the first feedback
		mock.ExpectBegin()
		mock.ExpectQuery(`INSERT INTO "feedbacks"`).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
		mock.ExpectCommit()

		// For the second feedback, expect the validation check inside CreateFeedback
		boardRows2 := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "name"}).
			AddRow(1, time.Now(), time.Now(), "Test Board")
		mock.ExpectQuery(`SELECT \* FROM "boards"`).
			WithArgs(1, 1).
			WillReturnRows(boardRows2)

		// Then expect the insert for the second feedback
		mock.ExpectBegin()
		mock.ExpectQuery(`INSERT INTO "feedbacks"`).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(2))
		mock.ExpectCommit()

		mock.ExpectCommit()

		// Execute the function being tested
		successCount, errs, err := FetchAndSaveFeedbacks(testFeedbacks)

		// Assert results
		assert.Nil(t, err)
		assert.Equal(t, len(testFeedbacks), successCount)
		assert.Empty(t, errs)

		// Verify all expectations were met
		err = mock.ExpectationsWereMet()
		assert.NoError(t, err)
	})

	t.Run("Transaction begin failure", func(t *testing.T) {
		// Setup test DB
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
		testFeedbacks := []Feedback.Feedback{
			{
				Date:    time.Now(),
				Channel: "email",
				Text:    "Test feedback",
				BoardID: 1,
			},
		}

		// Mock expectations for transaction failure
		mock.ExpectBegin().WillReturnError(errors.New("transaction begin error"))

		// Execute the function being tested
		successCount, dbErrors, err := FetchAndSaveFeedbacks(testFeedbacks)

		// Assert results
		assert.Error(t, err)
		assert.Equal(t, "failed to begin transaction: transaction begin error", err.Error())
		assert.Equal(t, 0, successCount)
		assert.Empty(t, dbErrors)

		// Verify all expectations were met
		err = mock.ExpectationsWereMet()
		assert.NoError(t, err)
	})

	t.Run("Error creating some feedbacks", func(t *testing.T) {
		// Setup test DB
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
		testFeedbacks := []Feedback.Feedback{
			{
				Date:    time.Now(),
				Channel: "email",
				Text:    "Test feedback 1",
				BoardID: 1,
			},
			{
				Date:    time.Now(),
				Channel: "chat",
				Text:    "Test feedback 2",
				BoardID: 2, // Different board ID
			},
		}

		// Mock expectations
		mock.ExpectBegin()

		// Board 1 exists - expected in the main transaction
		boardRows1 := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "name"}).
			AddRow(1, time.Now(), time.Now(), "Test Board 1")
		mock.ExpectQuery(`SELECT \* FROM "boards"`).
			WithArgs(1, 1).
			WillReturnRows(boardRows1)

		// Board 1 exists - expected in CreateFeedback call
		boardRows1Again := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "name"}).
			AddRow(1, time.Now(), time.Now(), "Test Board 1")
		mock.ExpectQuery(`SELECT \* FROM "boards"`).
			WithArgs(1, 1).
			WillReturnRows(boardRows1Again)

		// First feedback succeeds
		mock.ExpectBegin()
		mock.ExpectQuery(`INSERT INTO "feedbacks"`).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
		mock.ExpectCommit()

		// Board 2 exists - expected in the main transaction
		boardRows2 := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "name"}).
			AddRow(2, time.Now(), time.Now(), "Test Board 2")
		mock.ExpectQuery(`SELECT \* FROM "boards"`).
			WithArgs(2, 1).
			WillReturnRows(boardRows2)

		// Board 2 exists - expected in CreateFeedback call
		boardRows2Again := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "name"}).
			AddRow(2, time.Now(), time.Now(), "Test Board 2")
		mock.ExpectQuery(`SELECT \* FROM "boards"`).
			WithArgs(2, 1).
			WillReturnRows(boardRows2Again)

		// Second feedback fails
		mock.ExpectBegin()
		mock.ExpectQuery(`INSERT INTO "feedbacks"`).
			WillReturnError(errors.New("foreign key violation"))
		mock.ExpectRollback()

		mock.ExpectCommit()

		// Execute the function being tested
		successCount, dbErrors, err := FetchAndSaveFeedbacks(testFeedbacks)

		// Assert results
		assert.Nil(t, err)
		assert.Equal(t, 1, successCount)
		assert.Len(t, dbErrors, 1)
		assert.Contains(t, dbErrors[0], "foreign key violation")

		// Verify all expectations were met
		err = mock.ExpectationsWereMet()
		assert.NoError(t, err)
	})

	t.Run("No successful creations should rollback", func(t *testing.T) {
		// Setup test DB
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
		testFeedbacks := []Feedback.Feedback{
			{
				Date:    time.Now(),
				Channel: "email",
				Text:    "Test feedback 1",
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
		successCount, dbErrors, err := FetchAndSaveFeedbacks(testFeedbacks)

		// Assert results
		assert.Nil(t, err)
		assert.Equal(t, 0, successCount)
		assert.Len(t, dbErrors, 1)
		assert.Contains(t, dbErrors[0], "board with ID 999 does not exist")

		// Verify all expectations were met
		err = mock.ExpectationsWereMet()
		assert.NoError(t, err)
	})

	t.Run("Error during commit", func(t *testing.T) {
		// Setup test DB
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
		testFeedbacks := []Feedback.Feedback{
			{
				Date:    time.Now(),
				Channel: "email",
				Text:    "Test feedback 1",
				BoardID: 1,
			},
		}

		// Mock expectations
		mock.ExpectBegin()

		// Board exists - expected in the main transaction
		boardRows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "name"}).
			AddRow(1, time.Now(), time.Now(), "Test Board")
		mock.ExpectQuery(`SELECT \* FROM "boards"`).
			WithArgs(1, 1).
			WillReturnRows(boardRows)

		// Board exists - expected in CreateFeedback call
		boardRowsAgain := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "name"}).
			AddRow(1, time.Now(), time.Now(), "Test Board")
		mock.ExpectQuery(`SELECT \* FROM "boards"`).
			WithArgs(1, 1).
			WillReturnRows(boardRowsAgain)

		// Feedback creation succeeds
		mock.ExpectBegin()
		mock.ExpectQuery(`INSERT INTO "feedbacks"`).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
		mock.ExpectCommit()

		// But transaction commit fails
		mock.ExpectCommit().WillReturnError(errors.New("commit error"))

		// Execute the function being tested
		successCount, dbErrors, err := FetchAndSaveFeedbacks(testFeedbacks)

		// Assert results
		assert.Error(t, err)
		assert.Equal(t, "failed to commit transaction: commit error", err.Error())
		assert.Equal(t, 1, successCount)
		assert.Empty(t, dbErrors)

		// Verify all expectations were met
		err = mock.ExpectationsWereMet()
		assert.NoError(t, err)
	})

	t.Run("Empty feedbacks list", func(t *testing.T) {
		// Setup test DB
		mockDB, _, err := sqlmock.New()
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

		// Define empty test data
		var testFeedbacks []Feedback.Feedback

		// Execute the function being tested
		successCount, dbErrors, err := FetchAndSaveFeedbacks(testFeedbacks)

		// Assert results
		assert.Nil(t, err)
		assert.Equal(t, 0, successCount)
		assert.Empty(t, dbErrors)

		// For empty input, no DB operations should occur, so don't check expectations
		// The function should early-return before any DB interaction
	})
}
