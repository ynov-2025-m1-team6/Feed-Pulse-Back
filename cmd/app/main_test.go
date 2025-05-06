package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"testing"
	"time"

	swagger "github.com/arsmn/fiber-swagger/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/tot0p/env"
	"github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/api"
	"github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/database"
	"github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/models"
	"github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/sessionManager"
	"github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/utils/sentimentAnalysis"
)

func LunchServer(app *fiber.App, port string) {

	// Initialize the database connection
	err := database.InitDatabase(env.Get("DB_USER"), env.Get("DB_PASSWORD"), env.Get("DB_NAME"), env.Get("DB_HOST"), env.Get("DB_PORT"), env.Get("DB_SSLMODE"))
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Initialize and automigrate the models
	err = models.InitModels()
	if err != nil {
		log.Fatalf("Failed to initialize models: %v", err)
	}

	// Initialize the session manager
	sentimentAnalysis.InitSentimentAnalysis(env.Get("MISTRAL_API_KEY"), env.Get("EMAIL_PASSWORD"))

	// Initialize the SessionManager
	if env.Get("SECRET_KEY") == "" {
		log.Fatal("SECRET_KEY environment variable is not set")
	}
	sessionManager.InitSessionManager(env.Get("SECRET_KEY"), 3*time.Hour)

	api.SetupRoutes(app)
	app.Get("/swagger/*", swagger.HandlerDefault)

	if port == "" {
		port = "3000"
	}

	// Start the server
	if err := app.Listen(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

type metrics struct {
	sync.Mutex
	success int
	failure int
	total   int
}

func (m *metrics) increment(success bool) {
	m.Lock()
	defer m.Unlock()
	m.total++
	if success {
		m.success++
	} else {
		m.failure++
	}
}

func CallPingHandler(url string) (int, int, error) {
	// Create a new request to the /ping endpoint
	req, err := http.NewRequest("GET", url+"/ping", nil)
	if err != nil {
		return 0, 0, err
	}

	// Send the request using the http package
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, 0, err
	}
	defer resp.Body.Close()

	// Read and discard body to prevent connection leaks
	_, err = io.ReadAll(resp.Body)
	if err != nil {
		return 0, 0, err
	}

	return resp.StatusCode, 0, nil
}

func CallLoginHandler(url string) (int, int, error) {
	// Create login credentials
	credentials := map[string]string{
		"login":    "ftecher3",
		"password": "12345678",
	}

	jsonData, err := json.Marshal(credentials)
	if err != nil {
		return 0, 0, err
	}

	// Create a new request to the /api/auth/login endpoint
	req, err := http.NewRequest("POST", url+"/api/auth/login", bytes.NewBuffer(jsonData))
	if err != nil {
		return 0, 0, err
	}
	req.Header.Set("Content-Type", "application/json")

	// Send the request using the http package
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, 0, err
	}
	defer resp.Body.Close()

	// Read and discard body to prevent connection leaks
	_, err = io.ReadAll(resp.Body)
	if err != nil {
		return 0, 0, err
	}

	// Return the status code and the authorization header length (if present)
	authHeader := resp.Header.Get("Authorization")
	return resp.StatusCode, len(authHeader), nil
}

func Test_Metric(t *testing.T) {
	// Setup server
	app := fiber.New(fiber.Config{
		ErrorHandler: customErrorHandler,
	})
	err := env.Load()
	if err != nil {
		if !os.IsNotExist(err) {
			log.Fatalf("Failed to load environment variables: %v", err)
		}
	}

	// Start server in a goroutine
	go func() {
		LunchServer(app, "3000")
	}()

	// Wait for the server to start
	time.Sleep(10 * time.Second)

	// Setup metrics
	m := &metrics{}

	// Setup wait group for parallel requests
	var wg sync.WaitGroup
	numRequests := 100
	wg.Add(numRequests)

	// Start timer
	start := time.Now()

	// Make parallel requests
	for i := 0; i < numRequests; i++ {
		go func() {
			defer wg.Done()

			reqStart := time.Now()
			status, length, err := CallPingHandler("http://localhost:3000")
			elapsed := time.Since(reqStart)

			if err != nil {
				t.Logf("Request error: %v", err)
				m.increment(false)
				return
			}

			m.increment(status == http.StatusOK)
			t.Logf("Request took %s - Status: %d, Response length: %d", elapsed, status, length)
		}()
	}

	// Wait for all requests to complete
	wg.Wait()
	elapsed := time.Since(start)

	// Log final metrics
	t.Logf("Test completed in %s", elapsed)
	t.Logf("Total requests: %d", m.total)
	t.Logf("Successful requests: %d", m.success)
	t.Logf("Failed requests: %d", m.failure)

	// Cleanup
	if err := app.Shutdown(); err != nil {
		t.Errorf("Error shutting down server: %v", err)
	}
}

func Test_LoginMetrics(t *testing.T) {
	// Setup server
	app := fiber.New(fiber.Config{
		ErrorHandler: customErrorHandler,
	})
	err := env.Load()
	if err != nil {
		if !os.IsNotExist(err) {
			log.Fatalf("Failed to load environment variables: %v", err)
		}
	}

	// Start server in a goroutine
	go func() {
		LunchServer(app, "3001")
	}()

	// Wait for the server to start
	time.Sleep(10 * time.Second)

	// Setup metrics
	m := &metrics{}

	// Setup wait group for parallel requests
	var wg sync.WaitGroup
	numRequests := 100
	wg.Add(numRequests)

	// Start timer
	start := time.Now()

	// Make parallel requests
	for i := 0; i < numRequests; i++ {
		go func() {
			defer wg.Done()

			reqStart := time.Now()
			status, authLen, err := CallLoginHandler("http://localhost:3001")
			elapsed := time.Since(reqStart)

			if err != nil {
				t.Logf("Request error: %v", err)
				m.increment(false)
				return
			}

			m.increment(status == http.StatusOK)
			t.Logf("Request took %s - Status: %d, Auth length: %d", elapsed, status, authLen)
		}()
	}

	// Wait for all requests to complete
	wg.Wait()
	elapsed := time.Since(start)

	// Log final metrics
	t.Logf("Test completed in %s", elapsed)
	t.Logf("Total requests: %d", m.total)
	t.Logf("Successful requests: %d", m.success)
	t.Logf("Failed requests: %d", m.failure)

	// Cleanup
	if err := app.Shutdown(); err != nil {
		t.Errorf("Error shutting down server: %v", err)
	}
}
