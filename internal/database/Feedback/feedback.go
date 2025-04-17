package Feedback

import (
	"errors"
	"github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/database"
	"github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/models/Feedback"
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
func CreateFeedback(feedback Feedback.Feedback) (Feedback.Feedback, error) {
	result := database.DB.Create(&feedback)
	if result.Error != nil {
		return Feedback.Feedback{}, result.Error
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

// DeleteFeedback delete a feedback by its ID
func DeleteFeedback(id int) error {
	var feedback Feedback.Feedback
	result := database.DB.Where("id = ?", id).First(&feedback)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return errors.New("feedback not found")
		}
		return result.Error
	}

	result = database.DB.Delete(&feedback)
	if result.Error != nil {
		return result.Error
	}
	return nil
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
