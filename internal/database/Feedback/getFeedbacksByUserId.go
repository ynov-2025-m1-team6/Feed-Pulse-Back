package Feedback

import (
	"errors"

	"github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/database"
	"github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/models/Feedback"
	"github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/models/User"
)

var ErrUserNotFound = errors.New("user not found")
var ErrBoardNotFound = errors.New("no boards found for this user")

// GetFeedbacksWithAnalysesByUserId retrieves feedbacks with their analyses for a given user ID.
func GetFeedbacksWithAnalysesByUserId(userId int, channel string) ([]Feedback.FeedbackWithAnalysis, error) {
	// Check if the user exists
	var userCount int64
	err := database.DB.Model(&User.User{}).Where("id = ?", userId).Count(&userCount).Error
	if err != nil {
		return nil, err
	}
	if userCount == 0 {
		return nil, ErrUserNotFound
	}

	// Retrieves all boards for the user
	var boardsId []int
	err = database.DB.Table("user_boards").
		Select("board_id").
		Where("user_id = ?", userId).
		Pluck("board_id", &boardsId).Error
	if err != nil {
		return nil, err
	}

	// Check if the user has any boards
	if len(boardsId) == 0 {
		return nil, ErrBoardNotFound
	}

	// If a channel is specified, filter the feedbacks by channel
	if channel != "" {
		var feedbacks []Feedback.FeedbackWithAnalysis
		for _, boardId := range boardsId {
			var feedbacksForBoard []Feedback.FeedbackWithAnalysis
			err = database.DB.Table("feedbacks").
				Select("feedbacks.*, analyses.*").
				Joins("LEFT JOIN analyses ON feedbacks.id = analyses.feedback_id").
				Where("feedbacks.board_id = ? AND feedbacks.channel = ?", boardId, channel).
				Scan(&feedbacksForBoard).Error
			if err != nil {
				return nil, err
			}
			feedbacks = append(feedbacks, feedbacksForBoard...)
		}
		return feedbacks, nil
	}

	// If no channel is specified, retrieve all feedbacks for the boards
	var feedbacks []Feedback.FeedbackWithAnalysis
	for _, boardId := range boardsId {
		var feedbacksForBoard []Feedback.FeedbackWithAnalysis
		err = database.DB.Table("feedbacks").
			Select("feedbacks.*, analyses.*").
			Joins("LEFT JOIN analyses ON feedbacks.id = analyses.feedback_id").
			Where("feedbacks.board_id = ?", boardId).
			Scan(&feedbacksForBoard).Error
		if err != nil {
			return nil, err
		}
		feedbacks = append(feedbacks, feedbacksForBoard...)
	}
	return feedbacks, nil
}
