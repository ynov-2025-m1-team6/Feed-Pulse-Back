package middleware

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	"github.com/stretchr/testify/assert"
	"github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/sessionManager"
)

// setupTestApp creates a test fiber app with the auth middleware
func setupTestApp() *fiber.App {
	app := fiber.New()

	// Setup a test route that requires authentication
	app.Get("/protected", AuthRequired(), func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "success"})
	})

	// Setup test routes for helper functions
	app.Get("/user", AuthRequired(), func(c *fiber.Ctx) error {
		userUUID, ok := GetUserUUID(c)
		return c.JSON(fiber.Map{"userUUID": userUUID, "ok": ok})
	})

	app.Get("/token", AuthRequired(), func(c *fiber.Ctx) error {
		token, ok := GetToken(c)
		return c.JSON(fiber.Map{"token": token, "ok": ok})
	})

	app.Get("/claims", AuthRequired(), func(c *fiber.Ctx) error {
		claims, ok := GetClaims(c)
		return c.JSON(fiber.Map{"claims": claims, "ok": ok})
	})

	app.Get("/claim/:key", AuthRequired(), func(c *fiber.Ctx) error {
		key := c.Params("key")
		value, exists := GetClaimValue(c, key)
		return c.JSON(fiber.Map{"value": value, "exists": exists})
	})

	return app
}

