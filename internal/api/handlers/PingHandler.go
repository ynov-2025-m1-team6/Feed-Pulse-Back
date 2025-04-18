package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/utils/httpUtils"
)

// PingHandler godoc
// @Summary Show the status of server
// @Description Get the status of server
// @Tags Ping
// @Accept */*
// @Produce json
// @Success 200 {object} httpUtils.HTTPMessage "server status response"
// @Router /ping [get]
func PingHandler(c *fiber.Ctx) error {
	// Respond with a 200 OK status and a simple message
	return httpUtils.NewMessage(c, fiber.StatusOK, "pong")
}
