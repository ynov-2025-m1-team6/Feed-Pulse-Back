package auth

import (
	"errors"
	"github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/models/User"
	"golang.org/x/crypto/bcrypt"
	"testing"
)

// Mock implementation of the UserRepository interface
type mockUserRepo struct{}

// GetUserEitherByEmailOrUsername implements the UserRepository interface for testing
func (m *mockUserRepo) GetUserEitherByEmailOrUsername(login string) (*User.User, error) {
	if login == "testuser" || login == "testuser@example.com" {
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
		return &User.User{
			Username: "testuser",
			Email:    "testuser@example.com",
			Password: string(hashedPassword),
			UUID:     "test-uuid-123",
		}, nil
	}
	return nil, errors.New("user not found")
}

func TestValidateLoginRequest(t *testing.T) {
	// Create a mock repository
	mockRepo := &mockUserRepo{}

	tests := []struct {
		name          string
		login         string
		password      string
		expectedUUID  string
		expectedError error
	}{
		{
			name:          "Empty login",
			login:         "",
			password:      "password123",
			expectedUUID:  "",
			expectedError: ErrLoginRequired,
		},
		{
			name:          "Empty password",
			login:         "testuser",
			password:      "",
			expectedUUID:  "",
			expectedError: ErrPasswordRequired,
		},
		{
			name:          "Non-existent user",
			login:         "nonexistent",
			password:      "password123",
			expectedUUID:  "",
			expectedError: ErrUserNotFound,
		},
		{
			name:          "Invalid password",
			login:         "testuser",
			password:      "wrongpassword",
			expectedUUID:  "",
			expectedError: ErrInvalidCredentials,
		},
		{
			name:          "Valid credentials with username",
			login:         "testuser",
			password:      "password123",
			expectedUUID:  "test-uuid-123",
			expectedError: nil,
		},
		{
			name:          "Valid credentials with email",
			login:         "testuser@example.com",
			password:      "password123",
			expectedUUID:  "test-uuid-123",
			expectedError: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uuid, err := ValidateLoginRequestWithRepo(tt.login, tt.password, mockRepo)

			if tt.expectedError != nil {
				if err == nil || err.Error() != tt.expectedError.Error() {
					t.Errorf("ValidateLoginRequest() error = %v, expectedError %v", err, tt.expectedError)
				}
			} else if err != nil {
				t.Errorf("ValidateLoginRequest() unexpected error = %v", err)
			}

			if uuid != tt.expectedUUID {
				t.Errorf("ValidateLoginRequest() uuid = %v, expectedUUID %v", uuid, tt.expectedUUID)
			}
		})
	}
}
