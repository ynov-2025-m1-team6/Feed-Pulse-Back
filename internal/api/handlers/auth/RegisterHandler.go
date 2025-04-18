package auth

import (
	"errors"
	"github.com/gofiber/fiber/v2"
	userDB "github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/database/user"
	"github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/models/User"
	"github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/utils"
	"github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/utils/httpUtils"
	"golang.org/x/crypto/bcrypt"
)

// UserRepository defines an interface for user repository operations
type UserRepository interface {
	GetUserByUsername(username string) (*User.User, error)
	GetUserByEmail(email string) (*User.User, error)
	CreateUser(user *User.User) error
}

// DefaultUserRepository is the default implementation that uses the actual database
type DefaultUserRepository struct{}

func (r *DefaultUserRepository) GetUserByUsername(username string) (*User.User, error) {
	return userDB.GetUserByUsername(username)
}

func (r *DefaultUserRepository) GetUserByEmail(email string) (*User.User, error) {
	return userDB.GetUserByEmail(email)
}

func (r *DefaultUserRepository) CreateUser(user *User.User) error {
	return userDB.CreateUser(user)
}

// Default repository instance
var defaultUserRepo UserRepository = &DefaultUserRepository{}

type RegisterUser struct {
	Username string `json:"username" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8,max=50"`
}

// RegisterHandler godoc
// @Summary Register a new user
// @Description Register a new user with username, email and password
// @Tags Auth
// @Accept json
// @Produce json
// @Param user body RegisterUser true "User registration data"
// @Success 201 {object} httpUtils.HTTPMessage "registration successful response"
// @Failure 400 {object} httpUtils.HTTPError "bad request error"
// @Failure 409 {object} httpUtils.HTTPError "conflict error - resource already exists"
// @Failure 500 {object} httpUtils.HTTPError "internal server error"
// @Router /api/auth/register [post]
func RegisterHandler(c *fiber.Ctx) error {
	return RegisterHandlerWithRepo(c, defaultUserRepo)
}

// RegisterHandlerWithRepo is the testable version of RegisterHandler that accepts a repository
func RegisterHandlerWithRepo(c *fiber.Ctx, repo UserRepository) error {
	// Parse the request body into the RegisterUser struct
	var registerUser RegisterUser
	if err := c.BodyParser(&registerUser); err != nil {
		return httpUtils.NewError(c, fiber.StatusBadRequest, errors.New("invalid request payload"))
	}
	// Validate the request payload
	if registerUser.Username == "" || registerUser.Email == "" || registerUser.Password == "" {
		return httpUtils.NewError(c, fiber.StatusBadRequest, errors.New("username, email, and password are required"))
	}
	if len(registerUser.Password) < 8 {
		return httpUtils.NewError(c, fiber.StatusBadRequest, errors.New("password must be at least 8 characters long"))
	}
	if len(registerUser.Password) > 50 {
		return httpUtils.NewError(c, fiber.StatusBadRequest, errors.New("password must be at most 50 characters long"))
	}
	if !utils.IsValidEmail(registerUser.Email) {
		return httpUtils.NewError(c, fiber.StatusBadRequest, errors.New("invalid email format"))
	}
	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(registerUser.Password), bcrypt.DefaultCost)
	if err != nil {
		return httpUtils.NewError(c, fiber.StatusInternalServerError, errors.New("failed to hash password"))
	}
	// Create a new registerUser object
	newUser := &User.User{
		Username: registerUser.Username,
		Email:    registerUser.Email,
		Password: string(hashedPassword),
	}
	// get the user by username
	existingUser, err := repo.GetUserByUsername(newUser.Username)
	if err == nil && existingUser != nil {
		return httpUtils.NewError(c, fiber.StatusConflict, errors.New("username already exists"))
	}
	// get the user by email
	existingUser, err = repo.GetUserByEmail(newUser.Email)
	if err == nil && existingUser != nil {
		return httpUtils.NewError(c, fiber.StatusConflict, errors.New("email already exists"))
	}

	// Save the registerUser to the database
	err = repo.CreateUser(newUser)
	if err != nil {
		return httpUtils.NewError(c, fiber.StatusInternalServerError, errors.New("failed to create user"))
	}

	// Return a success response
	return httpUtils.NewMessage(c, fiber.StatusCreated, "user registered successfully")
}
