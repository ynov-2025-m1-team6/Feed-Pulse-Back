package calculmetric

import (
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/database"
	"github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/models/BaseModel"
	"github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/models/Feedback"
	metric "github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/models/Metric"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func setupTestCalculMetric(t *testing.T) sqlmock.Sqlmock {
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

func TestRoundToTwo(t *testing.T) {
	tests := []struct {
		name string
		num  float64
		want float64
	}{
		{
			name: "Positive number with more than 2 decimals",
			num:  3.14159,
			want: 3.14,
		},
		{
			name: "Negative number with more than 2 decimals",
			num:  -2.567,
			want: -2.57,
		},
		{
			name: "Number with exactly 2 decimals",
			num:  5.25,
			want: 5.25,
		},
		{
			name: "Whole number",
			num:  10.0,
			want: 10.0,
		},
		{
			name: "Zero",
			num:  0.0,
			want: 0.0,
		},
		{
			name: "Very small number",
			num:  0.001,
			want: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := roundToTwo(tt.num)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestCalculMetric(t *testing.T) {
	t.Run("Empty feedbacks", func(t *testing.T) {
		feedbacks := []Feedback.Feedback{}

		got, err := CalculMetric(feedbacks)
		assert.NoError(t, err)

		// Verify empty result
		assert.Empty(t, got.DistributionByChannel)
		assert.Empty(t, got.DistributionByTopic)
		assert.Empty(t, got.VolumetryByDay)
		assert.Equal(t, 0.0, got.AverageSentiment)
		assert.Equal(t, metric.Sentiment{Positive: 0, Neutral: 0, Negative: 0}, got.Sentiment)
		assert.Equal(t, 0.0, got.PercentageSentimentUnderTreshold)
	})

	t.Run("Success case with feedbacks", func(t *testing.T) {
		mock := setupTestCalculMetric(t)
		testDate := time.Now()

		feedbacks := []Feedback.Feedback{
			{
				BaseModel: BaseModel.BaseModel{
					Id:        1,
					CreatedAt: testDate,
				},
				Channel: "email",
				Text:    "Great product!",
				BoardID: 1,
			},
			{
				BaseModel: BaseModel.BaseModel{
					Id:        2,
					CreatedAt: testDate,
				},
				Channel: "web",
				Text:    "Poor service",
				BoardID: 1,
			},
		}

		// Mock analysis queries for first feedback
		mock.ExpectQuery(`SELECT \* FROM "analyses" WHERE feedback_id = \$1 ORDER BY "analyses"."id" LIMIT \$2`).
			WithArgs(1, 1).
			WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "feedback_id", "topic", "sentiment_score"}).
				AddRow(1, testDate, testDate, 1, "Support Client", 0.7))

		// Mock analysis queries for second feedback
		mock.ExpectQuery(`SELECT \* FROM "analyses" WHERE feedback_id = \$1 ORDER BY "analyses"."id" LIMIT \$2`).
			WithArgs(2, 1).
			WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "feedback_id", "topic", "sentiment_score"}).
				AddRow(2, testDate, testDate, 2, "Performance", -0.6))

		// Mock analysis queries for average sentiment calculation
		mock.ExpectQuery(`SELECT \* FROM "analyses" WHERE feedback_id = \$1 ORDER BY "analyses"."id" LIMIT \$2`).
			WithArgs(1, 1).
			WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "feedback_id", "topic", "sentiment_score"}).
				AddRow(1, testDate, testDate, 1, "Support Client", 0.7))

		mock.ExpectQuery(`SELECT \* FROM "analyses" WHERE feedback_id = \$1 ORDER BY "analyses"."id" LIMIT \$2`).
			WithArgs(2, 1).
			WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "feedback_id", "topic", "sentiment_score"}).
				AddRow(2, testDate, testDate, 2, "Performance", -0.6))

		// Mock analysis queries for sentiment percentage calculation
		mock.ExpectQuery(`SELECT \* FROM "analyses" WHERE feedback_id = \$1 ORDER BY "analyses"."id" LIMIT \$2`).
			WithArgs(1, 1).
			WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "feedback_id", "topic", "sentiment_score"}).
				AddRow(1, testDate, testDate, 1, "Support Client", 0.7))

		mock.ExpectQuery(`SELECT \* FROM "analyses" WHERE feedback_id = \$1 ORDER BY "analyses"."id" LIMIT \$2`).
			WithArgs(2, 1).
			WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "feedback_id", "topic", "sentiment_score"}).
				AddRow(2, testDate, testDate, 2, "Performance", -0.6))

		// Mock analysis queries for threshold calculation
		mock.ExpectQuery(`SELECT \* FROM "analyses" WHERE feedback_id = \$1 ORDER BY "analyses"."id" LIMIT \$2`).
			WithArgs(1, 1).
			WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "feedback_id", "topic", "sentiment_score"}).
				AddRow(1, testDate, testDate, 1, "Support Client", 0.7))

		mock.ExpectQuery(`SELECT \* FROM "analyses" WHERE feedback_id = \$1 ORDER BY "analyses"."id" LIMIT \$2`).
			WithArgs(2, 1).
			WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "feedback_id", "topic", "sentiment_score"}).
				AddRow(2, testDate, testDate, 2, "Performance", -0.6))

		got, err := CalculMetric(feedbacks)
		assert.NoError(t, err)

		// Verify channel distribution
		expectedChannels := map[string]float64{"email": 50.0, "web": 50.0}
		assert.Equal(t, expectedChannels, got.DistributionByChannel)

		// Verify topic distribution
		expectedTopics := map[string]float64{"Support Client": 50.0, "Performance": 50.0}
		assert.Equal(t, expectedTopics, got.DistributionByTopic)

		// Verify volumetry by day
		dateStr := testDate.Format("2006-01-02")
		expectedVolumetry := map[string]float64{dateStr: 100.0}
		assert.Equal(t, expectedVolumetry, got.VolumetryByDay)

		// Verify average sentiment
		assert.Equal(t, 0.05, got.AverageSentiment) // (0.7 + (-0.6)) / 2 = 0.05

		// Verify sentiment percentages (0.7 > 0.25 = positive, -0.6 < -0.25 = negative)
		expectedSentiment := metric.Sentiment{Positive: 50.0, Neutral: 0.0, Negative: 50.0}
		assert.Equal(t, expectedSentiment, got.Sentiment)

		// Verify percentage under threshold (-0.6 < -0.5)
		assert.Equal(t, 50.0, got.PercentageSentimentUnderTreshold)

		// Verify that all mock expectations were met
		err = mock.ExpectationsWereMet()
		assert.NoError(t, err)
	})

	t.Run("Error in CalculDistributionByTopic", func(t *testing.T) {
		mock := setupTestCalculMetric(t)
		testDate := time.Now()

		feedbacks := []Feedback.Feedback{
			{
				BaseModel: BaseModel.BaseModel{
					Id:        1,
					CreatedAt: testDate,
				},
				Channel: "email",
				Text:    "Test feedback",
				BoardID: 1,
			},
		}

		// Mock analysis query to return error
		mock.ExpectQuery(`SELECT \* FROM "analyses" WHERE feedback_id = \$1 ORDER BY "analyses"."id" LIMIT \$2`).
			WithArgs(1, 1).
			WillReturnError(errors.New("database error"))

		_, err := CalculMetric(feedbacks)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "database error")

		// Verify that all mock expectations were met
		err = mock.ExpectationsWereMet()
		assert.NoError(t, err)
	})
}

