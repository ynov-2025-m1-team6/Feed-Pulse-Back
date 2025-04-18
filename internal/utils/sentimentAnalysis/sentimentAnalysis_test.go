package sentimentAnalysis

import (
	"github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/models/Feedback"
	"log"
	"testing"
)

func Test_SentimentAnalysis(t *testing.T) {
	feedback := Feedback.Feedback{
		Text: "La performance est excellente, aucun ralentissement.",
	}
	analysis, err := SentimentAnalysis(feedback)
	if err != nil {
		log.Fatalf("Error analyzing sentiment: %v", err)
	}
	log.Printf("Sentiment analysis result: %+v\n", analysis)
}
