package models

import (
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/database"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestInitModels_Success(t *testing.T) {
	// Create a mock database connection that matches any query/exec
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	assert.NoError(t, err)
	defer db.Close()

	// Create a GORM DB instance with the mock
	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: db,
	}), &gorm.Config{})
	assert.NoError(t, err)

	// Set the mock database to the global DB variable
	originalDB := database.DB
	database.DB = gormDB
	defer func() {
		database.DB = originalDB
	}()

	// Mock all possible queries and executions that GORM might make
	// Use a generous number to cover all GORM operations
	for i := 0; i < 50; i++ {
		mock.ExpectExec(".*").WillReturnResult(sqlmock.NewResult(0, 0))
	}
	for i := 0; i < 50; i++ {
		mock.ExpectQuery(".*").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	}

	// Test the InitModels function
	err = InitModels()

	// Assert that no error occurred
	assert.NoError(t, err)
}

func TestInitModels_DatabaseError(t *testing.T) {
	// Create a mock database connection
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	assert.NoError(t, err)
	defer db.Close()

	// Create a GORM DB instance with the mock
	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: db,
	}), &gorm.Config{})
	assert.NoError(t, err)

	// Set the mock database to the global DB variable
	originalDB := database.DB
	database.DB = gormDB
	defer func() {
		database.DB = originalDB
	}()

	// Mock a database error on the first CREATE TABLE statement
	mock.ExpectExec("CREATE TABLE.*").WillReturnError(errors.New("database connection failed"))

	// Test the InitModels function
	err = InitModels()

	// Assert that an error occurred
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database connection failed")
}
