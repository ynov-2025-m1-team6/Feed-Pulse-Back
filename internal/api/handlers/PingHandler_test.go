package handlers

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
)

func TestPingHandler(t *testing.T) {
	// Setup the app
	app := fiber.New()
	app.Get("/ping", PingHandler)

	// Create a test request
	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to test request: %v", err)
	}

	// Check the status code
	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("Expected status code %d, got %d", fiber.StatusOK, resp.StatusCode)
	}

	// Check the body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	expected := `{"message":"pong"}`
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
