package auth

import (
	"errors"
	"github.com/gofiber/fiber/v2"
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
	// get the token from the Authorization Bearer header
	token := c.Get("Authorization")
	if token == "" {
		return httpUtils.NewError(c, fiber.StatusUnauthorized, errors.New("no token provided"))
	}
	if len(token) < 7 {
		return httpUtils.NewError(c, fiber.StatusUnauthorized, errors.New("invalid token format"))
	}
	// remove the "Bearer " prefix
	token = token[7:]
	// check if the token is valid
	valid, err := sessionManager.Instance.ValidateSession(token)
	if err != nil || !valid {
		return httpUtils.NewError(c, fiber.StatusUnauthorized, errors.New("invalid token"))
	}
	// remove the token from the session
	sessionManager.Instance.DeleteSession(token)
	// return a success message
	return httpUtils.NewMessage(c, fiber.StatusOK, "logged out successfully")
}
