package auth

import (
	"github.com/gofiber/fiber/v2"
	"github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/database/user"
	"github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/models/User"
	"github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/utils"
	"golang.org/x/crypto/bcrypt"
)

type RegisterUser struct {
	Username string `json:"username" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8,max=50"`
}

func RegisterHandler(c *fiber.Ctx) error {
	// Parse the request body into the RegisterUser struct
	var registerUser RegisterUser
	if err := c.BodyParser(&registerUser); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request payload",
		})
	}
	// Validate the request payload
	if registerUser.Username == "" || registerUser.Email == "" || registerUser.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Username, email, and password are required",
		})
	}
	if len(registerUser.Password) < 8 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Password must be at least 8 characters long",
		})
	}
	if len(registerUser.Password) > 50 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Password must be at most 50 characters long",
		})
	}
	if !utils.IsValidEmail(registerUser.Email) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid email format",
		})
	}
	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(registerUser.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to hash password",
		})
	}
	// Create a new registerUser object
	newUser := &User.User{
		Username: registerUser.Username,
		Email:    registerUser.Email,
		Password: string(hashedPassword),
	}
	// get the user by username
	existingUser, err := user.GetUserByUsername(newUser.Username)
	if err == nil && existingUser != nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"error": "Username already exists",
		})
	}
	// get the user by email
	existingUser, err = user.GetUserByEmail(newUser.Email)
	if err == nil && existingUser != nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"error": "Email already exists",
		})
	}

	// Save the registerUser to the database
	err = user.CreateUser(newUser)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create user",
		})
	}

	// Return a success response
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "User registered successfully",
	})
}
