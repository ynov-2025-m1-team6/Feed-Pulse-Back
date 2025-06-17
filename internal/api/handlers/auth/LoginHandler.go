package auth

import (
	"errors"
	"github.com/getsentry/sentry-go"
	"github.com/gofiber/fiber/v2"
	"github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/sessionManager"
	"github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/utils/auth"
	"github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/utils/httpUtils"
)

type LoginUser struct {
	Login    string `json:"login" validate:"required"`
	Password string `json:"password" validate:"required"`
}

// LoginHandler godoc
// @Summary Authenticate a user
// @Description Login a user with username/email and password
// @Tags Auth
// @Accept json
// @Produce json
// @Param user body LoginUser true "User login credentials"
// @Success 200 {object} httpUtils.HTTPMessage "login successful response"
// @Failure 400 {object} httpUtils.HTTPError "bad request error"
// @Failure 401 {object} httpUtils.HTTPError "authentication error"
// @Failure 404 {object} httpUtils.HTTPError "user not found error"
// @Failure 500 {object} httpUtils.HTTPError "internal server error"
// @Router /api/auth/login [post]
func LoginHandler(c *fiber.Ctx) error {
	return LoginHandlerWithRepo(c, &auth.DefaultUserRepository{})
}

// LoginHandlerWithRepo is the testable version of LoginHandler that accepts a repository
func LoginHandlerWithRepo(c *fiber.Ctx, repo auth.UserRepository) error {
	// Parse the request body into the LoginUser struct
	var user LoginUser
	if err := c.BodyParser(&user); err != nil {
		sentry.CaptureEvent(&sentry.Event{
			Message: "Failed to parse login request body",
			Extra: map[string]interface{}{
				"error": err.Error(),
				"body":  string(c.Body()),
			},
			Level: sentry.LevelWarning,
			Tags: map[string]string{
				"handler": "LoginHandler",
				"action":  "parse_request",
			},
		})
		return httpUtils.NewError(c, fiber.StatusBadRequest, errors.New("invalid request payload"))
	}

	// Perform login logic here (e.g., check credentials against a database)
	userUUID, err := auth.ValidateLoginRequestWithRepo(user.Login, user.Password, repo)
	if err != nil {
		switch {
		case errors.Is(err, auth.ErrLoginRequired):
			sentry.CaptureEvent(&sentry.Event{
				Message: "No login provided in request",
				Extra: map[string]interface{}{
					"error": err.Error(),
					"body":  string(c.Body()),
				},
				Level: sentry.LevelWarning,
				Tags: map[string]string{
					"handler": "LoginHandler",
					"action":  "validate_login",
				},
			})
			return httpUtils.NewError(c, fiber.StatusBadRequest, errors.New("login is required"))
		case errors.Is(err, auth.ErrPasswordRequired):
			sentry.CaptureEvent(&sentry.Event{
				Message: "No password provided in request",
				Extra: map[string]interface{}{
					"error": err.Error(),
					"body":  string(c.Body()),
				},
				Level: sentry.LevelWarning,
				Tags: map[string]string{
					"handler": "LoginHandler",
					"action":  "validate_password",
				},
			})
			return httpUtils.NewError(c, fiber.StatusBadRequest, errors.New("password is required"))
		case errors.Is(err, auth.ErrInvalidCredentials):
			sentry.CaptureEvent(&sentry.Event{
				Message: "Invalid login credentials provided",
				Extra: map[string]interface{}{
					"error": err.Error(),
					"body":  string(c.Body()),
				},
				User: sentry.User{
					Username: user.Login,
				},
				Level: sentry.LevelWarning,
				Tags: map[string]string{
					"handler": "LoginHandler",
					"action":  "validate_credentials",
				},
			})
			return httpUtils.NewError(c, fiber.StatusUnauthorized, errors.New("invalid credentials"))
		case errors.Is(err, auth.ErrUserNotFound):
			sentry.CaptureEvent(&sentry.Event{
				Message: "User not found during login attempt",
				Extra: map[string]interface{}{
					"error": err.Error(),
					"body":  string(c.Body()),
				},
				User: sentry.User{
					Username: user.Login,
				},
				Level: sentry.LevelWarning,
				Tags: map[string]string{
					"handler": "LoginHandler",
					"action":  "user_not_found",
				},
			})
			return httpUtils.NewError(c, fiber.StatusNotFound, errors.New("user not found"))
		default:
			sentry.CaptureEvent(&sentry.Event{
				Message: "Unexpected error during login",
				Extra: map[string]interface{}{
					"error": err.Error(),
					"body":  string(c.Body()),
				},
				User: sentry.User{
					Username: user.Login,
				},
				Level: sentry.LevelError,
				Tags: map[string]string{
					"handler": "LoginHandler",
					"action":  "unexpected_error",
				},
			})
			return httpUtils.NewError(c, fiber.StatusInternalServerError, errors.New("internal server error"))
		}
	}

	// return the sessions JWT token
	token, err := sessionManager.Instance.CreateSession(userUUID)
	if err != nil {
		sentry.CaptureEvent(&sentry.Event{
			Message: "Failed to create session during login",
			Extra: map[string]interface{}{
				"error": err.Error(),
			},
			User: sentry.User{
				Username: user.Login,
			},
			Level: sentry.LevelError,
			Tags: map[string]string{
				"handler": "LoginHandler",
				"action":  "create_session",
			},
		})
		return httpUtils.NewError(c, fiber.StatusInternalServerError, errors.New("failed to create session"))
	}
	// Set the token in the response header
	c.Set("Authorization", "Bearer "+token)

	// Respond with a success message
	return httpUtils.NewMessage(c, fiber.StatusOK, "login successful")
}