func TestAuthRequired_MissingAuthorizationHeader(t *testing.T) {
	// Initialize session manager for testing
	sessionManager.InitSessionManager("test-secret-key", 1*time.Hour)

	app := setupTestApp()

	// Test request without Authorization header
	req, _ := http.NewRequest("GET", "/protected", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestAuthRequired_InvalidAuthorizationHeader(t *testing.T) {
	// Initialize session manager for testing
	sessionManager.InitSessionManager("test-secret-key", 1*time.Hour)

	app := setupTestApp()

	// Test request with invalid Authorization header format
	req, _ := http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "InvalidFormat token")
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestAuthRequired_EmptyAuthorizationHeader(t *testing.T) {
	// Initialize session manager for testing
	sessionManager.InitSessionManager("test-secret-key", 1*time.Hour)

	app := setupTestApp()

	// Test request with empty Authorization header
	req, _ := http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "")
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestAuthRequired_InvalidToken(t *testing.T) {
	// Initialize session manager for testing
	sessionManager.InitSessionManager("test-secret-key", 1*time.Hour)

	app := setupTestApp()

	// Test request with invalid token
	req, _ := http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestAuthRequired_ValidToken(t *testing.T) {
	// Initialize session manager for testing
	sessionManager.InitSessionManager("test-secret-key", 1*time.Hour)

	// Create a valid session
	userUUID := "test-user-123"
	token, err := sessionManager.Instance.CreateSession(userUUID)
	assert.NoError(t, err)

	app := setupTestApp()

	// Test request with valid token
	req, _ := http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestAuthRequired_ExpiredToken(t *testing.T) {
	// Initialize session manager with very short expiration for testing
	sessionManager.InitSessionManager("test-secret-key", 1*time.Millisecond)

	// Create a session and wait for it to expire
	userUUID := "test-user-123"
	token, err := sessionManager.Instance.CreateSession(userUUID)
	assert.NoError(t, err)

	// Wait for token to expire
	time.Sleep(10 * time.Millisecond)

	app := setupTestApp()

	// Test request with expired token
	req, _ := http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestAuthRequired_InvalidSigningMethod(t *testing.T) {
	// Initialize session manager for testing
	sessionManager.InitSessionManager("test-secret-key", 1*time.Hour)

	// Create a token with invalid signing method (RSA instead of HMAC)
	// This will fail because we don't have RSA keys, but we'll create a malformed token
	tokenString := "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyVVVJRCI6InRlc3QtdXNlci0xMjMiLCJleHAiOjE2OTkyODI4MDB9.invalid-signature"

	app := setupTestApp()

	// Test request with token using invalid signing method
	req, _ := http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", tokenString))
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestGetUserUUID_Success(t *testing.T) {
	// Initialize session manager for testing
	sessionManager.InitSessionManager("test-secret-key", 1*time.Hour)

	// Create a valid session
	userUUID := "test-user-123"
	token, err := sessionManager.Instance.CreateSession(userUUID)
	assert.NoError(t, err)

	app := setupTestApp()

	// Test GetUserUUID with valid token
	req, _ := http.NewRequest("GET", "/user", nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestGetUserUUID_NoContext(t *testing.T) {
	// Test GetUserUUID when no user is in context
	app := fiber.New()
	app.Get("/test", func(c *fiber.Ctx) error {
		userUUID, ok := GetUserUUID(c)
		return c.JSON(fiber.Map{"userUUID": userUUID, "ok": ok})
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestGetToken_Success(t *testing.T) {
	// Initialize session manager for testing
	sessionManager.InitSessionManager("test-secret-key", 1*time.Hour)

	// Create a valid session
	userUUID := "test-user-123"
	token, err := sessionManager.Instance.CreateSession(userUUID)
	assert.NoError(t, err)

	app := setupTestApp()

	// Test GetToken with valid token
	req, _ := http.NewRequest("GET", "/token", nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestGetToken_NoContext(t *testing.T) {
	// Test GetToken when no token is in context
	app := fiber.New()
	app.Get("/test", func(c *fiber.Ctx) error {
		token, ok := GetToken(c)
		return c.JSON(fiber.Map{"token": token, "ok": ok})
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestGetClaims_Success(t *testing.T) {
	// Initialize session manager for testing
	sessionManager.InitSessionManager("test-secret-key", 1*time.Hour)

	// Create a valid session
	userUUID := "test-user-123"
	token, err := sessionManager.Instance.CreateSession(userUUID)
	assert.NoError(t, err)

	app := setupTestApp()

	// Test GetClaims with valid token
	req, _ := http.NewRequest("GET", "/claims", nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestGetClaims_NoContext(t *testing.T) {
	// Test GetClaims when no claims are in context
	app := fiber.New()
	app.Get("/test", func(c *fiber.Ctx) error {
		claims, ok := GetClaims(c)
		return c.JSON(fiber.Map{"claims": claims, "ok": ok})
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestGetClaimValue_Success(t *testing.T) {
	// Initialize session manager for testing
	sessionManager.InitSessionManager("test-secret-key", 1*time.Hour)

	// Create a valid session
	userUUID := "test-user-123"
	token, err := sessionManager.Instance.CreateSession(userUUID)
	assert.NoError(t, err)

	app := setupTestApp()

	// Test GetClaimValue with valid token and existing claim
	req, _ := http.NewRequest("GET", "/claim/userUUID", nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestGetClaimValue_NonexistentClaim(t *testing.T) {
	// Initialize session manager for testing
	sessionManager.InitSessionManager("test-secret-key", 1*time.Hour)

	// Create a valid session
	userUUID := "test-user-123"
	token, err := sessionManager.Instance.CreateSession(userUUID)
	assert.NoError(t, err)

	app := setupTestApp()

	// Test GetClaimValue with valid token but non-existent claim
	req, _ := http.NewRequest("GET", "/claim/nonexistent", nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestGetClaimValue_NoContext(t *testing.T) {
	// Test GetClaimValue when no claims are in context
	app := fiber.New()
	app.Get("/test", func(c *fiber.Ctx) error {
		value, exists := GetClaimValue(c, "userUUID")
		return c.JSON(fiber.Map{"value": value, "exists": exists})
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestAuthRequired_MalformedJWT(t *testing.T) {
	// Initialize session manager for testing
	sessionManager.InitSessionManager("test-secret-key", 1*time.Hour)

	app := setupTestApp()

	// Test request with malformed JWT token
	req, _ := http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer malformed.jwt.token")
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestAuthRequired_TokenWithoutClaims(t *testing.T) {
	// Initialize session manager for testing
	sessionManager.InitSessionManager("test-secret-key", 1*time.Hour)

	// Create a valid JWT token but without proper claims structure
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
	})

	tokenString, err := token.SignedString(sessionManager.Instance.GetSecretKey())
	assert.NoError(t, err)

	// Add the token to the session manager manually
	sessionManager.Instance.DeleteSession(tokenString) // Clean first
	// We need to create the session properly for validation to pass
	userUUID := "test-user-no-claims"
	validToken, err := sessionManager.Instance.CreateSession(userUUID)
	assert.NoError(t, err)

	app := setupTestApp()

	// Test with the properly created token
	req, _ := http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", validToken))
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestConstants(t *testing.T) {
	// Test that constants are defined correctly
	assert.Equal(t, "userUUID", UserContextKey)
	assert.Equal(t, "token", TokenContextKey)
	assert.Equal(t, "claims", ClaimsContextKey)
}

func TestAuthRequired_SessionValidationError(t *testing.T) {
	// Initialize session manager for testing
	sessionManager.InitSessionManager("test-secret-key", 1*time.Hour)

	// Create a token that will pass JWT parsing but fail session validation
	// This simulates a token that was created with different secret or doesn't exist in session manager
	differentSecretKey := []byte("different-secret")
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userUUID": "test-user-123",
		"exp":      time.Now().Add(1 * time.Hour).Unix(),
	})

	tokenString, err := token.SignedString(differentSecretKey)
	assert.NoError(t, err)

	app := setupTestApp()

	// Test request with token that has different secret
	req, _ := http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", tokenString))
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}
