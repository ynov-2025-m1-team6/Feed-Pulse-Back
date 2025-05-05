package Feedback

import "time"

type FeedbackWithAnalysis struct {
	FeedbackID     int       `json:"feedback_id"`
	Date           time.Time `json:"date"`
	Channel        string    `json:"channel"`
	Text           string    `json:"text"`
	BoardID        int       `json:"board_id"`
	SentimentScore float64   `json:"sentiment_score"`
	Topic          string    `json:"topic"`
}
