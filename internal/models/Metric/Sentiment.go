package metric

type Sentiment struct {
	Positive float64 `json:"positive"`
	Neutral  float64 `json:"neutral"`
	Negative float64 `json:"negative"`
}
