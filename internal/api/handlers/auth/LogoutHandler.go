package auth

import (
	"errors"
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
	// Get token from context (already validated by middleware)
	token, ok := middleware.GetToken(c)
	if !ok {
		return httpUtils.NewError(c, fiber.StatusUnauthorized, errors.New("no token found in context"))
	}
	
	// Remove the token from the session
	sessionManager.Instance.DeleteSession(token)
	
	// Return a success message
	return httpUtils.NewMessage(c, fiber.StatusOK, "logged out successfully")
}
