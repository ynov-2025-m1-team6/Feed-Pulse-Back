package httpUtils

import (
	"github.com/gofiber/fiber/v2"
)

func NewMessage(ctx *fiber.Ctx, status int, message string) error {
	m := HTTPMessage{
		Message: message,
	}
	return ctx.Status(status).JSON(m)
}

// HTTPSuccess example
type HTTPMessage struct {
	Message string `json:"message" example:"ok"`
}
