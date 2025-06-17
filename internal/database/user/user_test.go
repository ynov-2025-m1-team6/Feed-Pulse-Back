package user

import (
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/database"
	"github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/models/BaseModel"
	"github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/models/User"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func setupTestDB() (sqlmock.Sqlmock, func()) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	if err != nil {
		panic(err)
	}

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: db,
	}), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	originalDB := database.DB
	database.DB = gormDB

	cleanup := func() {
		database.DB = originalDB
		db.Close()
	}

	return mock, cleanup
}

func TestCreateUser_Success(t *testing.T) {
	mock, cleanup := setupTestDB()
	defer cleanup()

	testUser := &User.User{
		BaseModel: BaseModel.BaseModel{Id: 1},
		UUID:      "test-uuid",
		Username:  "testuser",
		Email:     "test@example.com",
		Password:  "password",
	}

	// Mock user creation
	mock.ExpectBegin()
	mock.ExpectQuery(".*INSERT INTO.*users.*").WillReturnRows(sqlmock.NewRows([]string{"id", "uuid"}).AddRow(1, "test-uuid"))
	mock.ExpectCommit()

	// Mock board creation
	mock.ExpectBegin()
	mock.ExpectQuery(".*INSERT INTO.*boards.*").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()

	// Mock board existence check for association (this is what was missing)
	mock.ExpectQuery(".*SELECT.*FROM.*boards.*WHERE.*id.*=").
		WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(1, "testuser Board"))

	// Mock user existence check for association
	mock.ExpectQuery(".*SELECT.*FROM.*users.*WHERE.*id.*=").
		WillReturnRows(sqlmock.NewRows([]string{"id", "username"}).AddRow(1, "testuser"))

	// Mock association creation
	mock.ExpectExec(".*INSERT INTO user_boards.*").WillReturnResult(sqlmock.NewResult(1, 1))

	err := CreateUser(testUser)
	assert.NoError(t, err)
}

func TestCreateUser_DatabaseError(t *testing.T) {
	mock, cleanup := setupTestDB()
	defer cleanup()

	testUser := &User.User{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password",
	}

	// Mock database error
	mock.ExpectBegin()
	mock.ExpectQuery(".*").WillReturnError(errors.New("database error"))
	mock.ExpectRollback()

	err := CreateUser(testUser)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database error")
}

func TestCreateUser_CreateBoardError(t *testing.T) {
	mock, cleanup := setupTestDB()
	defer cleanup()

	testUser := &User.User{
		BaseModel: BaseModel.BaseModel{Id: 1},
		Username:  "testuser",
		Email:     "test@example.com",
		Password:  "password",
	}
	// Mock user creation
	mock.ExpectBegin()
	mock.ExpectQuery(".*INSERT INTO.*users.*").WillReturnRows(sqlmock.NewRows([]string{"id", "uuid"}).AddRow(1, "test-uuid"))
	mock.ExpectCommit()
	// Mock board creation error
	mock.ExpectBegin()
	mock.ExpectQuery(".*INSERT INTO.*boards.*").WillReturnError(errors.New("board creation error"))
	mock.ExpectRollback()
	err := CreateUser(testUser)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "board creation error")
}

func TestGetUserByUsername_Success(t *testing.T) {
	mock, cleanup := setupTestDB()
	defer cleanup()

	expectedUser := User.User{
		BaseModel: BaseModel.BaseModel{Id: 1},
		Username:  "testuser",
		Email:     "test@example.com",
	}

	rows := sqlmock.NewRows([]string{"id", "username", "email"}).
		AddRow(expectedUser.Id, expectedUser.Username, expectedUser.Email)

	mock.ExpectQuery(".*").WillReturnRows(rows)

	result, err := GetUserByUsername("testuser")
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, expectedUser.Username, result.Username)
}

func TestGetUserByUsername_NotFound(t *testing.T) {
	mock, cleanup := setupTestDB()
	defer cleanup()

	mock.ExpectQuery(".*").WillReturnError(gorm.ErrRecordNotFound)

	result, err := GetUserByUsername("nonexistent")
	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestGetUserByID_Success(t *testing.T) {
	mock, cleanup := setupTestDB()
	defer cleanup()

	expectedUser := User.User{
		BaseModel: BaseModel.BaseModel{Id: 1},
		Username:  "testuser",
		Email:     "test@example.com",
	}

	rows := sqlmock.NewRows([]string{"id", "username", "email"}).
		AddRow(expectedUser.Id, expectedUser.Username, expectedUser.Email)

	mock.ExpectQuery(".*").WillReturnRows(rows)

	result, err := GetUserByID(1)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, expectedUser.Id, result.Id)
}

