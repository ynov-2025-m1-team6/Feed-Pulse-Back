package auth

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/middleware"
	"github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/models/User"
	"github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/sessionManager"
)

// MockUserInfoRepo implements the UserRepository interface for testing
type MockUserInfoRepo struct{}

func (m *MockUserInfoRepo) GetUserByUUID(uuid string) (*User.User, error) {
	if uuid == "test-user-uuid" {
		return &User.User{
			UUID:     "test-user-uuid",
			Username: "testuser",
			Email:    "testuser@example.com",
			Password: "hashed-password",
		}, nil
	}
	return nil, errors.New("user not found")
}

func (m *MockUserInfoRepo) GetUserByUsername(username string) (*User.User, error) {
	return nil, errors.New("not implemented")
}

func (m *MockUserInfoRepo) GetUserByEmail(email string) (*User.User, error) {
	return nil, errors.New("not implemented")
}

func (m *MockUserInfoRepo) CreateUser(user *User.User) error {
	return errors.New("not implemented")
}

func TestUserInfoHandler(t *testing.T) {
	// Initialize session manager for testing
	sessionManager.InitSessionManager("test-secret-key", time.Hour)

	// Create a valid token for testing
	userUUID := "test-user-uuid"
	validToken, err := sessionManager.Instance.CreateSession(userUUID)
	if err != nil {
		t.Fatalf("Failed to create test session: %v", err)
	}

	// Setup test app with middleware
	app := fiber.New()
	app.Get("/api/auth/user", middleware.AuthRequired(), func(c *fiber.Ctx) error {
		return UserInfoHandlerWithRepo(c, &MockUserInfoRepo{})
	})

	tests := []struct {
		name           string
		authHeader     string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "No token provided",
			authHeader:     "",
			expectedStatus: fiber.StatusUnauthorized,
			expectedBody:   `{"error":"unauthorized: Missing or invalid authorization header"}`,
		},
		{
			name:           "Invalid token",
			authHeader:     "Bearer invalid-token",
			expectedStatus: fiber.StatusUnauthorized,
			expectedBody:   `{"error":"unauthorized: Invalid or expired token"}`,
		},
		{
			name:           "Valid token",
			authHeader:     "Bearer " + validToken,
			expectedStatus: fiber.StatusOK,
			expectedBody:   `{"uuid":"test-user-uuid","username":"testuser","email":"testuser@example.com","password":""}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			req := httptest.NewRequest(http.MethodGet, "/api/auth/user", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

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

			// For valid token test, compare JSON objects
			if tt.name == "Valid token" {
				var expectedUser, actualUser User.User
				if err := json.Unmarshal([]byte(tt.expectedBody), &expectedUser); err != nil {
					t.Fatalf("Failed to unmarshal expected body: %v", err)
				}
				if err := json.Unmarshal(body, &actualUser); err != nil {
					t.Fatalf("Failed to unmarshal actual body: %v", err)
				}
				if expectedUser.UUID != actualUser.UUID ||
					expectedUser.Username != actualUser.Username ||
					expectedUser.Email != actualUser.Email ||
					actualUser.Password != "" {
					t.Errorf("Response body mismatch. Expected %+v, got %+v", expectedUser, actualUser)
				}
			} else {
				if trimmedBody := strings.TrimSpace(string(body)); trimmedBody != tt.expectedBody {
					t.Errorf("Expected body %s, got %s", tt.expectedBody, trimmedBody)
				}
			}
		})
	}
}

func TestUserInfoHandler_Unauthorized(t *testing.T) {
	// Test case where user UUID is not found in context
	// This simulates a scenario where the middleware didn't set the user UUID properly

	app := fiber.New()

	// Create a direct route without authentication middleware
	// This will simulate the case where context doesn't have user UUID
	app.Get("/api/auth/user-direct", func(c *fiber.Ctx) error {
		return UserInfoHandlerWithRepo(c, &MockUserInfoRepo{})
	})

	// Create request without going through auth middleware
	req := httptest.NewRequest(http.MethodGet, "/api/auth/user-direct", nil)

	// Execute request
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to test request: %v", err)
	}

	// Check status code
	expectedStatus := fiber.StatusUnauthorized
	if resp.StatusCode != expectedStatus {
		t.Errorf("Expected status code %d, got %d", expectedStatus, resp.StatusCode)
	}

	// Check response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	expectedBody := `{"error":"unauthorized: user not found in context"}`
	if trimmedBody := strings.TrimSpace(string(body)); trimmedBody != expectedBody {
		t.Errorf("Expected body %s, got %s", expectedBody, trimmedBody)
	}
}

func TestUserInfoHandler_UserNotFoundInDB(t *testing.T) {
	// Initialize session manager for testing
	sessionManager.InitSessionManager("test-secret-key", time.Hour)

	// Create a valid token for a non-existent user
	nonExistentUUID := "non-existent-uuid"
	validToken, err := sessionManager.Instance.CreateSession(nonExistentUUID)
	if err != nil {
		t.Fatalf("Failed to create test session: %v", err)
	}

	// Setup test app with middleware
	app := fiber.New()
	app.Get("/api/auth/user", middleware.AuthRequired(), func(c *fiber.Ctx) error {
		return UserInfoHandlerWithRepo(c, &MockUserInfoRepo{})
	})

	// Create request with valid token but for non-existent user
	req := httptest.NewRequest(http.MethodGet, "/api/auth/user", nil)
	req.Header.Set("Authorization", "Bearer "+validToken)

	// Execute request
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to test request: %v", err)
	}

	// Check status code
	expectedStatus := fiber.StatusNotFound
	if resp.StatusCode != expectedStatus {
		t.Errorf("Expected status code %d, got %d", expectedStatus, resp.StatusCode)
	}

	// Check response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	expectedBody := `{"error":"user not found"}`
	if trimmedBody := strings.TrimSpace(string(body)); trimmedBody != expectedBody {
		t.Errorf("Expected body %s, got %s", expectedBody, trimmedBody)
	}
}

func TestUserInfoHandler_WithDefaultRepo(t *testing.T) {
	// This test verifies that UserInfoHandler correctly delegates to UserInfoHandlerWithRepo
	// Since UserInfoHandler uses the defaultUserRepo which requires a real database connection,
	// we'll test by mocking the defaultUserRepo temporarily

	// Save the original defaultUserRepo
	originalRepo := defaultUserRepo
	defer func() {
		// Restore the original repo after test
		defaultUserRepo = originalRepo
	}()

	// Replace defaultUserRepo with our mock for this test
	defaultUserRepo = &MockUserInfoRepo{}

	// Initialize session manager for testing
	sessionManager.InitSessionManager("test-secret-key", time.Hour)

	// Create a valid token for testing
	userUUID := "test-user-uuid"
	validToken, err := sessionManager.Instance.CreateSession(userUUID)
	if err != nil {
		t.Fatalf("Failed to create test session: %v", err)
	}

	// Setup test app with middleware and the actual UserInfoHandler function
	app := fiber.New()
	app.Get("/api/auth/user", middleware.AuthRequired(), UserInfoHandler)

	// Create request with valid token
	req := httptest.NewRequest(http.MethodGet, "/api/auth/user", nil)
	req.Header.Set("Authorization", "Bearer "+validToken)

	// Execute request
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to test request: %v", err)
	}

	// Check status code
	expectedStatus := fiber.StatusOK
	if resp.StatusCode != expectedStatus {
		t.Errorf("Expected status code %d, got %d", expectedStatus, resp.StatusCode)
	}

	// Check response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	// Verify that we get the expected user data from our mock
	var actualUser User.User
	if err := json.Unmarshal(body, &actualUser); err != nil {
		t.Fatalf("Failed to unmarshal actual body: %v", err)
	}

	expectedUser := User.User{
		UUID:     "test-user-uuid",
		Username: "testuser",
		Email:    "testuser@example.com",
		Password: "", // Password should be cleared in response
	}

	if actualUser.UUID != expectedUser.UUID ||
		actualUser.Username != expectedUser.Username ||
		actualUser.Email != expectedUser.Email ||
		actualUser.Password != "" {
		t.Errorf("Response body mismatch. Expected %+v, got %+v", expectedUser, actualUser)
	}
}