func TestCalculVolumetryByDay(t *testing.T) {
	date1 := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	date2 := time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC)
	date3 := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC) // Same day as date1

	tests := []struct {
		name      string
		feedbacks []Feedback.Feedback
		want      map[string]float64
	}{
		{
			name:      "Empty feedbacks",
			feedbacks: []Feedback.Feedback{},
			want:      map[string]float64{},
		},
		{
			name: "Single day",
			feedbacks: []Feedback.Feedback{
				{BaseModel: BaseModel.BaseModel{CreatedAt: date1}},
			},
			want: map[string]float64{
				"2023-01-01": 100.0,
			},
		},
		{
			name: "Multiple days",
			feedbacks: []Feedback.Feedback{
				{BaseModel: BaseModel.BaseModel{CreatedAt: date1}},
				{BaseModel: BaseModel.BaseModel{CreatedAt: date2}},
			},
			want: map[string]float64{
				"2023-01-01": 50.0,
				"2023-01-02": 50.0,
			},
		},
		{
			name: "Same day multiple times",
			feedbacks: []Feedback.Feedback{
				{BaseModel: BaseModel.BaseModel{CreatedAt: date1}},
				{BaseModel: BaseModel.BaseModel{CreatedAt: date3}}, // Same day, different time
				{BaseModel: BaseModel.BaseModel{CreatedAt: date2}},
			},
			want: map[string]float64{
				"2023-01-01": 66.67,
				"2023-01-02": 33.33,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculVolumetryByDay(tt.feedbacks)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestCalculAverageSentiment(t *testing.T) {
	t.Run("Success case", func(t *testing.T) {
		mock := setupTestCalculMetric(t)
		testDate := time.Now()

		feedbacks := []Feedback.Feedback{
			{BaseModel: BaseModel.BaseModel{Id: 1}},
			{BaseModel: BaseModel.BaseModel{Id: 2}},
			{BaseModel: BaseModel.BaseModel{Id: 3}},
		}

		// Mock analysis queries
		mock.ExpectQuery(`SELECT \* FROM "analyses" WHERE feedback_id = \$1 ORDER BY "analyses"."id" LIMIT \$2`).
			WithArgs(1, 1).
			WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "feedback_id", "topic", "sentiment_score"}).
				AddRow(1, testDate, testDate, 1, "Topic1", 0.5))

		mock.ExpectQuery(`SELECT \* FROM "analyses" WHERE feedback_id = \$1 ORDER BY "analyses"."id" LIMIT \$2`).
			WithArgs(2, 1).
			WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "feedback_id", "topic", "sentiment_score"}).
				AddRow(2, testDate, testDate, 2, "Topic2", -0.3))

		mock.ExpectQuery(`SELECT \* FROM "analyses" WHERE feedback_id = \$1 ORDER BY "analyses"."id" LIMIT \$2`).
			WithArgs(3, 1).
			WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "feedback_id", "topic", "sentiment_score"}).
				AddRow(3, testDate, testDate, 3, "Topic3", 0.8))

		got := CalculAverageSentiment(feedbacks)
		expected := roundToTwo((0.5 + (-0.3) + 0.8) / 3) // = 0.33
		assert.Equal(t, expected, got)

		// Verify that all mock expectations were met
		err := mock.ExpectationsWereMet()
		assert.NoError(t, err)
	})

	t.Run("Database error", func(t *testing.T) {
		mock := setupTestCalculMetric(t)

		feedbacks := []Feedback.Feedback{
			{BaseModel: BaseModel.BaseModel{Id: 1}},
		}

		// Mock analysis query to return error
		mock.ExpectQuery(`SELECT \* FROM "analyses" WHERE feedback_id = \$1 ORDER BY "analyses"."id" LIMIT \$2`).
			WithArgs(1, 1).
			WillReturnError(errors.New("database error"))

		got := CalculAverageSentiment(feedbacks)
		assert.Equal(t, -100.0, got) // Error return value

		// Verify that all mock expectations were met
		err := mock.ExpectationsWereMet()
		assert.NoError(t, err)
	})
}

