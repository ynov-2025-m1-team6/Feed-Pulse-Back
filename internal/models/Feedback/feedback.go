package Feedback

import (
	"github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/models/BaseModel"
	"time"
)

type Feedback struct {
	BaseModel.BaseModel
	FeedbackID string    `json:"feedback_id" gorm:"not null;unique"`
	Date       time.Time `json:"date" gorm:"not null"`
	Channel    string    `json:"channel" gorm:"not null"`
	Text       string    `json:"text" gorm:"not null"`
}
