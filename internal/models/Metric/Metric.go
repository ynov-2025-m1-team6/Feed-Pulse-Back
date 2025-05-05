package metric

// Package metric provides the structure for storing and processing metrics related to feedback analysis.
// It includes the distribution of feedback by channel and theme, volumetry by day, average sentiment, and sentiment breakdown.
type Metric struct {
	DistributionByChannel            map[string]float64 `json:"distributionByChannel"`
	DistributionByTopic              map[string]float64 `json:"distributionByTopic"`
	VolumetryByDay                   map[string]float64 `json:"volumetryByDay"`
	AverageSentiment                 float64            `json:"averageSentiment"`
	Sentiment                        Sentiment          `json:"Sentiment"`
	PercentageSentimentUnderTreshold float64            `json:"percentageSentimentUnderTreshold"`
}
