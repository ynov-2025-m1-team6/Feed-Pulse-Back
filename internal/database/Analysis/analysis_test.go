package Analysis

import (
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/database"
	"github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/models/Analysis"
	"github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/models/BaseModel"
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

func TestGetAllAnalyses(t *testing.T) {
	mock, err := setupTest()
	if err != nil {
		t.Fatalf("Error setting up test: %v", err)
	}

	// Define test data
	testAnalyses := []Analysis.Analysis{
		{
			BaseModel:      BaseModel.BaseModel{Id: 1},
			FeedbackID:     1,
			Topic:          "Product",
			SentimentScore: 0.8,
		},
		{
			BaseModel:      BaseModel.BaseModel{Id: 2},
			FeedbackID:     2,
			Topic:          "Service",
			SentimentScore: 0.5,
		},
	}

	// Setup expectations
	rows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "feedback_id", "topic", "sentiment_score"})
	for _, a := range testAnalyses {
		rows.AddRow(a.Id, a.CreatedAt, a.UpdatedAt, a.FeedbackID, a.Topic, a.SentimentScore)
	}

	// Match any SELECT query
	mock.ExpectQuery(`SELECT (.+) FROM "analyses"`).
		WillReturnRows(rows)

	// Call the function we're testing
	analyses, err := GetAllAnalyses()

	// Assert expectations
	assert.Nil(t, err)
	assert.Len(t, analyses, 2)
	if len(analyses) >= 2 {
		assert.Equal(t, testAnalyses[0].Id, analyses[0].BaseModel.Id)
		assert.Equal(t, testAnalyses[1].Id, analyses[1].BaseModel.Id)
		assert.Equal(t, testAnalyses[0].FeedbackID, analyses[0].FeedbackID)
		assert.Equal(t, testAnalyses[1].FeedbackID, analyses[1].FeedbackID)
		assert.Equal(t, testAnalyses[0].Topic, analyses[0].Topic)
		assert.Equal(t, testAnalyses[1].Topic, analyses[1].Topic)
		assert.Equal(t, testAnalyses[0].SentimentScore, analyses[0].SentimentScore)
		assert.Equal(t, testAnalyses[1].SentimentScore, analyses[1].SentimentScore)
	}

	// Test with database error
	mock.ExpectQuery(`SELECT (.+) FROM "analyses"`).
		WillReturnError(errors.New("database error"))

	_, err = GetAllAnalyses()
	assert.NotNil(t, err)
}

