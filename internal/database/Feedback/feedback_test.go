package Feedback

import (
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/database"
	"github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/models/BaseModel"
	"github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/models/Feedback"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// setupTest creates a mock database connection for testing
func setupTest() (sqlmock.Sqlmock, error) {
	// Create a mock database connection
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	if err != nil {
		return nil, err
	}

	pgDB := postgres.New(postgres.Config{
		Conn:                 db,
		PreferSimpleProtocol: true,
	})

	gormDB, err := gorm.Open(pgDB, &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// Replace the global DB with our mock
	database.DB = gormDB

	return mock, nil
}

func TestGetAllFeedbacks(t *testing.T) {
	mock, err := setupTest()
	if err != nil {
		t.Fatalf("Error setting up test: %v", err)
	}

	// Define test data
	testDate := time.Now()
	testFeedbacks := []Feedback.Feedback{
		{
			BaseModel: BaseModel.BaseModel{Id: 1},
			Date:      testDate,
			Channel:   "email",
			Text:      "Test feedback 1",
		},
		{
			BaseModel: BaseModel.BaseModel{Id: 2},
			Date:      testDate,
			Channel:   "web",
			Text:      "Test feedback 2",
		},
	}

	// Setup expectations
	rows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "date", "channel", "text"})
	for _, f := range testFeedbacks {
		rows.AddRow(f.Id, f.CreatedAt, f.UpdatedAt, f.Date, f.Channel, f.Text)
	}

	// Match any SELECT query
	mock.ExpectQuery(`SELECT (.+) FROM "feedbacks"`).
		WillReturnRows(rows)

	// Call the function we're testing
	feedbacks, err := GetAllFeedbacks()

	// Assert expectations
	assert.Nil(t, err)
	assert.Len(t, feedbacks, 2)
	if len(feedbacks) >= 2 {
		assert.Equal(t, testFeedbacks[0].Id, feedbacks[0].Id)
		assert.Equal(t, testFeedbacks[1].Id, feedbacks[1].Id)
		assert.Equal(t, testFeedbacks[0].Channel, feedbacks[0].Channel)
		assert.Equal(t, testFeedbacks[1].Channel, feedbacks[1].Channel)
		assert.Equal(t, testFeedbacks[0].Text, feedbacks[0].Text)
		assert.Equal(t, testFeedbacks[1].Text, feedbacks[1].Text)
	}

	// Test with database error
	mock.ExpectQuery(`SELECT (.+) FROM "feedbacks"`).
		WillReturnError(errors.New("database error"))

	_, err = GetAllFeedbacks()
	assert.NotNil(t, err)
}

func TestGetFeedbackByID(t *testing.T) {
	mock, err := setupTest()
	if err != nil {
		t.Fatalf("Error setting up test: %v", err)
	}

	// Define test data
	testDate := time.Now()
	testFeedback := Feedback.Feedback{
		BaseModel: BaseModel.BaseModel{Id: 1},
		Date:      testDate,
		Channel:   "email",
		Text:      "Test feedback 1",
	}

	// Test successful case
	rows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "date", "channel", "text"}).
		AddRow(testFeedback.Id, testFeedback.CreatedAt, testFeedback.UpdatedAt, testFeedback.Date, testFeedback.Channel, testFeedback.Text)

	// GORM generates a query with multiple params, one for ID and one for LIMIT
	mock.ExpectQuery(`SELECT (.+) FROM "feedbacks" WHERE (.+)`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(rows)

	feedback, err := GetFeedbackByID(1)
	assert.Nil(t, err)
	assert.Equal(t, testFeedback.Id, feedback.Id)
	assert.Equal(t, testFeedback.Channel, feedback.Channel)
	assert.Equal(t, testFeedback.Text, feedback.Text)

	// Test not found error
	mock.ExpectQuery(`SELECT (.+) FROM "feedbacks" WHERE (.+)`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnError(gorm.ErrRecordNotFound)

	_, err = GetFeedbackByID(999)
	assert.NotNil(t, err)
	assert.Equal(t, "feedback not found", err.Error())

	// Test other error
	mock.ExpectQuery(`SELECT (.+) FROM "feedbacks" WHERE (.+)`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnError(errors.New("database error"))

	_, err = GetFeedbackByID(2)
	assert.NotNil(t, err)
}

func TestCreateFeedback(t *testing.T) {
	mock, err := setupTest()
	if err != nil {
		t.Fatalf("Error setting up test: %v", err)
	}

	// Define test data
	testDate := time.Now()
	testFeedback := Feedback.Feedback{
		Date:    testDate,
		Channel: "email",
		Text:    "Test feedback",
	}

	// Setup expectations for the create operation
	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "feedbacks"`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), testDate, "email", "Test feedback").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()

	// Call the function we're testing
	createdFeedback, err := CreateFeedback(testFeedback)

	// Assert expectations
	assert.Nil(t, err)
	assert.Equal(t, testFeedback.Channel, createdFeedback.Channel)
	assert.Equal(t, testFeedback.Text, createdFeedback.Text)

	// Test with database error
	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "feedbacks"`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), testDate, "email", "Test feedback").
		WillReturnError(errors.New("database error"))
	mock.ExpectRollback()

	_, err = CreateFeedback(testFeedback)
	assert.NotNil(t, err)
}

