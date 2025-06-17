package auth

import (
	"errors"
	"github.com/getsentry/sentry-go"

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
	GetUserByUUID(uuid string) (*User.User, error)
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

func (r *DefaultUserRepository) GetUserByUUID(uuid string) (*User.User, error) {
	return userDB.GetUserByUUID(uuid)
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
		sentry.CaptureEvent(&sentry.Event{
			Message: "Failed to parse registration request body",
			Extra: map[string]interface{}{
				"error": err.Error(),
				"body":  string(c.Body()),
			},
			Level: sentry.LevelWarning,
			Tags: map[string]string{
				"handler": "RegisterHandler",
				"action":  "parse_request",
			},
		})
		return httpUtils.NewError(c, fiber.StatusBadRequest, errors.New("invalid request payload"))
	}
	// Validate the request payload
	if registerUser.Username == "" || registerUser.Email == "" || registerUser.Password == "" {
		sentry.CaptureEvent(&sentry.Event{
			Message: "Missing required fields in registration request",
			Extra: map[string]interface{}{
				"username": registerUser.Username,
				"email":    registerUser.Email,
				"password": registerUser.Password,
			},
			Level: sentry.LevelWarning,
			Tags: map[string]string{
				"handler": "RegisterHandler",
				"action":  "validate_request",
			},
		})
		return httpUtils.NewError(c, fiber.StatusBadRequest, errors.New("username, email, and password are required"))
	}
	if len(registerUser.Password) < 8 {
		sentry.CaptureEvent(&sentry.Event{
			Message: "Password too short in registration request",
			Extra: map[string]interface{}{
				"password_length": len(registerUser.Password),
				"username":        registerUser.Username,
				"email":           registerUser.Email,
			},
			Level: sentry.LevelWarning,
			Tags: map[string]string{
				"handler": "RegisterHandler",
				"action":  "validate_password_length",
			},
		})
		return httpUtils.NewError(c, fiber.StatusBadRequest, errors.New("password must be at least 8 characters long"))
	}
	if len(registerUser.Password) > 50 {
		sentry.CaptureEvent(&sentry.Event{
			Message: "Password too long in registration request",
			Extra: map[string]interface{}{
				"password_length": len(registerUser.Password),
				"username":        registerUser.Username,
				"email":           registerUser.Email,
			},
			Level: sentry.LevelWarning,
			Tags: map[string]string{
				"handler": "RegisterHandler",
				"action":  "validate_password_length",
			},
		})
		return httpUtils.NewError(c, fiber.StatusBadRequest, errors.New("password must be at most 50 characters long"))
	}
	if !utils.IsValidEmail(registerUser.Email) {
		sentry.CaptureEvent(&sentry.Event{
			Message: "Invalid email format in registration request",
			Extra: map[string]interface{}{
				"email":    registerUser.Email,
				"username": registerUser.Username,
			},
			Level: sentry.LevelWarning,
			Tags: map[string]string{
				"handler": "RegisterHandler",
				"action":  "validate_email_format",
			},
		})
		return httpUtils.NewError(c, fiber.StatusBadRequest, errors.New("invalid email format"))
	}
	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(registerUser.Password), bcrypt.DefaultCost)
	if err != nil {
		sentry.CaptureEvent(&sentry.Event{
			Message: "Failed to hash password during registration",
			Extra: map[string]interface{}{
				"error":    err.Error(),
				"username": registerUser.Username,
				"email":    registerUser.Email,
			},
			Level: sentry.LevelError,
			Tags: map[string]string{
				"handler": "RegisterHandler",
				"action":  "hash_password",
			},
		})
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
		sentry.CaptureEvent(&sentry.Event{
			Message: "Username already exists during registration",
			Extra: map[string]interface{}{
				"username": newUser.Username,
				"email":    newUser.Email,
			},
			Level: sentry.LevelWarning,
			Tags: map[string]string{
				"handler": "RegisterHandler",
				"action":  "check_username_exists",
			},
		})
		return httpUtils.NewError(c, fiber.StatusConflict, errors.New("username already exists"))
	}
	// get the user by email
	existingUser, err = repo.GetUserByEmail(newUser.Email)
	if err == nil && existingUser != nil {
		sentry.CaptureEvent(&sentry.Event{
			Message: "Email already exists during registration",
			Extra: map[string]interface{}{
				"email":    newUser.Email,
				"username": newUser.Username,
			},
			Level: sentry.LevelWarning,
			Tags: map[string]string{
				"handler": "RegisterHandler",
				"action":  "check_email_exists",
			},
		})
		return httpUtils.NewError(c, fiber.StatusConflict, errors.New("email already exists"))
	}

	// Save the registerUser to the database
	err = repo.CreateUser(newUser)
	if err != nil {
		sentry.CaptureEvent(&sentry.Event{
			Message: "Failed to create user during registration",
			Extra: map[string]interface{}{
				"error":    err.Error(),
				"username": newUser.Username,
				"email":    newUser.Email,
			},
			Level: sentry.LevelError,
			Tags: map[string]string{
				"handler": "RegisterHandler",
				"action":  "create_user",
			},
		})
		return httpUtils.NewError(c, fiber.StatusInternalServerError, errors.New("failed to create user"))
	}

	// Return a success response
	return httpUtils.NewMessage(c, fiber.StatusCreated, "user registered successfully")
}
