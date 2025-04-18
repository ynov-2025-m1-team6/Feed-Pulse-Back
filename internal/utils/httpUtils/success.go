package httpUtils

import (
	"github.com/gofiber/fiber/v2"
)

func NewSuccess(ctx *fiber.Ctx, status int, message string) error {
	s := HTTPSuccess{
		Success: message,
	}
	return ctx.Status(status).JSON(s)
}

// HTTPSuccess example
type HTTPSuccess struct {
	Success string `json:"success" example:"ok"`
}