func TestGetAnalysisByID(t *testing.T) {
	mock, err := setupTest()
	if err != nil {
		t.Fatalf("Error setting up test: %v", err)
	}

	// Define test data
	testAnalysis := Analysis.Analysis{
		BaseModel:      BaseModel.BaseModel{Id: 1},
		FeedbackID:     1,
		Topic:          "Product",
		SentimentScore: 0.8,
	}

	// Test successful case
	rows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "feedback_id", "topic", "sentiment_score"}).
		AddRow(testAnalysis.Id, testAnalysis.CreatedAt, testAnalysis.UpdatedAt, testAnalysis.FeedbackID, testAnalysis.Topic, testAnalysis.SentimentScore)

	// GORM generates a query with multiple params, one for ID and one for LIMIT
	mock.ExpectQuery(`SELECT (.+) FROM "analyses" WHERE (.+)`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(rows)

	analysis, err := GetAnalysisByID(1)
	assert.Nil(t, err)
	assert.Equal(t, testAnalysis.Id, analysis.BaseModel.Id)
	assert.Equal(t, testAnalysis.FeedbackID, analysis.FeedbackID)
	assert.Equal(t, testAnalysis.Topic, analysis.Topic)
	assert.Equal(t, testAnalysis.SentimentScore, analysis.SentimentScore)

	// Test not found error
	mock.ExpectQuery(`SELECT (.+) FROM "analyses" WHERE (.+)`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnError(gorm.ErrRecordNotFound)

	_, err = GetAnalysisByID(999)
	assert.NotNil(t, err)
	assert.Equal(t, "analysis not found", err.Error())

	// Test other error
	mock.ExpectQuery(`SELECT (.+) FROM "analyses" WHERE (.+)`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnError(errors.New("database error"))

	_, err = GetAnalysisByID(2)
	assert.NotNil(t, err)
}

func TestAddAnalysis(t *testing.T) {
	mock, err := setupTest()
	if err != nil {
		t.Fatalf("Error setting up test: %v", err)
	}

	// Define test data
	testAnalysis := Analysis.Analysis{
		FeedbackID:     1,
		Topic:          "Product",
		SentimentScore: 0.8,
	}

	// Setup expectations for checking if feedback exists
	feedbackRows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "date", "channel", "text"}).
		AddRow(1, time.Now(), time.Now(), time.Now(), "email", "Test feedback")

	// The AddAnalysis function first checks if the feedback exists with the FeedbackID
	mock.ExpectQuery(`SELECT \* FROM "feedbacks" WHERE id = \$1 ORDER BY "feedbacks"."id" LIMIT \$2`).
		WithArgs(1, 1).
		WillReturnRows(feedbackRows)

	// Setup expectations for the create operation
	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "analyses"`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), 0.8, "Product", 1).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()

	// Call the function we're testing
	createdAnalysis, err := AddAnalysis(testAnalysis)

	// Assert expectations
	assert.Nil(t, err)
	assert.Equal(t, testAnalysis.FeedbackID, createdAnalysis.FeedbackID)
	assert.Equal(t, testAnalysis.Topic, createdAnalysis.Topic)
	assert.Equal(t, testAnalysis.SentimentScore, createdAnalysis.SentimentScore)

	// Test with feedback not found error
	mock.ExpectQuery(`SELECT \* FROM "feedbacks" WHERE id = \$1 ORDER BY "feedbacks"."id" LIMIT \$2`).
		WithArgs(1, 1).
		WillReturnError(gorm.ErrRecordNotFound)

	_, err = AddAnalysis(testAnalysis)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "feedback not found")

	// Test with database error during analysis creation
	mock.ExpectQuery(`SELECT \* FROM "feedbacks" WHERE id = \$1 ORDER BY "feedbacks"."id" LIMIT \$2`).
		WithArgs(1, 1).
		WillReturnRows(feedbackRows)

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "analyses"`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), 0.8, "Product", 1).
		WillReturnError(errors.New("database error"))
	mock.ExpectRollback()

	_, err = AddAnalysis(testAnalysis)
	assert.NotNil(t, err)
}

func TestUpdateAnalysis(t *testing.T) {
	mock, err := setupTest()
	if err != nil {
		t.Fatalf("Error setting up test: %v", err)
	}

	// Define test data
	testAnalysis := Analysis.Analysis{
		BaseModel:      BaseModel.BaseModel{Id: 1},
		FeedbackID:     1,
		Topic:          "Updated Topic",
		SentimentScore: 0.9,
	}

	// Setup expectations for fetch operation (check if exists)
	rows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "feedback_id", "topic", "sentiment_score"}).
		AddRow(1, time.Now(), time.Now(), 1, "Old Topic", 0.5)

	mock.ExpectQuery(`SELECT (.+) FROM "analyses" WHERE (.+)`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(rows)

	// Setup expectations for the update operation
	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "analyses" SET (.+) WHERE (.+)`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	// Call the function we're testing
	updatedAnalysis, err := UpdateAnalysis(testAnalysis)

	// Assert expectations
	assert.Nil(t, err)
	assert.Equal(t, testAnalysis.Id, updatedAnalysis.BaseModel.Id)
	assert.Equal(t, testAnalysis.Topic, updatedAnalysis.Topic)
	assert.Equal(t, testAnalysis.SentimentScore, updatedAnalysis.SentimentScore)

	// Test not found error
	mock.ExpectQuery(`SELECT (.+) FROM "analyses" WHERE (.+)`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnError(gorm.ErrRecordNotFound)

	_, err = UpdateAnalysis(Analysis.Analysis{BaseModel: BaseModel.BaseModel{Id: 999}})
	assert.NotNil(t, err)
	assert.Equal(t, "analysis not found", err.Error())

	// Test other error during fetch
	mock.ExpectQuery(`SELECT (.+) FROM "analyses" WHERE (.+)`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnError(errors.New("database error"))

	_, err = UpdateAnalysis(Analysis.Analysis{BaseModel: BaseModel.BaseModel{Id: 2}})
	assert.NotNil(t, err)

	// Test error during update
	mock.ExpectQuery(`SELECT (.+) FROM "analyses" WHERE (.+)`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(3))

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "analyses" SET (.+) WHERE (.+)`).
		WillReturnError(errors.New("update error"))
	mock.ExpectRollback()

	_, err = UpdateAnalysis(Analysis.Analysis{BaseModel: BaseModel.BaseModel{Id: 3}})
	assert.NotNil(t, err)
}

