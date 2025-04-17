package auth

import (
	"errors"
	"github.com/gofiber/fiber/v2"
	"github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/sessionManager"
	"github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/utils/auth"
)

type LoginUser struct {
	Login    string `json:"login" validate:"required"`
	Password string `json:"password" validate:"required"`
}

func LoginHandler(c *fiber.Ctx) error {
	// Parse the request body into the LoginUser struct
	var user LoginUser
	if err := c.BodyParser(&user); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request payload",
		})
	}

	// Perform login logic here (e.g., check credentials against a database)
	userUUID, err := auth.ValidateLoginRequest(user.Login, user.Password)
	if err != nil {
		switch {
		case errors.Is(err, auth.ErrLoginRequired):
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Login is required",
			})
		case errors.Is(err, auth.ErrPasswordRequired):
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Password is required",
			})
		case errors.Is(err, auth.ErrInvalidCredentials):
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid credentials",
			})
		case errors.Is(err, auth.ErrUserNotFound):
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "User not found",
			})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Internal server error",
			})
		}
	}

	// return the sessions JWT token
	token, err := sessionManager.Instance.CreateSession(userUUID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create session",
		})
	}
	// Set the token in the response header
	c.Set("Authorization", "Bearer "+token)

	// Respond with a success message
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Login successful",
	})
}
