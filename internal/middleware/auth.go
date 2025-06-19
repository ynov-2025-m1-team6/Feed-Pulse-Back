// Package middleware contains application middleware handlers
package middleware

import (
	"errors"
	"fmt"
	"github.com/getsentry/sentry-go"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"

	"github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/sessionManager"
	"github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/utils/httpUtils"
)

const (
	// UserContextKey is the key used to store the user UUID in the context
	UserContextKey = "userUUID"
	// TokenContextKey is the key used to store the raw token in the context
	TokenContextKey = "token"
	// ClaimsContextKey is the key used to store token claims in the context
	ClaimsContextKey = "claims"
)

// AuthRequired middleware ensures that the request contains a valid session token
// It extracts the token from the Authorization header in Bearer format
func AuthRequired() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get authorization header
		authHeader := c.Get("Authorization")

		// Check if the authorization header exists and has the bearer format
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			sentry.CaptureEvent(&sentry.Event{
				Message: "Unauthorized access attempt: Missing or invalid authorization header",
				Level:   sentry.LevelError,
				Tags: map[string]string{
					"handler": "AuthRequired",
					"action":  "validate_token",
				},
			})
			return httpUtils.NewError(c, fiber.StatusUnauthorized, fmt.Errorf("unauthorized: Missing or invalid authorization header"))
		}

		// Extract the token from the authorization header
		// Format: "Bearer <token>"
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		// Validate the session using the session manager
		valid, err := sessionManager.Instance.ValidateSession(tokenString)
		if err != nil || !valid {
			sentry.CaptureEvent(&sentry.Event{
				Message: fmt.Sprintf("Unauthorized access attempt: %v", err),
				Level:   sentry.LevelError,
				Tags: map[string]string{
					"handler": "AuthRequired",
					"action":  "validate_token",
				},
			})
			return httpUtils.NewError(c, fiber.StatusUnauthorized, fmt.Errorf("unauthorized: Invalid or expired token"))
		}

		// Parse the JWT token to extract claims
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("invalid signing method")
			}
			return sessionManager.Instance.GetSecretKey(), nil
		})

		if err != nil || !token.Valid {
			sentry.CaptureEvent(&sentry.Event{
				Message: fmt.Sprintf("Unauthorized access attempt: %v", err),
				Level:   sentry.LevelError,
				Tags: map[string]string{
					"handler": "AuthRequired",
					"action":  "parse_token",
				},
			})
			return httpUtils.NewError(c, fiber.StatusUnauthorized, fmt.Errorf("unauthorized: Invalid token"))
		}

		// Extract claims
		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			// Store the token and claims in the context
			c.Locals(TokenContextKey, tokenString)
			c.Locals(ClaimsContextKey, claims)

			// Store the user UUID in the context for easy access
			if userUUID, ok := claims["userUUID"].(string); ok {
				c.Locals(UserContextKey, userUUID)
			}
		}

		// If the token is valid, proceed with the next handler
		return c.Next()
	}
}

// GetUserUUID retrieves the user UUID from the Fiber context
func GetUserUUID(c *fiber.Ctx) (string, bool) {
	userUUID, ok := c.Locals(UserContextKey).(string)
	return userUUID, ok && userUUID != ""
}

// GetToken retrieves the raw JWT token from the Fiber context
func GetToken(c *fiber.Ctx) (string, bool) {
	token, ok := c.Locals(TokenContextKey).(string)
	return token, ok && token != ""
}

// GetClaims retrieves all claims from the JWT token
func GetClaims(c *fiber.Ctx) (jwt.MapClaims, bool) {
	claims, ok := c.Locals(ClaimsContextKey).(jwt.MapClaims)
	return claims, ok
}

// GetClaimValue retrieves a specific claim value from the JWT token
func GetClaimValue(c *fiber.Ctx, key string) (interface{}, bool) {
	claims, ok := GetClaims(c)
	if !ok {
		return nil, false
	}

	value, exists := claims[key]
	return value, exists
}
