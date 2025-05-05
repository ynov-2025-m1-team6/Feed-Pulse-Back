package Board

import (
	"errors"
	"github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/models/User"

	"github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/database"
	"github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/models/Board"
	"github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/models/Feedback"
	"gorm.io/gorm"
)

// GetAllBoards returns all boards
func GetAllBoards() ([]Board.Board, error) {
	var boards []Board.Board
	result := database.DB.Find(&boards)
	if result.Error != nil {
		return nil, result.Error
	}
	return boards, nil
}

// GetBoardByID returns a board by its ID
func GetBoardByID(id int) (Board.Board, error) {
	var board Board.Board
	result := database.DB.Where("id = ?", id).First(&board)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return Board.Board{}, errors.New("board not found")
		}
		return Board.Board{}, result.Error
	}
	return board, nil
}

// CreateBoard creates a new board
func CreateBoard(board Board.Board) (Board.Board, error) {
	result := database.DB.Create(&board)
	if result.Error != nil {
		return Board.Board{}, result.Error
	}
	return board, nil
}

// UpdateBoard updates an existing board
func UpdateBoard(board Board.Board) (Board.Board, error) {
	var existingBoard Board.Board
	result := database.DB.Where("id = ?", board.BaseModel.Id).First(&existingBoard)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return Board.Board{}, errors.New("board not found")
		}
		return Board.Board{}, result.Error
	}

	result = database.DB.Model(&existingBoard).Updates(board)
	if result.Error != nil {
		return Board.Board{}, result.Error
	}
	return board, nil
}

// DeleteBoard deletes a board by its ID and its associated feedbacks
func DeleteBoard(id int) error {
	var board Board.Board
	result := database.DB.Where("id = ?", id).First(&board)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return errors.New("board not found")
		}
		return result.Error
	}

	// Delete associated data in a transaction to ensure atomicity
	err := database.DB.Transaction(func(tx *gorm.DB) error {
		// First, get all feedbacks related to this board
		var feedbacks []Feedback.Feedback
		if err := tx.Where("board_id = ?", id).Find(&feedbacks).Error; err != nil {
			return err
		}

		// For each feedback, delete its analyses and then the feedback itself
		for _, feedback := range feedbacks {
			// Delete analyses related to this feedback
			if err := tx.Exec("DELETE FROM analyses WHERE feedback_id = ?", feedback.BaseModel.Id).Error; err != nil {
				return err
			}

			// Delete the feedback
			if err := tx.Delete(&feedback).Error; err != nil {
				return err
			}
		}

		// Remove associations with users (from junction table)
		if err := tx.Exec("DELETE FROM user_boards WHERE board_id = ?", id).Error; err != nil {
			return err
		}

		// Finally delete the board itself
		if err := tx.Delete(&board).Error; err != nil {
			return err
		}

		return nil
	})

	return err
}

// GetBoardByName returns boards that match the given name (exact match)
func GetBoardByName(name string) ([]Board.Board, error) {
	var boards []Board.Board
	result := database.DB.Where("name = ?", name).Find(&boards)
	if result.Error != nil {
		return nil, result.Error
	}
	return boards, nil
}

// GetBoardsWithUsers returns a board with its associated users
func GetBoardsWithUsers(id int) (Board.Board, error) {
	var board Board.Board
	result := database.DB.Preload("Users").Where("id = ?", id).First(&board)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return Board.Board{}, errors.New("board not found")
		}
		return Board.Board{}, result.Error
	}
	return board, nil
}

// GetBoardsWithFeedbacks returns a board with its associated feedbacks
func GetBoardsWithFeedbacks(id int) (Board.Board, error) {
	var board Board.Board
	result := database.DB.Preload("Feedbacks").Where("id = ?", id).First(&board)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return Board.Board{}, errors.New("board not found")
		}
		return Board.Board{}, result.Error
	}
	return board, nil
}

// GetBoardsByUserID returns all boards that a specific user has access to
func GetBoardsByUserID(userId int) ([]Board.Board, error) {
	var boards []Board.Board

	// Using joins to get boards that are associated with the user through the user_boards junction table
	result := database.DB.Joins("JOIN user_boards ON boards.id = user_boards.board_id").
		Where("user_boards.user_id = ?", userId).
		Find(&boards)

	if result.Error != nil {
		return nil, result.Error
	}

	return boards, nil
}

func GetBoardsByUserUUID(userUUID string) ([]Board.Board, error) {
	var boards []Board.Board

	// Using joins to get boards that are associated with the user through the user_boards junction table
	result := database.DB.Joins("JOIN user_boards ON boards.id = user_boards.board_id").
		Joins("JOIN users ON user_boards.user_id = users.id").
		Where("users.uuid = ?", userUUID).
		Find(&boards)

	if result.Error != nil {
		return nil, result.Error
	}

	return boards, nil
}

// AssociateBoardUser associates a user with a board in the junction table
func AssociateBoardUser(boardID int, userID int) error {
	// Check if the board exists
	var board Board.Board
	if err := database.DB.First(&board, boardID).Error; err != nil {
		return err
	}

	// Check if the user exists
	var user User.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		return err
	}

	// Create the association in the junction table
	if err := database.DB.Exec("INSERT INTO user_boards (user_id, board_id) VALUES (?, ?)", userID, boardID).Error; err != nil {
		return err
	}

	return nil
}