func TestDeleteAnalysis(t *testing.T) {
	mock, err := setupTest()
	if err != nil {
		t.Fatalf("Error setting up test: %v", err)
	}

	// Setup expectations for fetch operation (check if exists)
	rows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "feedback_id", "topic", "sentiment_score"}).
		AddRow(1, time.Now(), time.Now(), 1, "Topic", 0.8)

	mock.ExpectQuery(`SELECT (.+) FROM "analyses" WHERE (.+)`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(rows)

	// Setup expectations for the delete operation
	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM "analyses" WHERE (.+)`).
		WithArgs(sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1)) // 1 row affected
	mock.ExpectCommit()

	// Call the function we're testing
	err = DeleteAnalysis(1)

	// Assert expectations
	assert.Nil(t, err)

	// Test not found error
	mock.ExpectQuery(`SELECT (.+) FROM "analyses" WHERE (.+)`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnError(gorm.ErrRecordNotFound)

	err = DeleteAnalysis(999)
	assert.NotNil(t, err)
	assert.Equal(t, "analysis not found", err.Error())

	// Test error during delete
	mock.ExpectQuery(`SELECT (.+) FROM "analyses" WHERE (.+)`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(2))

	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM "analyses" WHERE (.+)`).
		WithArgs(sqlmock.AnyArg()).
		WillReturnError(errors.New("delete error"))
	mock.ExpectRollback()

	err = DeleteAnalysis(2)
	assert.NotNil(t, err)
}

func TestGetAnalysisByFeedbackID(t *testing.T) {
	mock, err := setupTest()
	if err != nil {
		t.Fatalf("Error setting up test: %v", err)
	}

	// Define test data
	testAnalysis := Analysis.Analysis{
		BaseModel:      BaseModel.BaseModel{Id: 1},
		FeedbackID:     1,
		Topic:          "Product",
		SentimentScore: 0.8,
	}

	// Setup expectations
	rows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "feedback_id", "topic", "sentiment_score"}).
		AddRow(testAnalysis.Id, testAnalysis.CreatedAt, testAnalysis.UpdatedAt, testAnalysis.FeedbackID, testAnalysis.Topic, testAnalysis.SentimentScore)

	mock.ExpectQuery(`SELECT (.+) FROM "analyses" WHERE (.+)`).
		WithArgs(1, sqlmock.AnyArg()).
		WillReturnRows(rows)

	// Call the function we're testing
	analysis, err := GetAnalysisByFeedbackID(1)

	// Assert expectations
	assert.Nil(t, err)
	assert.Equal(t, testAnalysis.Id, analysis.BaseModel.Id)
	assert.Equal(t, testAnalysis.FeedbackID, analysis.FeedbackID)
	assert.Equal(t, testAnalysis.Topic, analysis.Topic)
	assert.Equal(t, testAnalysis.SentimentScore, analysis.SentimentScore)

	// Test with database error
	mock.ExpectQuery(`SELECT (.+) FROM "analyses" WHERE (.+)`).
		WithArgs(2, sqlmock.AnyArg()).
		WillReturnError(errors.New("database error"))

	_, err = GetAnalysisByFeedbackID(2)
	assert.NotNil(t, err)
}

