package auth

import (
	"fmt"
	"github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/database/user"
	"github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/models/User"
	"golang.org/x/crypto/bcrypt"
)

var ErrLoginRequired = fmt.Errorf("login is required")
var ErrPasswordRequired = fmt.Errorf("password is required")
var ErrInvalidCredentials = fmt.Errorf("invalid credentials")
var ErrUserNotFound = fmt.Errorf("user not found")

// UserRepository defines an interface for user database operations
type UserRepository interface {
	GetUserEitherByEmailOrUsername(login string) (*User.User, error)
}

// DefaultUserRepository is the default implementation that uses the actual database
type DefaultUserRepository struct{}

// GetUserEitherByEmailOrUsername implements the UserRepository interface
func (r *DefaultUserRepository) GetUserEitherByEmailOrUsername(login string) (*User.User, error) {
	return user.GetUserEitherByEmailOrUsername(login)
}

// defaultRepo is the default repository used by ValidateLoginRequest
var defaultRepo UserRepository = &DefaultUserRepository{}

// ValidateLoginRequest validates the login request payload with a login and a password
func ValidateLoginRequest(login, password string) (string, error) {
	return ValidateLoginRequestWithRepo(login, password, defaultRepo)
}

// ValidateLoginRequestWithRepo validates the login request using a specific repository
func ValidateLoginRequestWithRepo(login, password string, repo UserRepository) (string, error) {
	if login == "" {
		return "", ErrLoginRequired
	}
	if password == "" {
		return "", ErrPasswordRequired
	}
	dbUser, err := repo.GetUserEitherByEmailOrUsername(login)
	if err != nil {
		return "", ErrUserNotFound
	}
	// bcrypt compare password with the hashed password
	if err := bcrypt.CompareHashAndPassword([]byte(dbUser.Password), []byte(password)); err != nil {
		return "", ErrInvalidCredentials
	}
	return dbUser.UUID, nil
}
