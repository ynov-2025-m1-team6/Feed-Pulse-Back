package httpUtils

import (
	"github.com/gofiber/fiber/v2"
)

func NewError(ctx *fiber.Ctx, status int, err error) error {
	er := HTTPError{
		Error: err.Error(),
	}
	return ctx.Status(status).JSON(er)
}

// HTTPError example
type HTTPError struct {
	Error string `json:"error" example:"bad request"`
}