func TestCalculateSentimentPercentage(t *testing.T) {
	t.Run("All sentiment types", func(t *testing.T) {
		mock := setupTestCalculMetric(t)
		testDate := time.Now()

		feedbacks := []Feedback.Feedback{
			{BaseModel: BaseModel.BaseModel{Id: 1}}, // Positive (0.8 > 0.25)
			{BaseModel: BaseModel.BaseModel{Id: 2}}, // Negative (-0.7 < -0.25)
			{BaseModel: BaseModel.BaseModel{Id: 3}}, // Neutral (0.1 between -0.25 and 0.25)
			{BaseModel: BaseModel.BaseModel{Id: 4}}, // Positive (0.3 > 0.25)
		}

		// Mock analysis queries
		mock.ExpectQuery(`SELECT \* FROM "analyses" WHERE feedback_id = \$1 ORDER BY "analyses"."id" LIMIT \$2`).
			WithArgs(1, 1).
			WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "feedback_id", "topic", "sentiment_score"}).
				AddRow(1, testDate, testDate, 1, "Topic1", 0.8))

		mock.ExpectQuery(`SELECT \* FROM "analyses" WHERE feedback_id = \$1 ORDER BY "analyses"."id" LIMIT \$2`).
			WithArgs(2, 1).
			WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "feedback_id", "topic", "sentiment_score"}).
				AddRow(2, testDate, testDate, 2, "Topic2", -0.7))

		mock.ExpectQuery(`SELECT \* FROM "analyses" WHERE feedback_id = \$1 ORDER BY "analyses"."id" LIMIT \$2`).
			WithArgs(3, 1).
			WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "feedback_id", "topic", "sentiment_score"}).
				AddRow(3, testDate, testDate, 3, "Topic3", 0.1))

		mock.ExpectQuery(`SELECT \* FROM "analyses" WHERE feedback_id = \$1 ORDER BY "analyses"."id" LIMIT \$2`).
			WithArgs(4, 1).
			WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "feedback_id", "topic", "sentiment_score"}).
				AddRow(4, testDate, testDate, 4, "Topic4", 0.3))

		got := CalculateSentimentPercentage(feedbacks)
		expected := metric.Sentiment{
			Positive: 50.0, // 2 out of 4
			Neutral:  25.0, // 1 out of 4
			Negative: 25.0, // 1 out of 4
		}
		assert.Equal(t, expected, got)

		// Verify that all mock expectations were met
		err := mock.ExpectationsWereMet()
		assert.NoError(t, err)
	})

	t.Run("Database error - continue processing", func(t *testing.T) {
		mock := setupTestCalculMetric(t)
		testDate := time.Now()

		feedbacks := []Feedback.Feedback{
			{BaseModel: BaseModel.BaseModel{Id: 1}}, // Will error
			{BaseModel: BaseModel.BaseModel{Id: 2}}, // Will succeed
		}

		// Mock first query to return error
		mock.ExpectQuery(`SELECT \* FROM "analyses" WHERE feedback_id = \$1 ORDER BY "analyses"."id" LIMIT \$2`).
			WithArgs(1, 1).
			WillReturnError(errors.New("database error"))

		// Mock second query to succeed
		mock.ExpectQuery(`SELECT \* FROM "analyses" WHERE feedback_id = \$1 ORDER BY "analyses"."id" LIMIT \$2`).
			WithArgs(2, 1).
			WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "feedback_id", "topic", "sentiment_score"}).
				AddRow(2, testDate, testDate, 2, "Topic2", 0.8))

		got := CalculateSentimentPercentage(feedbacks)
		// Only second feedback is counted (1 positive out of 2 total = 50%)
		expected := metric.Sentiment{
			Positive: 50.0,
			Neutral:  0.0,
			Negative: 0.0,
		}
		assert.Equal(t, expected, got)

		// Verify that all mock expectations were met
		err := mock.ExpectationsWereMet()
		assert.NoError(t, err)
	})
}

