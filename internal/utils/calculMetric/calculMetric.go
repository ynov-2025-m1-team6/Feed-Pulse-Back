package calculmetric

import (
	"math"

	"github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/database/Analysis"
	"github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/models/Feedback"
	metric "github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/models/Metric"
)

// roundToTwo rounds a float64 to 2 decimal places
func roundToTwo(num float64) float64 {
	return math.Round(num*100) / 100
}

func CalculMetric(feedbacks []Feedback.Feedback) (metric.Metric, error) {
	// Initialize the metric struct
	var m = metric.Metric{
		DistributionByChannel: make(map[string]float64),
		DistributionByTopic:   make(map[string]float64),
		VolumetryByDay:        make(map[string]float64),
		Sentiment: metric.Sentiment{
			Positive: 0,
			Neutral:  0,
			Negative: 0,
		},
	}

	// Check if feedbacks is empty
	if len(feedbacks) == 0 {
		return m, nil
	}

	// Calculate the distribution by channel
	m.DistributionByChannel = CalculDistributionByChannel(feedbacks)

	// Calculate the distribution by topic
	distributionByTopic, err := CalculDistributionByTopic(feedbacks)
	if err != nil {
		return m, err
	}
	m.DistributionByTopic = distributionByTopic

	// Calculate the volumetry by day
	m.VolumetryByDay = CalculVolumetryByDay(feedbacks)

	// Calculate the average sentiment
	m.AverageSentiment = CalculAverageSentiment(feedbacks)

	// Calculate the sentiment percentage
	m.Sentiment = CalculateSentimentPercentage(feedbacks)

	// Calculate the percentage of sentiment under threshold
	m.PercentageSentimentUnderTreshold = CalculatePercentageSentimentUnderThreshold(feedbacks, -0.5)

	return m, nil
}

func CalculDistributionByChannel(feedbacks []Feedback.Feedback) map[string]float64 {
	distributionByChannel := make(map[string]float64)

	// Iterate over feedbacks and calculate the distribution by channel
	for _, feedback := range feedbacks {
		channel := feedback.Channel
		distributionByChannel[channel]++
	}

	// Calculate the total number of feedbacks
	totalFeedbacks := float64(len(feedbacks))

	// Calculate the percentage for each channel
	for channel, count := range distributionByChannel {
		distributionByChannel[channel] = roundToTwo((count / totalFeedbacks) * 100)
	}

	return distributionByChannel
}

func CalculDistributionByTopic(feedbacks []Feedback.Feedback) (map[string]float64, error) {
	distributionByTheme := make(map[string]float64)

	// Iterate over feedbacks and calculate the distribution by theme
	for _, feedback := range feedbacks {
		analysis, err := Analysis.GetAnalysisByFeedbackID(feedback.Id)
		if err != nil {
			return nil, err
		}
		Topic := analysis.Topic
		distributionByTheme[Topic]++
	}

	// Calculate the total number of feedbacks
	totalFeedbacks := float64(len(feedbacks))

	// Calculate the percentage for each theme
	for theme, count := range distributionByTheme {
		distributionByTheme[theme] = roundToTwo((count / totalFeedbacks) * 100)
	}

	return distributionByTheme, nil
}

func CalculVolumetryByDay(feedbacks []Feedback.Feedback) map[string]float64 {
	volumetryByDay := make(map[string]float64)

	// Iterate over feedbacks and calculate the volumetry by day
	for _, feedback := range feedbacks {
		date := feedback.CreatedAt.Format("2006-01-02")
		volumetryByDay[date]++
	}

	// Calculate the total number of feedbacks
	totalFeedbacks := float64(len(feedbacks))

	// Calculate the percentage for each day
	for date, count := range volumetryByDay {
		volumetryByDay[date] = roundToTwo((count / totalFeedbacks) * 100)
	}

	return volumetryByDay
}

func CalculAverageSentiment(feedbacks []Feedback.Feedback) float64 {
	// Initialize the total sentiment score
	var totalSentiment float64

	// Iterate over feedbacks and calculate the total sentiment score
	for _, feedback := range feedbacks {
		analysis, err := Analysis.GetAnalysisByFeedbackID(feedback.Id)
		if err != nil {
			return -100 // Handle error appropriately
		}
		totalSentiment += analysis.SentimentScore
	}

	// Calculate the average sentiment score
	averageSentiment := totalSentiment / float64(len(feedbacks))

	return roundToTwo(averageSentiment)
}

func CalculateSentimentPercentage(feedbacks []Feedback.Feedback) metric.Sentiment {
	// Initialize the sentiment counts
	var positiveCount, neutralCount, negativeCount int

	// Iterate over feedbacks and count the sentiment types
	for _, feedback := range feedbacks {
		analysis, err := Analysis.GetAnalysisByFeedbackID(feedback.Id)
		if err != nil {
			continue // Handle error appropriately
		}
		if analysis.SentimentScore > 0.25 {
			positiveCount++
		} else if analysis.SentimentScore < -0.25 {
			negativeCount++
		} else {
			neutralCount++
		}
	}

	totalFeedbacks := float64(len(feedbacks))

	return metric.Sentiment{
		Positive: roundToTwo((float64(positiveCount) / totalFeedbacks) * 100),
		Neutral:  roundToTwo((float64(neutralCount) / totalFeedbacks) * 100),
		Negative: roundToTwo((float64(negativeCount) / totalFeedbacks) * 100),
	}
}

func CalculatePercentageSentimentUnderThreshold(feedbacks []Feedback.Feedback, threshold float64) float64 {
	// Initialize the count of feedbacks under the threshold
	var countUnderThreshold int

	// Iterate over feedbacks and count the ones under the threshold
	for _, feedback := range feedbacks {
		analysis, err := Analysis.GetAnalysisByFeedbackID(feedback.Id)
		if err != nil {
			continue // Handle error appropriately
		}
		if analysis.SentimentScore < threshold {
			countUnderThreshold++
		}
	}

	totalFeedbacks := float64(len(feedbacks))

	return roundToTwo((float64(countUnderThreshold) / totalFeedbacks) * 100)
}