func TestGetUserByID_Error(t *testing.T) {
	mock, cleanup := setupTestDB()
	defer cleanup()

	mock.ExpectQuery(".*").WillReturnError(gorm.ErrRecordNotFound)

	result, err := GetUserByID(999)
	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestGetUserByUUID_Success(t *testing.T) {
	mock, cleanup := setupTestDB()
	defer cleanup()

	expectedUser := User.User{
		BaseModel: BaseModel.BaseModel{Id: 1},
		UUID:      "test-uuid",
		Username:  "testuser",
		Email:     "test@example.com",
	}
	rows := sqlmock.NewRows([]string{"id", "uuid", "username", "email"}).
		AddRow(expectedUser.Id, expectedUser.UUID, expectedUser.Username, expectedUser.Email)
	mock.ExpectQuery(".*").WillReturnRows(rows)
	result, err := GetUserByUUID("test-uuid")
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, expectedUser.UUID, result.UUID)
}

func TestGetUserByUUID_NotFound(t *testing.T) {
	mock, cleanup := setupTestDB()
	defer cleanup()

	mock.ExpectQuery(".*").WillReturnError(gorm.ErrRecordNotFound)

	result, err := GetUserByUUID("nonexistent-uuid")
	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestGetUserByEmail_Success(t *testing.T) {
	mock, cleanup := setupTestDB()
	defer cleanup()

	expectedUser := User.User{
		BaseModel: BaseModel.BaseModel{Id: 1},
		Email:     "test@example.com",
		Username:  "testuser",
	}
	rows := sqlmock.NewRows([]string{"id", "email", "username"}).
		AddRow(expectedUser.Id, expectedUser.Email, expectedUser.Username)
	mock.ExpectQuery(".*").WillReturnRows(rows)
	result, err := GetUserByEmail("test@example.com")
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, expectedUser.Email, result.Email)
}

func TestGetUserByEmail_NotFound(t *testing.T) {
	mock, cleanup := setupTestDB()
	defer cleanup()

	mock.ExpectQuery(".*").WillReturnError(gorm.ErrRecordNotFound)

	result, err := GetUserByEmail("test@example.com")
	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestGetUserEitherByEmailOrUsername_Success(t *testing.T) {
	mock, cleanup := setupTestDB()
	defer cleanup()

	expectedUser := User.User{
		BaseModel: BaseModel.BaseModel{Id: 1},
		Email:     "test@example.com",
		Username:  "testuser",
	}
	rows := sqlmock.NewRows([]string{"id", "email", "username"}).
		AddRow(expectedUser.Id, expectedUser.Email, expectedUser.Username)
	mock.ExpectQuery(".*").WillReturnRows(rows)
	result, err := GetUserEitherByEmailOrUsername("testuser")
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, expectedUser.Username, result.Username)
}

func TestGetUserEitherByEmailOrUsername_NotFound(t *testing.T) {
	mock, cleanup := setupTestDB()
	defer cleanup()

	mock.ExpectQuery(".*").WillReturnError(gorm.ErrRecordNotFound)

	result, err := GetUserEitherByEmailOrUsername("nonexistent")
	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestGetAllUsers_Success(t *testing.T) {
	mock, cleanup := setupTestDB()
	defer cleanup()

	expectedUsers := []User.User{
		{BaseModel: BaseModel.BaseModel{Id: 1}, Username: "user1", Email: "user1@example.com"},
		{BaseModel: BaseModel.BaseModel{Id: 2}, Username: "user2", Email: "user2@example.com"},
	}

	rows := sqlmock.NewRows([]string{"id", "username", "email"})
	for _, user := range expectedUsers {
		rows.AddRow(user.Id, user.Username, user.Email)
	}

	mock.ExpectQuery(".*").WillReturnRows(rows)

	result, err := GetAllUsers()
	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, expectedUsers[0].Username, result[0].Username)
}

func TestGetAllUsers_Error(t *testing.T) {
	mock, cleanup := setupTestDB()
	defer cleanup()

	mock.ExpectQuery(".*").WillReturnError(gorm.ErrRecordNotFound)

	result, err := GetAllUsers()
	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestUpdateUser_Success(t *testing.T) {
	mock, cleanup := setupTestDB()
	defer cleanup()

	testUser := &User.User{
		BaseModel: BaseModel.BaseModel{Id: 1},
		Username:  "updateduser",
		Email:     "updated@example.com",
	}

	mock.ExpectBegin()
	mock.ExpectExec(".*").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := UpdateUser(testUser)
	assert.NoError(t, err)
}

func TestDeleteUser_Success(t *testing.T) {
	mock, cleanup := setupTestDB()
	defer cleanup()

	mock.ExpectBegin()
	mock.ExpectExec(".*").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := DeleteUser(1)
	assert.NoError(t, err)
}

func TestDeleteUser_NotFound(t *testing.T) {
	mock, cleanup := setupTestDB()
	defer cleanup()

	mock.ExpectBegin()
	mock.ExpectExec(".*").WillReturnError(gorm.ErrRecordNotFound)
	mock.ExpectRollback()

	err := DeleteUser(999)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "record not found")
}
