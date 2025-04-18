package auth

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/models/User"
	"github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/sessionManager"
	"golang.org/x/crypto/bcrypt"
)

// MockLoginRepo implements the UserRepository interface for testing the login handler
type MockLoginRepo struct{}

// GetUserEitherByEmailOrUsername implements the UserRepository interface for testing
func (m *MockLoginRepo) GetUserEitherByEmailOrUsername(login string) (*User.User, error) {
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

func TestLoginHandler(t *testing.T) {
	// Initialize session manager for testing
	sessionManager.InitSessionManager("test-secret-key", time.Hour)
	// Setup test app with mock repository
	mockRepo := &MockLoginRepo{}
	app := fiber.New()
	app.Post("/api/auth/login", func(c *fiber.Ctx) error {
		return LoginHandlerWithRepo(c, mockRepo)
	})

	tests := []struct {
		name           string
		requestBody    map[string]string
		expectedStatus int
		expectedBody   string
		checkHeader    bool
	}{
		{
			name: "Empty login",
			requestBody: map[string]string{
				"login":    "",
				"password": "password123",
			},
			expectedStatus: fiber.StatusBadRequest,
			expectedBody:   `{"error":"login is required"}`,
		},
		{
			name: "Empty password",
			requestBody: map[string]string{
				"login":    "testuser",
				"password": "",
			},
			expectedStatus: fiber.StatusBadRequest,
			expectedBody:   `{"error":"password is required"}`,
		},
		{
			name: "User not found",
			requestBody: map[string]string{
				"login":    "nonexistent",
				"password": "password123",
			},
			expectedStatus: fiber.StatusNotFound,
			expectedBody:   `{"error":"user not found"}`,
		},
		{
			name: "Invalid credentials",
			requestBody: map[string]string{
				"login":    "testuser",
				"password": "wrongpassword",
			},
			expectedStatus: fiber.StatusUnauthorized,
			expectedBody:   `{"error":"invalid credentials"}`,
		},
		{
			name: "Successful login",
			requestBody: map[string]string{
				"login":    "testuser",
				"password": "password123",
			},
			expectedStatus: fiber.StatusOK,
			expectedBody:   `{"message":"login successful"}`,
			checkHeader:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Convert request body to JSON
			jsonBody, _ := json.Marshal(tt.requestBody)

			// Create request
			req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(jsonBody))
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

			// If this is a successful login, check for the Authorization header
			if tt.checkHeader {
				authHeader := resp.Header.Get("Authorization")
				if !strings.HasPrefix(authHeader, "Bearer ") {
					t.Errorf("Expected Authorization header with Bearer token, got %s", authHeader)
				}

				// The token should be a non-empty string
				token := strings.TrimPrefix(authHeader, "Bearer ")
				if token == "" {
					t.Error("Expected non-empty token in Authorization header")
				}
			}
		})
	}
}
