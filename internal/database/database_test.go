package database

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestInitDatabase_AlreadyInitialized(t *testing.T) {
	// Save original DB to restore later
	originalDB := DB
	defer func() {
		DB = originalDB
	}()

	// Create a dummy non-nil value to simulate already initialized database
	DB = new(gorm.DB)

	// Test parameters
	user := "testuser"
	password := "testpass"
	dbname := "testdb"
	host := "localhost"
	port := "5432"
	sslMode := "disable"

	// Call InitDatabase when DB is already set
	err := InitDatabase(user, password, dbname, host, port, sslMode)

	// Should return nil (no error) and DB should remain unchanged
	assert.NoError(t, err)
	assert.NotNil(t, DB)
}

func TestInitDatabase_InvalidConnection(t *testing.T) {
	// Save original DB to restore later
	originalDB := DB
	defer func() {
		DB = originalDB
	}()

	// Reset DB to nil for testing
	DB = nil

	// Test with invalid connection parameters that will cause gorm.Open to fail
	user := ""
	password := ""
	dbname := ""
	host := "invalid-host-that-does-not-exist-12345"
	port := "0"
	sslMode := "invalid"

	// Call InitDatabase with invalid parameters
	err := InitDatabase(user, password, dbname, host, port, sslMode)

	// Should return an error
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to connect to database")
	assert.Nil(t, DB)
}