func TestUpdateFeedback(t *testing.T) {
	mock, err := setupTest()
	if err != nil {
		t.Fatalf("Error setting up test: %v", err)
	}

	// Define test data
	testDate := time.Now()
	testFeedback := Feedback.Feedback{
		BaseModel: BaseModel.BaseModel{Id: 1},
		Date:      testDate,
		Channel:   "email",
		Text:      "Updated feedback",
	}

	// Setup expectations for fetch operation (check if exists)
	rows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "date", "channel", "text"}).
		AddRow(1, time.Now(), time.Now(), testDate, "old_channel", "old_text")

	// The SELECT query has multiple args (id and limit)
	mock.ExpectQuery(`SELECT (.+) FROM "feedbacks" WHERE (.+)`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(rows)

	// Setup expectations for the update operation
	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "feedbacks" SET (.+) WHERE (.+)`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	// Call the function we're testing
	updatedFeedback, err := UpdateFeedback(testFeedback)

	// Assert expectations
	assert.Nil(t, err)
	assert.Equal(t, testFeedback.Id, updatedFeedback.Id)
	assert.Equal(t, testFeedback.Channel, updatedFeedback.Channel)
	assert.Equal(t, testFeedback.Text, updatedFeedback.Text)

	// Test not found error
	mock.ExpectQuery(`SELECT (.+) FROM "feedbacks" WHERE (.+)`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnError(gorm.ErrRecordNotFound)

	_, err = UpdateFeedback(Feedback.Feedback{BaseModel: BaseModel.BaseModel{Id: 999}})
	assert.NotNil(t, err)
	assert.Equal(t, "feedback not found", err.Error())

	// Test other error during fetch
	mock.ExpectQuery(`SELECT (.+) FROM "feedbacks" WHERE (.+)`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnError(errors.New("database error"))

	_, err = UpdateFeedback(Feedback.Feedback{BaseModel: BaseModel.BaseModel{Id: 2}})
	assert.NotNil(t, err)

	// Test error during update
	mock.ExpectQuery(`SELECT (.+) FROM "feedbacks" WHERE (.+)`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(3))

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "feedbacks" SET (.+) WHERE (.+)`).
		WillReturnError(errors.New("update error"))
	mock.ExpectRollback()

	_, err = UpdateFeedback(Feedback.Feedback{BaseModel: BaseModel.BaseModel{Id: 3}})
	assert.NotNil(t, err)
}

func TestDeleteFeedback(t *testing.T) {
	mock, err := setupTest()
	if err != nil {
		t.Fatalf("Error setting up test: %v", err)
	}

	// Setup expectations for fetch operation (check if exists)
	rows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "date", "channel", "text"}).
		AddRow(1, time.Now(), time.Now(), time.Now(), "email", "Test feedback")

	mock.ExpectQuery(`SELECT (.+) FROM "feedbacks" WHERE (.+)`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(rows)

	// Setup expectations for transaction
	mock.ExpectBegin()

	// Expect deletion of associated analyses
	mock.ExpectExec(`DELETE FROM "analyses" WHERE (.+)`).
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 2)) // Assume 2 analyses were deleted

	// Expect deletion of feedback
	mock.ExpectExec(`DELETE FROM "feedbacks" WHERE (.+)`).
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 1)) // 1 feedback deleted

	mock.ExpectCommit()

	// Call the function we're testing
	err = DeleteFeedback(1)

	// Assert expectations
	assert.Nil(t, err)

	// Test not found error
	mock.ExpectQuery(`SELECT (.+) FROM "feedbacks" WHERE (.+)`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnError(gorm.ErrRecordNotFound)

	err = DeleteFeedback(999)
	assert.NotNil(t, err)
	assert.Equal(t, "feedback not found", err.Error())

	// Test error during delete analyses
	mock.ExpectQuery(`SELECT (.+) FROM "feedbacks" WHERE (.+)`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(2))

	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM "analyses" WHERE (.+)`).
		WithArgs(2).
		WillReturnError(errors.New("delete error"))
	mock.ExpectRollback()

	err = DeleteFeedback(2)
	assert.NotNil(t, err)
}

func TestGetFeedbacksByChannel(t *testing.T) {
	mock, err := setupTest()
	if err != nil {
		t.Fatalf("Error setting up test: %v", err)
	}

	// Define test data
	testDate := time.Now()
	testFeedbacks := []Feedback.Feedback{
		{
			BaseModel: BaseModel.BaseModel{Id: 1},
			Date:      testDate,
			Channel:   "email",
			Text:      "Test feedback 1",
		},
		{
			BaseModel: BaseModel.BaseModel{Id: 2},
			Date:      testDate,
			Channel:   "email",
			Text:      "Test feedback 2",
		},
	}

	// Setup expectations
	rows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "date", "channel", "text"})
	for _, f := range testFeedbacks {
		rows.AddRow(f.Id, f.CreatedAt, f.UpdatedAt, f.Date, f.Channel, f.Text)
	}

	mock.ExpectQuery(`SELECT (.+) FROM "feedbacks" WHERE (.+)`).
		WithArgs("email").
		WillReturnRows(rows)

	// Call the function we're testing
	feedbacks, err := GetFeedbacksByChannel("email")

	// Assert expectations
	assert.Nil(t, err)
	assert.Len(t, feedbacks, 2)
	if len(feedbacks) >= 2 {
		assert.Equal(t, "email", feedbacks[0].Channel)
		assert.Equal(t, "email", feedbacks[1].Channel)
	}

	// Test with database error
	mock.ExpectQuery(`SELECT (.+) FROM "feedbacks" WHERE (.+)`).
		WithArgs("web").
		WillReturnError(errors.New("database error"))

	_, err = GetFeedbacksByChannel("web")
	assert.NotNil(t, err)
}
