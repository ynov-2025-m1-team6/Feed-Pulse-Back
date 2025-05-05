package Analysis

import (
	"errors"

	"github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/database"
	"github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/models/Analysis"
	"github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/models/Feedback"
	"gorm.io/gorm"
)

// GetAllAnalyses returns all analyses
func GetAllAnalyses() ([]Analysis.CleanAnalysis, error) {
	var analyses []Analysis.Analysis
	result := database.DB.Find(&analyses)
	if result.Error != nil {
		return nil, result.Error
	}
	var cleanAnalyses []Analysis.CleanAnalysis
	for _, analysis := range analyses {
		cleanAnalysis := Analysis.CleanAnalysis{
			BaseModel:      analysis.BaseModel,
			FeedbackID:     analysis.FeedbackID,
			Topic:          analysis.Topic,
			SentimentScore: analysis.SentimentScore,
		}
		cleanAnalyses = append(cleanAnalyses, cleanAnalysis)
	}
	return cleanAnalyses, nil
}

// GetAnalysisByID returns an analysis by its ID
func GetAnalysisByID(id int) (Analysis.CleanAnalysis, error) {
	var analysis Analysis.Analysis
	result := database.DB.Where("id = ?", id).First(&analysis)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return Analysis.CleanAnalysis{}, errors.New("analysis not found")
		}
		return Analysis.CleanAnalysis{}, result.Error
	}
	cleanAnalysis := Analysis.CleanAnalysis{
		BaseModel:      analysis.BaseModel,
		FeedbackID:     analysis.FeedbackID,
		Topic:          analysis.Topic,
		SentimentScore: analysis.SentimentScore,
	}
	return cleanAnalysis, nil
}

// AddAnalysis adds a new analysis
func AddAnalysis(analysis Analysis.Analysis, db *gorm.DB) (Analysis.CleanAnalysis, error) {
	// Check if the referenced feedback exists in the database
	var feedback Feedback.Feedback
	result := db.Where("id = ?", analysis.FeedbackID).First(&feedback)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return Analysis.CleanAnalysis{}, errors.New("feedback not found: the referenced feedback_id does not exist")
		}
		return Analysis.CleanAnalysis{}, result.Error
	}

	// Create the analysis
	result = db.Create(&analysis)
	if result.Error != nil {
		return Analysis.CleanAnalysis{}, result.Error
	}
	cleanAnalysis := Analysis.CleanAnalysis{
		BaseModel:      analysis.BaseModel,
		FeedbackID:     analysis.FeedbackID,
		Topic:          analysis.Topic,
		SentimentScore: analysis.SentimentScore,
	}
	return cleanAnalysis, nil
}

// UpdateAnalysis updates an existing analysis
func UpdateAnalysis(analysis Analysis.Analysis) (Analysis.CleanAnalysis, error) {
	var existingAnalysis Analysis.Analysis
	result := database.DB.Where("feedback_id = ?", analysis.BaseModel.Id).First(&existingAnalysis)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return Analysis.CleanAnalysis{}, errors.New("analysis not found")
		}
		return Analysis.CleanAnalysis{}, result.Error
	}

	result = database.DB.Model(&existingAnalysis).Updates(analysis)
	if result.Error != nil {
		return Analysis.CleanAnalysis{}, result.Error
	}
	cleanAnalysis := Analysis.CleanAnalysis{
		BaseModel:      existingAnalysis.BaseModel,
		FeedbackID:     existingAnalysis.FeedbackID,
		Topic:          existingAnalysis.Topic,
		SentimentScore: existingAnalysis.SentimentScore,
	}
	return cleanAnalysis, nil
}

// DeleteAnalysis deletes an analysis by its ID
func DeleteAnalysis(id int) error {
	var analysis Analysis.Analysis
	result := database.DB.Where("id = ?", id).First(&analysis)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return errors.New("analysis not found")
		}
		return result.Error
	}

	result = database.DB.Delete(&analysis)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// GetAnalysisByFeedbackID returns an analysis by feedback ID
func GetAnalysisByFeedbackID(feedbackID int) (Analysis.CleanAnalysis, error) {
	var analysis Analysis.Analysis
	result := database.DB.Where("feedback_id = ?", feedbackID).First(&analysis)
	if result.Error != nil {
		return Analysis.CleanAnalysis{}, result.Error
	}
	cleanAnalysis := Analysis.CleanAnalysis{
		BaseModel:      analysis.BaseModel,
		FeedbackID:     analysis.FeedbackID,
		Topic:          analysis.Topic,
		SentimentScore: analysis.SentimentScore,
	}
	return cleanAnalysis, nil
}

// GetAnalysesByTopic returns analyses filtered by topic
func GetAnalysesByTopic(topic string) ([]Analysis.CleanAnalysis, error) {
	var analyses []Analysis.Analysis
	result := database.DB.Where("topic = ?", topic).Find(&analyses)
	if result.Error != nil {
		return nil, result.Error
	}
	var cleanAnalyses []Analysis.CleanAnalysis
	for _, analysis := range analyses {
		cleanAnalysis := Analysis.CleanAnalysis{
			BaseModel:      analysis.BaseModel,
			FeedbackID:     analysis.FeedbackID,
			Topic:          analysis.Topic,
			SentimentScore: analysis.SentimentScore,
		}
		cleanAnalyses = append(cleanAnalyses, cleanAnalysis)
	}
	return cleanAnalyses, nil
}

// GetAnalysesBySentimentRange returns analyses with sentiment score in a specific range
func GetAnalysesBySentimentRange(minScore, maxScore float64) ([]Analysis.CleanAnalysis, error) {
	var analyses []Analysis.Analysis
	result := database.DB.Where("sentiment_score BETWEEN ? AND ?", minScore, maxScore).Find(&analyses)
	if result.Error != nil {
		return nil, result.Error
	}
	var cleanAnalyses []Analysis.CleanAnalysis
	for _, analysis := range analyses {
		cleanAnalysis := Analysis.CleanAnalysis{
			BaseModel:      analysis.BaseModel,
			FeedbackID:     analysis.FeedbackID,
			Topic:          analysis.Topic,
			SentimentScore: analysis.SentimentScore,
		}
		cleanAnalyses = append(cleanAnalyses, cleanAnalysis)
	}
	return cleanAnalyses, nil
}
