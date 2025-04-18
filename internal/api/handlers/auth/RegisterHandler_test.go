package auth

import (
	"bytes"
	"encoding/json"
	"github.com/gofiber/fiber/v2"
	"github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/models/User"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// MockUserRepository implements the UserRepository interface for testing
type MockUserRepository struct{}

// GetUserByUsername implements the UserRepository interface for testing
func (m *MockUserRepository) GetUserByUsername(username string) (*User.User, error) {
	if username == "existinguser" {
		return &User.User{
			Username: "existinguser",
			Email:    "existing@example.com",
		}, nil
	}
	return nil, fiber.NewError(fiber.StatusNotFound, "user not found")
}

// GetUserByEmail implements the UserRepository interface for testing
func (m *MockUserRepository) GetUserByEmail(email string) (*User.User, error) {
	if email == "existing@example.com" {
		return &User.User{
			Username: "existinguser",
			Email:    "existing@example.com",
		}, nil
	}
	return nil, fiber.NewError(fiber.StatusNotFound, "user not found")
}

// CreateUser implements the UserRepository interface for testing
func (m *MockUserRepository) CreateUser(user *User.User) error {
	// Simple mock just returns success
	return nil
}

func TestRegisterHandler(t *testing.T) {
	mockRepo := &MockUserRepository{}
	
	app := fiber.New()
	app.Post("/api/auth/register", func(c *fiber.Ctx) error {
		return RegisterHandlerWithRepo(c, mockRepo)
	})

	tests := []struct {
		name           string
		requestBody    map[string]string
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "Valid registration",
			requestBody: map[string]string{
				"username": "newuser",
				"email":    "newuser@example.com",
				"password": "password123",
			},
			expectedStatus: fiber.StatusCreated,
			expectedBody:   `{"message":"user registered successfully"}`,
		},
		{
			name: "Missing username",
			requestBody: map[string]string{
				"email":    "test@example.com",
				"password": "password123",
			},
			expectedStatus: fiber.StatusBadRequest,
			expectedBody:   `{"error":"username, email, and password are required"}`,
		},
		{
			name: "Missing email",
			requestBody: map[string]string{
				"username": "testuser",
				"password": "password123",
			},
			expectedStatus: fiber.StatusBadRequest,
			expectedBody:   `{"error":"username, email, and password are required"}`,
		},
		{
			name: "Missing password",
			requestBody: map[string]string{
				"username": "testuser",
				"email":    "test@example.com",
			},
			expectedStatus: fiber.StatusBadRequest,
			expectedBody:   `{"error":"username, email, and password are required"}`,
		},
		{
			name: "Password too short",
			requestBody: map[string]string{
				"username": "testuser",
				"email":    "test@example.com",
				"password": "short",
			},
			expectedStatus: fiber.StatusBadRequest,
			expectedBody:   `{"error":"password must be at least 8 characters long"}`,
		},
		{
			name: "Password too long",
			requestBody: map[string]string{
				"username": "testuser",
				"email":    "test@example.com",
				"password": strings.Repeat("a", 51),
			},
			expectedStatus: fiber.StatusBadRequest,
			expectedBody:   `{"error":"password must be at most 50 characters long"}`,
		},
		{
			name: "Invalid email format",
			requestBody: map[string]string{
				"username": "testuser",
				"email":    "invalid-email",
				"password": "password123",
			},
			expectedStatus: fiber.StatusBadRequest,
			expectedBody:   `{"error":"invalid email format"}`,
		},
		{
			name: "Username already exists",
			requestBody: map[string]string{
				"username": "existinguser",
				"email":    "new@example.com",
				"password": "password123",
			},
			expectedStatus: fiber.StatusConflict,
			expectedBody:   `{"error":"username already exists"}`,
		},
		{
			name: "Email already exists",
			requestBody: map[string]string{
				"username": "brandnewuser",
				"email":    "existing@example.com",
				"password": "password123",
			},
			expectedStatus: fiber.StatusConflict,
			expectedBody:   `{"error":"email already exists"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Convert request body to JSON
			jsonBody, _ := json.Marshal(tt.requestBody)
			
			// Create request
			req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader(jsonBody))
			req.Header.Set("Content-Type", "application/json")
			
			// Execute request
			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("Failed to test request: %v", err)
			}
			
			// Check status code
			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("Expected status code %d, got %d", tt.expectedStatus, resp.StatusCode)
			}
			
			// Check response body
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("Failed to read response body: %v", err)
			}
			
			if trimmedBody := strings.TrimSpace(string(body)); trimmedBody != tt.expectedBody {
				t.Errorf("Expected body %s, got %s", tt.expectedBody, trimmedBody)
			}
		})
	}
}