func TestCalculatePercentageSentimentUnderThreshold(t *testing.T) {
	t.Run("Some feedbacks under threshold", func(t *testing.T) {
		mock := setupTestCalculMetric(t)
		testDate := time.Now()

		feedbacks := []Feedback.Feedback{
			{BaseModel: BaseModel.BaseModel{Id: 1}}, // -0.7 < -0.5 (under threshold)
			{BaseModel: BaseModel.BaseModel{Id: 2}}, // 0.3 > -0.5 (above threshold)
			{BaseModel: BaseModel.BaseModel{Id: 3}}, // -0.6 < -0.5 (under threshold)
			{BaseModel: BaseModel.BaseModel{Id: 4}}, // -0.4 > -0.5 (above threshold)
		}

		// Mock analysis queries
		mock.ExpectQuery(`SELECT \* FROM "analyses" WHERE feedback_id = \$1 ORDER BY "analyses"."id" LIMIT \$2`).
			WithArgs(1, 1).
			WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "feedback_id", "topic", "sentiment_score"}).
				AddRow(1, testDate, testDate, 1, "Topic1", -0.7))

		mock.ExpectQuery(`SELECT \* FROM "analyses" WHERE feedback_id = \$1 ORDER BY "analyses"."id" LIMIT \$2`).
			WithArgs(2, 1).
			WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "feedback_id", "topic", "sentiment_score"}).
				AddRow(2, testDate, testDate, 2, "Topic2", 0.3))

		mock.ExpectQuery(`SELECT \* FROM "analyses" WHERE feedback_id = \$1 ORDER BY "analyses"."id" LIMIT \$2`).
			WithArgs(3, 1).
			WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "feedback_id", "topic", "sentiment_score"}).
				AddRow(3, testDate, testDate, 3, "Topic3", -0.6))

		mock.ExpectQuery(`SELECT \* FROM "analyses" WHERE feedback_id = \$1 ORDER BY "analyses"."id" LIMIT \$2`).
			WithArgs(4, 1).
			WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "feedback_id", "topic", "sentiment_score"}).
				AddRow(4, testDate, testDate, 4, "Topic4", -0.4))

		got := CalculatePercentageSentimentUnderThreshold(feedbacks, -0.5)
		expected := 50.0 // 2 out of 4 feedbacks are under -0.5
		assert.Equal(t, expected, got)

		// Verify that all mock expectations were met
		err := mock.ExpectationsWereMet()
		assert.NoError(t, err)
	})

	t.Run("No feedbacks under threshold", func(t *testing.T) {
		mock := setupTestCalculMetric(t)
		testDate := time.Now()

		feedbacks := []Feedback.Feedback{
			{BaseModel: BaseModel.BaseModel{Id: 1}}, // 0.5 > -0.3
			{BaseModel: BaseModel.BaseModel{Id: 2}}, // -0.2 > -0.3
		}

		// Mock analysis queries
		mock.ExpectQuery(`SELECT \* FROM "analyses" WHERE feedback_id = \$1 ORDER BY "analyses"."id" LIMIT \$2`).
			WithArgs(1, 1).
			WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "feedback_id", "topic", "sentiment_score"}).
				AddRow(1, testDate, testDate, 1, "Topic1", 0.5))

		mock.ExpectQuery(`SELECT \* FROM "analyses" WHERE feedback_id = \$1 ORDER BY "analyses"."id" LIMIT \$2`).
			WithArgs(2, 1).
			WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "feedback_id", "topic", "sentiment_score"}).
				AddRow(2, testDate, testDate, 2, "Topic2", -0.2))

		got := CalculatePercentageSentimentUnderThreshold(feedbacks, -0.3)
		expected := 0.0 // 0 out of 2 feedbacks are under -0.3
		assert.Equal(t, expected, got)

		// Verify that all mock expectations were met
		err := mock.ExpectationsWereMet()
		assert.NoError(t, err)
	})

	t.Run("Database error - continue processing", func(t *testing.T) {
		mock := setupTestCalculMetric(t)
		testDate := time.Now()

		feedbacks := []Feedback.Feedback{
			{BaseModel: BaseModel.BaseModel{Id: 1}}, // Will error
			{BaseModel: BaseModel.BaseModel{Id: 2}}, // Will succeed, under threshold
		}

		// Mock first query to return error
		mock.ExpectQuery(`SELECT \* FROM "analyses" WHERE feedback_id = \$1 ORDER BY "analyses"."id" LIMIT \$2`).
			WithArgs(1, 1).
			WillReturnError(errors.New("database error"))

		// Mock second query to succeed
		mock.ExpectQuery(`SELECT \* FROM "analyses" WHERE feedback_id = \$1 ORDER BY "analyses"."id" LIMIT \$2`).
			WithArgs(2, 1).
			WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "feedback_id", "topic", "sentiment_score"}).
				AddRow(2, testDate, testDate, 2, "Topic2", -0.7))

		got := CalculatePercentageSentimentUnderThreshold(feedbacks, -0.5)
		// Only second feedback is counted (1 under threshold out of 2 total = 50%)
		expected := 50.0
		assert.Equal(t, expected, got)

		// Verify that all mock expectations were met
		err := mock.ExpectationsWereMet()
		assert.NoError(t, err)
	})
}

