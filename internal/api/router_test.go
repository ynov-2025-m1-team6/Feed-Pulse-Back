package api

import (
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

func TestSetupRoutes(t *testing.T) {
	t.Run("Setup routes successfully", func(t *testing.T) {
		app := fiber.New()
		err := SetupRoutes(app)
		assert.NoError(t, err)

		// Test that the app was properly configured
		assert.NotNil(t, app)
	})
}
