package httpUtils

import (
	"github.com/gofiber/fiber/v2"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNewSuccess(t *testing.T) {
	app := fiber.New()

	// Define a test route
	app.Get("/test-success", func(c *fiber.Ctx) error {
		return NewSuccess(c, fiber.StatusCreated, "operation successful")
	})

	// Perform the request
	req := httptest.NewRequest(http.MethodGet, "/test-success", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to test request: %v", err)
	}

	// Check the status code
	if resp.StatusCode != fiber.StatusCreated {
		t.Errorf("Expected status code %d, got %d", fiber.StatusCreated, resp.StatusCode)
	}

	// Check the body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	expected := `{"success":"operation successful"}`
	if trimmedBody := strings.TrimSpace(string(body)); trimmedBody != expected {
		t.Errorf("Expected body %s, got %s", expected, trimmedBody)
	}

	// Check the content type
	contentType := resp.Header.Get("Content-Type")
	expectedContentType := "application/json"
	if !strings.Contains(contentType, expectedContentType) {
		t.Errorf("Expected Content-Type to contain %s, got %s", expectedContentType, contentType)
	}
}
