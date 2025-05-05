package Feedback

import (
	"errors"
	"github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/database"
	Analysis2 "github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/database/Analysis"
	"github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/models/Analysis"
	"github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/models/Board"
	"github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/models/Feedback"
	"github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/utils/sentimentAnalysis"
	"gorm.io/gorm"
)

// GetAllFeedbacks returns all feedbacks
func GetAllFeedbacks() ([]Feedback.Feedback, error) {
	var feedbacks []Feedback.Feedback
	result := database.DB.Find(&feedbacks)
	if result.Error != nil {
		return nil, result.Error
	}
	return feedbacks, nil
}

// GetFeedbackByID returns a feedback by its ID
func GetFeedbackByID(id int) (Feedback.Feedback, error) {
	var feedback Feedback.Feedback
	result := database.DB.Where("id = ?", id).First(&feedback)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return Feedback.Feedback{}, errors.New("feedback not found")
		}
		return Feedback.Feedback{}, result.Error
	}
	return feedback, nil
}

// CreateFeedback create a new feedback
func CreateFeedback(feedback Feedback.Feedback, userEmail string) (Feedback.Feedback, error) {
	// Check if the referenced board exists in the database
	var board Board.Board
	result := database.DB.Where("id = ?", feedback.BoardID).First(&board)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return Feedback.Feedback{}, errors.New("board not found: the referenced board_id does not exist")
		}
		return Feedback.Feedback{}, result.Error
	}
	// start a transaction
	err := database.DB.Transaction(func(tx *gorm.DB) error {
		// Create the feedback
		result = tx.Create(&feedback)
		if result.Error != nil {
			return result.Error
		}
		analysis, err := sentimentAnalysis.SentimentAnalysis(feedback, userEmail)
		if err != nil {
			return err
		}
		_, err = Analysis2.AddAnalysis(analysis, tx)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return Feedback.Feedback{}, err
	}

	return feedback, nil
}

// UpdateFeedback update an existing feedback
func UpdateFeedback(feedback Feedback.Feedback) (Feedback.Feedback, error) {
	var existingFeedback Feedback.Feedback
	result := database.DB.Where("id = ?", feedback.BaseModel.Id).First(&existingFeedback)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return Feedback.Feedback{}, errors.New("feedback not found")
		}
		return Feedback.Feedback{}, result.Error
	}

	result = database.DB.Model(&existingFeedback).Updates(feedback)
	if result.Error != nil {
		return Feedback.Feedback{}, result.Error
	}
	return feedback, nil
}

// DeleteFeedback delete a feedback by its ID and its associated analyses
func DeleteFeedback(id int) error {
	var feedback Feedback.Feedback
	result := database.DB.Where("id = ?", id).First(&feedback)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return errors.New("feedback not found")
		}
		return result.Error
	}

	// Delete associated analyses first
	// We use a transaction to ensure atomicity
	err := database.DB.Transaction(func(tx *gorm.DB) error {
		// Delete all analyses related to this feedback
		if err := tx.Where("feedback_id = ?", id).Delete(&Analysis.Analysis{}).Error; err != nil {
			return err
		}

		// Delete the feedback itself
		if err := tx.Delete(&feedback).Error; err != nil {
			return err
		}

		return nil
	})

	return err
}

// GetFeedbacksByChannel returns feedbacks by channel
func GetFeedbacksByChannel(channel string) ([]Feedback.Feedback, error) {
	var feedbacks []Feedback.Feedback
	result := database.DB.Where("channel = ?", channel).Find(&feedbacks)
	if result.Error != nil {
		return nil, result.Error
	}
	return feedbacks, nil
}
