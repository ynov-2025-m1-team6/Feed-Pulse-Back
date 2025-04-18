package Analysis

import (
	"github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/models/BaseModel"
	"github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/models/Feedback"
)

type Analysis struct {
	BaseModel.BaseModel
	SentimentScore float64 `json:"sentiment_score" gorm:"not null"`
	Topic          string  `json:"topic" gorm:"not null"`
	FeedbackID     int     `json:"feedback_id" gorm:"not null;unique"`

	// Associate Feedback
	Feedback Feedback.Feedback
}
