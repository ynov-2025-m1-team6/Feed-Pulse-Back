package auth

import (
	"github.com/gofiber/fiber/v2"
	"github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/sessionManager"
)

func LogoutHandler(c *fiber.Ctx) error {
	// get the token from the Authorization Bearer header
	token := c.Get("Authorization")
	if token == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "No token provided",
		})
	}
	if len(token) < 7 {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid token format",
		})
	}
	// remove the "Bearer " prefix
	token = token[7:]
	// check if the token is valid
	valid, err := sessionManager.Instance.ValidateSession(token)
	if err != nil || !valid {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid token",
		})
	}
	// remove the token from the session
	sessionManager.Instance.DeleteSession(token)
	// return a success message
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Logged out successfully",
	})
}
