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
		// Setup test DB
		mockDB, mock, err := sqlmock.New()
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

		// Expect first feedback insert - use MatchAnyArgs to avoid argument count issues
		mock.ExpectQuery(`INSERT INTO "feedbacks"`).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

		// Expect second feedback insert
		mock.ExpectQuery(`INSERT INTO "feedbacks"`).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(2))

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
		mockDB, mock, err := sqlmock.New()
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
		assert.Equal(t, "transaction begin error", err.Error())
		assert.Equal(t, 0, successCount)
		assert.Empty(t, dbErrors)

		// Verify all expectations were met
		err = mock.ExpectationsWereMet()
		assert.NoError(t, err)
	})

	t.Run("Error creating some feedbacks", func(t *testing.T) {
		// Setup test DB
		mockDB, mock, err := sqlmock.New()
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
				BoardID: 999, // Invalid board ID to trigger error
			},
		}

		// Mock expectations
		mock.ExpectBegin()

		// First feedback succeeds
		mock.ExpectQuery(`INSERT INTO "feedbacks"`).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

		// Second feedback fails
		mock.ExpectQuery(`INSERT INTO "feedbacks"`).
			WillReturnError(errors.New("foreign key violation"))

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
		mockDB, mock, err := sqlmock.New()
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
				BoardID: 999, // Invalid board ID
			},
		}

		// Mock expectations
		mock.ExpectBegin()
		mock.ExpectQuery(`INSERT INTO "feedbacks"`).
			WillReturnError(errors.New("foreign key violation"))
		mock.ExpectRollback()

		// Execute the function being tested
		successCount, dbErrors, err := FetchAndSaveFeedbacks(testFeedbacks)

		// Assert results
		assert.Nil(t, err)
		assert.Equal(t, 0, successCount)
		assert.Len(t, dbErrors, 1)
		assert.Contains(t, dbErrors[0], "foreign key violation")

		// Verify all expectations were met
		err = mock.ExpectationsWereMet()
		assert.NoError(t, err)
	})

	t.Run("Error during commit", func(t *testing.T) {
		// Setup test DB
		mockDB, mock, err := sqlmock.New()
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
		mock.ExpectQuery(`INSERT INTO "feedbacks"`).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
		mock.ExpectCommit().WillReturnError(errors.New("commit error"))

		// Execute the function being tested
		successCount, dbErrors, err := FetchAndSaveFeedbacks(testFeedbacks)

		// Assert results
		assert.Error(t, err)
		assert.Equal(t, "commit error", err.Error())
		assert.Equal(t, 1, successCount) // Still counts as successful creation before commit error
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
