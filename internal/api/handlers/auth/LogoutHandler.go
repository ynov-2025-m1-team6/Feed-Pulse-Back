package auth

import (
	"errors"
	"github.com/getsentry/sentry-go"
	"github.com/gofiber/fiber/v2"
	"github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/middleware"
	"github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/sessionManager"
	"github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/utils/httpUtils"
)

// LogoutHandler godoc
// @Summary Logout a user
// @Description Logout a user by invalidating their session token
// @Tags Auth
// @Accept */*
// @Produce json
// @Success 200 {object} httpUtils.HTTPMessage "user successfully logged out"
// @Failure 401 {object} httpUtils.HTTPError "authentication error"
// @Router /api/auth/logout [get]
// @Security ApiKeyAuth
func LogoutHandler(c *fiber.Ctx) error {
	// First try to get token from context (set by middleware)
	token, ok := middleware.GetToken(c)

	// If not found in context, try to get it from the Authorization header (for backward compatibility)
	if !ok {
		// get the token from the Authorization Bearer header
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			sentry.CaptureEvent(&sentry.Event{
				Message: "No token provided in Authorization header",
				Level:   sentry.LevelError,
				Tags: map[string]string{
					"handler": "LogoutHandler",
					"action":  "logout",
				},
			})
			return httpUtils.NewError(c, fiber.StatusUnauthorized, errors.New("no token provided"))
		}
		if len(authHeader) < 7 {
			sentry.CaptureEvent(&sentry.Event{
				Message: "Invalid token format in Authorization header",
				Level:   sentry.LevelError,
				Tags: map[string]string{
					"handler": "LogoutHandler",
					"action":  "logout",
				},
			})
			return httpUtils.NewError(c, fiber.StatusUnauthorized, errors.New("invalid token format"))
		}
		// remove the "Bearer " prefix
		token = authHeader[7:]

		// check if the token is valid
		valid, err := sessionManager.Instance.ValidateSession(token)
		if err != nil || !valid {
			sentry.CaptureEvent(&sentry.Event{
				Message: "Invalid token provided",
				Level:   sentry.LevelError,
				Extra: map[string]interface{}{
					"token": token,
				},
				Tags: map[string]string{
					"handler": "LogoutHandler",
					"action":  "logout",
				},
			})
			return httpUtils.NewError(c, fiber.StatusUnauthorized, errors.New("invalid token"))
		}
	}

	// Remove the token from the session
	sessionManager.Instance.DeleteSession(token)

	// Return a success message
	return httpUtils.NewMessage(c, fiber.StatusOK, "logged out successfully")
}
