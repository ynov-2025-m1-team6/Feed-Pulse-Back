package auth

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/middleware"
	"github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/sessionManager"
)

func TestLogoutHandlerWithMiddleware(t *testing.T) {
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
	app.Get("/api/auth/logout", middleware.AuthRequired(), LogoutHandler)

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
			name:           "Invalid token format",
			authHeader:     "InvalidFormat",
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
			expectedBody:   `{"message":"logged out successfully"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			req := httptest.NewRequest(http.MethodGet, "/api/auth/logout", nil)
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
			
			if trimmedBody := strings.TrimSpace(string(body)); trimmedBody != tt.expectedBody {
				t.Errorf("Expected body %s, got %s", tt.expectedBody, trimmedBody)
			}
			
			// For valid token test, verify the token is invalidated
			if tt.name == "Valid token" {
				valid, _ := sessionManager.Instance.ValidateSession(validToken)
				if valid {
					t.Error("Session should be invalidated after logout")
				}
			}
		})
	}
}
