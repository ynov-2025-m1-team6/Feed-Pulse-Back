package auth

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/middleware"
	"github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/utils/httpUtils"
)

// UserInfoHandler godoc
// @Summary Get user information
// @Description Get the current user's information
// @Tags Auth
// @Accept json
// @Produce json
// @Success 200 {object} User.User "user information"
// @Failure 401 {object} httpUtils.HTTPError "authentication error"
// @Failure 404 {object} httpUtils.HTTPError "user not found error"
// @Failure 500 {object} httpUtils.HTTPError "internal server error"
// @Router /api/auth/user [get]
// @Security ApiKeyAuth
func UserInfoHandler(c *fiber.Ctx) error {
	return UserInfoHandlerWithRepo(c, defaultUserRepo)
}

// UserInfoHandlerWithRepo is the testable version of UserInfoHandler that accepts a repository
func UserInfoHandlerWithRepo(c *fiber.Ctx, repo UserRepository) error {
	// Get user UUID from context (set by middleware)
	userUUID, ok := middleware.GetUserUUID(c)
	if !ok {
		return httpUtils.NewError(c, fiber.StatusUnauthorized, errors.New("unauthorized: user not found in context"))
	}

	// Get user information from database
	user, err := repo.GetUserByUUID(userUUID)
	if err != nil {
		return httpUtils.NewError(c, fiber.StatusNotFound, errors.New("user not found"))
	}

	// Clear sensitive information
	user.Password = ""

	// Return user information
	return c.JSON(user)
}
