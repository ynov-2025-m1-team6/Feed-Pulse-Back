package calculmetric

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/models/Feedback"
)

func TestCalculDistributionByChannel(t *testing.T) {
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
			name: "Single channel",
			feedbacks: []Feedback.Feedback{
				{Channel: "email"},
			},
			want: map[string]float64{
				"email": 100,
			},
		},
		{
			name: "Multiple channels with equal distribution",
			feedbacks: []Feedback.Feedback{
				{Channel: "email"},
				{Channel: "web"},
			},
			want: map[string]float64{
				"email": 50,
				"web":   50,
			},
		},
		{
			name: "Multiple channels with unequal distribution",
			feedbacks: []Feedback.Feedback{
				{Channel: "email"},
				{Channel: "email"},
				{Channel: "web"},
				{Channel: "mobile"},
			},
			want: map[string]float64{
				"email":  50,
				"web":    25,
				"mobile": 25,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculDistributionByChannel(tt.feedbacks)
			assert.Equal(t, tt.want, got)
		})
	}
}
