package user

import (
	"github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/database"
	"github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/database/Board"
	BoardModel "github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/models/Board"
	"github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/models/User"
)

// CreateUser inserts a new user into the database
func CreateUser(user *User.User) error {
	err := database.DB.Create(user).Error
	if err != nil {
		return err
	}
	boardObj := BoardModel.Board{
		Name: user.Username + " Board",
	}
	board, err := Board.CreateBoard(boardObj)
	if err != nil {
		return err
	}

	return Board.AssociateBoardUser(board.Id, user.Id)
}

// GetUserByUsername retrieves a user by their username
func GetUserByUsername(username string) (*User.User, error) {
	var user User.User
	if err := database.DB.Where("username = ?", username).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserByID retrieves a user by their ID
func GetUserByID(id uint) (*User.User, error) {
	var user User.User
	if err := database.DB.First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserByUUID retrieves a user by their UUID
func GetUserByUUID(uuid string) (*User.User, error) {
	var user User.User
	if err := database.DB.Where("uuid = ?", uuid).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserByEmail retrieves a user by their email
func GetUserByEmail(email string) (*User.User, error) {
	var user User.User
	if err := database.DB.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func GetUserEitherByEmailOrUsername(login string) (*User.User, error) {
	var user User.User
	if err := database.DB.Where("email = ? OR username = ?", login, login).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// GetAllUsers retrieves all users
func GetAllUsers() ([]User.User, error) {
	var users []User.User
	if err := database.DB.Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

// UpdateUser updates an existing user's details
func UpdateUser(user *User.User) error {
	return database.DB.Save(user).Error
}

// DeleteUser deletes a user by their ID
func DeleteUser(id uint) error {
	if err := database.DB.Delete(&User.User{}, id).Error; err != nil {
		return err
	}
	return nil
}
