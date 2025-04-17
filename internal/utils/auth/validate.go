package auth

import (
	"fmt"
	"github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/database/user"
	"golang.org/x/crypto/bcrypt"
)

var ErrLoginRequired = fmt.Errorf("login is required")
var ErrPasswordRequired = fmt.Errorf("password is required")
var ErrInvalidCredentials = fmt.Errorf("invalid credentials")
var ErrUserNotFound = fmt.Errorf("user not found")

// ValidateLoginRequest validates the login request payload with a login and a password
func ValidateLoginRequest(login, password string) (string, error) {
	if login == "" {
		return "", ErrLoginRequired
	}
	if password == "" {
		return "", ErrPasswordRequired
	}
	dbUser, err := user.GetUserEitherByEmailOrUsername(login)
	if err != nil {
		return "", ErrUserNotFound
	}
	// bcrypt compare password with the hashed password
	if err := bcrypt.CompareHashAndPassword([]byte(dbUser.Password), []byte(password)); err != nil {
		return "", ErrInvalidCredentials
	}
	return dbUser.UUID, nil
}