func TestGetAnalysesByTopic(t *testing.T) {
	mock, err := setupTest()
	if err != nil {
		t.Fatalf("Error setting up test: %v", err)
	}

	// Define test data
	testAnalyses := []Analysis.Analysis{
		{
			BaseModel:      BaseModel.BaseModel{Id: 1},
			FeedbackID:     1,
			Topic:          "Product",
			SentimentScore: 0.8,
		},
		{
			BaseModel:      BaseModel.BaseModel{Id: 2},
			FeedbackID:     2,
			Topic:          "Product",
			SentimentScore: 0.6,
		},
	}

	// Setup expectations
	rows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "feedback_id", "topic", "sentiment_score"})
	for _, a := range testAnalyses {
		rows.AddRow(a.Id, a.CreatedAt, a.UpdatedAt, a.FeedbackID, a.Topic, a.SentimentScore)
	}

	mock.ExpectQuery(`SELECT (.+) FROM "analyses" WHERE (.+)`).
		WithArgs("Product").
		WillReturnRows(rows)

	// Call the function we're testing
	analyses, err := GetAnalysesByTopic("Product")

	// Assert expectations
	assert.Nil(t, err)
	assert.Len(t, analyses, 2)
	if len(analyses) >= 2 {
		assert.Equal(t, "Product", analyses[0].Topic)
		assert.Equal(t, "Product", analyses[1].Topic)
	}

	// Test with database error
	mock.ExpectQuery(`SELECT (.+) FROM "analyses" WHERE (.+)`).
		WithArgs("Service").
		WillReturnError(errors.New("database error"))

	_, err = GetAnalysesByTopic("Service")
	assert.NotNil(t, err)
}

func TestGetAnalysesBySentimentRange(t *testing.T) {
	mock, err := setupTest()
	if err != nil {
		t.Fatalf("Error setting up test: %v", err)
	}

	// Define test data
	testAnalyses := []Analysis.Analysis{
		{
			BaseModel:      BaseModel.BaseModel{Id: 1},
			FeedbackID:     1,
			Topic:          "Product",
			SentimentScore: 0.6,
		},
		{
			BaseModel:      BaseModel.BaseModel{Id: 2},
			FeedbackID:     2,
			Topic:          "Service",
			SentimentScore: 0.7,
		},
	}

	// Setup expectations
	rows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "feedback_id", "topic", "sentiment_score"})
	for _, a := range testAnalyses {
		rows.AddRow(a.Id, a.CreatedAt, a.UpdatedAt, a.FeedbackID, a.Topic, a.SentimentScore)
	}

	mock.ExpectQuery(`SELECT (.+) FROM "analyses" WHERE (.+)`).
		WithArgs(0.5, 0.8).
		WillReturnRows(rows)

	// Call the function we're testing
	analyses, err := GetAnalysesBySentimentRange(0.5, 0.8)

	// Assert expectations
	assert.Nil(t, err)
	assert.Len(t, analyses, 2)
	if len(analyses) >= 2 {
		assert.GreaterOrEqual(t, analyses[0].SentimentScore, 0.5)
		assert.LessOrEqual(t, analyses[0].SentimentScore, 0.8)
		assert.GreaterOrEqual(t, analyses[1].SentimentScore, 0.5)
		assert.LessOrEqual(t, analyses[1].SentimentScore, 0.8)
	}

	// Test with database error
	mock.ExpectQuery(`SELECT (.+) FROM "analyses" WHERE (.+)`).
		WithArgs(-1.0, 1.0).
		WillReturnError(errors.New("database error"))

	_, err = GetAnalysesBySentimentRange(-1.0, 1.0)
	assert.NotNil(t, err)
}