func TestCalculDistributionByTopic(t *testing.T) {
	t.Run("Success case", func(t *testing.T) {
		mock := setupTestCalculMetric(t)
		testDate := time.Now()

		feedbacks := []Feedback.Feedback{
			{BaseModel: BaseModel.BaseModel{Id: 1}},
			{BaseModel: BaseModel.BaseModel{Id: 2}},
			{BaseModel: BaseModel.BaseModel{Id: 3}},
		}

		// Mock analysis queries
		mock.ExpectQuery(`SELECT \* FROM "analyses" WHERE feedback_id = \$1 ORDER BY "analyses"."id" LIMIT \$2`).
			WithArgs(1, 1).
			WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "feedback_id", "topic", "sentiment_score"}).
				AddRow(1, testDate, testDate, 1, "Support Client", 0.5))

		mock.ExpectQuery(`SELECT \* FROM "analyses" WHERE feedback_id = \$1 ORDER BY "analyses"."id" LIMIT \$2`).
			WithArgs(2, 1).
			WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "feedback_id", "topic", "sentiment_score"}).
				AddRow(2, testDate, testDate, 2, "Support Client", -0.3))

		mock.ExpectQuery(`SELECT \* FROM "analyses" WHERE feedback_id = \$1 ORDER BY "analyses"."id" LIMIT \$2`).
			WithArgs(3, 1).
			WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "feedback_id", "topic", "sentiment_score"}).
				AddRow(3, testDate, testDate, 3, "Performance", 0.8))

		got, err := CalculDistributionByTopic(feedbacks)
		assert.NoError(t, err)

		expected := map[string]float64{
			"Support Client": 66.67, // 2 out of 3
			"Performance":    33.33, // 1 out of 3
		}
		assert.Equal(t, expected, got)

		// Verify that all mock expectations were met
		err = mock.ExpectationsWereMet()
		assert.NoError(t, err)
	})

	t.Run("Database error", func(t *testing.T) {
		mock := setupTestCalculMetric(t)

		feedbacks := []Feedback.Feedback{
			{BaseModel: BaseModel.BaseModel{Id: 1}},
		}

		// Mock analysis query to return error
		mock.ExpectQuery(`SELECT \* FROM "analyses" WHERE feedback_id = \$1 ORDER BY "analyses"."id" LIMIT \$2`).
			WithArgs(1, 1).
			WillReturnError(errors.New("database error"))

		got, err := CalculDistributionByTopic(feedbacks)
		assert.Error(t, err)
		assert.Nil(t, got)
		assert.Contains(t, err.Error(), "database error")

		// Verify that all mock expectations were met
		err = mock.ExpectationsWereMet()
		assert.NoError(t, err)
	})

	t.Run("Empty feedbacks", func(t *testing.T) {
		feedbacks := []Feedback.Feedback{}

		got, err := CalculDistributionByTopic(feedbacks)
		assert.NoError(t, err)
		assert.Empty(t, got)
	})
}
