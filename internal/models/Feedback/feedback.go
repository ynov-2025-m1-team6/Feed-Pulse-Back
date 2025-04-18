package Feedback

import (
	"time"

	"github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/models/BaseModel"
)

type Feedback struct {
	BaseModel.BaseModel
	Date    time.Time `json:"date" gorm:"not null"`
	Channel string    `json:"channel" gorm:"not null"`
	Text    string    `json:"text" gorm:"not null"`

	BoardID int `json:"board_id" gorm:"not null"`
}

type FeedbackJson struct {
	Date    time.Time `json:"date"`
	Channel string    `json:"channel"`
	Text    string    `json:"text"`
}
