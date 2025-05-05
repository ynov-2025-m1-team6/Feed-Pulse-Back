package Board

import (
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/database"
	"github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/models/BaseModel"
	"github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/models/Board"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// setupTest creates a mock database connection for testing
func setupTest() (sqlmock.Sqlmock, error) {
	// Create a mock database connection
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	if err != nil {
		return nil, err
	}

	pgDB := postgres.New(postgres.Config{
		Conn:                 db,
		PreferSimpleProtocol: true,
	})

	gormDB, err := gorm.Open(pgDB, &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// Replace the global DB with our mock
	database.DB = gormDB

	return mock, nil
}

func TestGetAllBoards(t *testing.T) {
	mock, err := setupTest()
	if err != nil {
		t.Fatalf("Error setting up test: %v", err)
	}

	// Define test data
	testBoards := []Board.Board{
		{
			BaseModel: BaseModel.BaseModel{Id: 1},
			Name:      "Test Board 1",
		},
		{
			BaseModel: BaseModel.BaseModel{Id: 2},
			Name:      "Test Board 2",
		},
	}

	// Setup expectations
	rows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "name"})
	for _, b := range testBoards {
		rows.AddRow(b.Id, b.CreatedAt, b.UpdatedAt, b.Name)
	}

	// Match any SELECT query
	mock.ExpectQuery(`SELECT (.+) FROM "boards"`).
		WillReturnRows(rows)

	// Call the function we're testing
	boards, err := GetAllBoards()

	// Assert expectations
	assert.Nil(t, err)
	assert.Len(t, boards, 2)
	if len(boards) >= 2 {
		assert.Equal(t, testBoards[0].Id, boards[0].Id)
		assert.Equal(t, testBoards[1].Id, boards[1].Id)
		assert.Equal(t, testBoards[0].Name, boards[0].Name)
		assert.Equal(t, testBoards[1].Name, boards[1].Name)
	}

	// Test with database error
	mock.ExpectQuery(`SELECT (.+) FROM "boards"`).
		WillReturnError(errors.New("database error"))

	_, err = GetAllBoards()
	assert.NotNil(t, err)
}

func TestGetBoardByID(t *testing.T) {
	mock, err := setupTest()
	if err != nil {
		t.Fatalf("Error setting up test: %v", err)
	}

	// Define test data
	testBoard := Board.Board{
		BaseModel: BaseModel.BaseModel{Id: 1},
		Name:      "Test Board 1",
	}

	// Test successful case
	rows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "name"}).
		AddRow(testBoard.Id, testBoard.CreatedAt, testBoard.UpdatedAt, testBoard.Name)

	// GORM generates a query with multiple params, one for ID and one for LIMIT
	mock.ExpectQuery(`SELECT (.+) FROM "boards" WHERE (.+)`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(rows)

	board, err := GetBoardByID(1)
	assert.Nil(t, err)
	assert.Equal(t, testBoard.Id, board.Id)
	assert.Equal(t, testBoard.Name, board.Name)

	// Test not found error
	mock.ExpectQuery(`SELECT (.+) FROM "boards" WHERE (.+)`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnError(gorm.ErrRecordNotFound)

	_, err = GetBoardByID(999)
	assert.NotNil(t, err)
	assert.Equal(t, "board not found", err.Error())

	// Test other error
	mock.ExpectQuery(`SELECT (.+) FROM "boards" WHERE (.+)`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnError(errors.New("database error"))

	_, err = GetBoardByID(2)
	assert.NotNil(t, err)
}

func TestCreateBoard(t *testing.T) {
	mock, err := setupTest()
	if err != nil {
		t.Fatalf("Error setting up test: %v", err)
	}

	// Define test data
	testBoard := Board.Board{
		Name: "New Test Board",
	}

	// Setup expectations for the create operation
	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "boards"`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), "New Test Board").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()

	// Call the function we're testing
	createdBoard, err := CreateBoard(testBoard)

	// Assert expectations
	assert.Nil(t, err)
	assert.Equal(t, testBoard.Name, createdBoard.Name)

	// Test with database error
	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "boards"`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), "New Test Board").
		WillReturnError(errors.New("database error"))
	mock.ExpectRollback()

	_, err = CreateBoard(testBoard)
	assert.NotNil(t, err)
}

func TestUpdateBoard(t *testing.T) {
	mock, err := setupTest()
	if err != nil {
		t.Fatalf("Error setting up test: %v", err)
	}

	// Define test data
	testBoard := Board.Board{
		BaseModel: BaseModel.BaseModel{Id: 1},
		Name:      "Updated Board Name",
	}

	// Setup expectations for fetch operation (check if exists)
	rows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "name"}).
		AddRow(1, time.Now(), time.Now(), "Old Board Name")

	mock.ExpectQuery(`SELECT (.+) FROM "boards" WHERE (.+)`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(rows)

	// Setup expectations for the update operation
	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "boards" SET (.+) WHERE (.+)`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	// Call the function we're testing
	updatedBoard, err := UpdateBoard(testBoard)

	// Assert expectations
	assert.Nil(t, err)
	assert.Equal(t, testBoard.Id, updatedBoard.Id)
	assert.Equal(t, testBoard.Name, updatedBoard.Name)

	// Test not found error
	mock.ExpectQuery(`SELECT (.+) FROM "boards" WHERE (.+)`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnError(gorm.ErrRecordNotFound)

	_, err = UpdateBoard(Board.Board{BaseModel: BaseModel.BaseModel{Id: 999}})
	assert.NotNil(t, err)
	assert.Equal(t, "board not found", err.Error())

	// Test other error during fetch
	mock.ExpectQuery(`SELECT (.+) FROM "boards" WHERE (.+)`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnError(errors.New("database error"))

	_, err = UpdateBoard(Board.Board{BaseModel: BaseModel.BaseModel{Id: 2}})
	assert.NotNil(t, err)

	// Test error during update
	mock.ExpectQuery(`SELECT (.+) FROM "boards" WHERE (.+)`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(3))

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "boards" SET (.+) WHERE (.+)`).
		WillReturnError(errors.New("update error"))
	mock.ExpectRollback()

	_, err = UpdateBoard(Board.Board{BaseModel: BaseModel.BaseModel{Id: 3}})
	assert.NotNil(t, err)
}

func TestDeleteBoard(t *testing.T) {
	mock, err := setupTest()
	if err != nil {
		t.Fatalf("Error setting up test: %v", err)
	}

	// Setup expectations for fetch operation (check if exists)
	rows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "name"}).
		AddRow(1, time.Now(), time.Now(), "Board to Delete")

	mock.ExpectQuery(`SELECT (.+) FROM "boards" WHERE (.+)`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(rows)

	// Setup expectations for transaction
	mock.ExpectBegin()

	// Expect retrieval of related feedbacks
	feedbackRows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "date", "channel", "text", "board_id"}).
		AddRow(1, time.Now(), time.Now(), time.Now(), "email", "Test feedback 1", 1).
		AddRow(2, time.Now(), time.Now(), time.Now(), "web", "Test feedback 2", 1)

	mock.ExpectQuery(`SELECT (.+) FROM "feedbacks" WHERE (.+)`).
		WithArgs(1).
		WillReturnRows(feedbackRows)

	// Expect deletion of analyses for each feedback
	mock.ExpectExec(`DELETE FROM analyses WHERE feedback_id = \$1`).
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectExec(`DELETE FROM "feedbacks"`).
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectExec(`DELETE FROM analyses WHERE feedback_id = \$1`).
		WithArgs(2).
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectExec(`DELETE FROM "feedbacks"`).
		WithArgs(2).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Expect deletion of user associations
	mock.ExpectExec(`DELETE FROM user_boards WHERE board_id = \$1`).
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 2)) // 2 user associations deleted

	// Expect deletion of board
	mock.ExpectExec(`DELETE FROM "boards"`).
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 1)) // 1 board deleted

	mock.ExpectCommit()

	// Call the function we're testing
	err = DeleteBoard(1)

	// Assert expectations
	assert.Nil(t, err)

	// Test not found error
	mock.ExpectQuery(`SELECT (.+) FROM "boards" WHERE (.+)`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnError(gorm.ErrRecordNotFound)

	err = DeleteBoard(999)
	assert.NotNil(t, err)
	assert.Equal(t, "board not found", err.Error())

	// Test transaction error (e.g., during feedback retrieval)
	mock.ExpectQuery(`SELECT (.+) FROM "boards" WHERE (.+)`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(2))

	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT (.+) FROM "feedbacks" WHERE (.+)`).
		WithArgs(2).
		WillReturnError(errors.New("transaction error"))
	mock.ExpectRollback()

	err = DeleteBoard(2)
	assert.NotNil(t, err)
}

func TestGetBoardByName(t *testing.T) {
	mock, err := setupTest()
	if err != nil {
		t.Fatalf("Error setting up test: %v", err)
	}

	// Define test data
	testBoards := []Board.Board{
		{
			BaseModel: BaseModel.BaseModel{Id: 1},
			Name:      "Specific Board Name",
		},
		{
			BaseModel: BaseModel.BaseModel{Id: 2},
			Name:      "Specific Board Name", // Same name to test multiple results
		},
	}

	// Setup expectations
	rows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "name"})
	for _, b := range testBoards {
		rows.AddRow(b.Id, b.CreatedAt, b.UpdatedAt, b.Name)
	}

	mock.ExpectQuery(`SELECT (.+) FROM "boards" WHERE (.+)`).
		WithArgs("Specific Board Name").
		WillReturnRows(rows)

	// Call the function we're testing
	boards, err := GetBoardByName("Specific Board Name")

	// Assert expectations
	assert.Nil(t, err)
	assert.Len(t, boards, 2)
	if len(boards) >= 2 {
		assert.Equal(t, "Specific Board Name", boards[0].Name)
		assert.Equal(t, "Specific Board Name", boards[1].Name)
	}

	// Test with database error
	mock.ExpectQuery(`SELECT (.+) FROM "boards" WHERE (.+)`).
		WithArgs("Non-existent Board").
		WillReturnError(errors.New("database error"))

	_, err = GetBoardByName("Non-existent Board")
	assert.NotNil(t, err)
}

func TestGetBoardsWithUsers(t *testing.T) {
	mock, err := setupTest()
	if err != nil {
		t.Fatalf("Error setting up test: %v", err)
	}

	// Define test data
	testBoard := Board.Board{
		BaseModel: BaseModel.BaseModel{Id: 1},
		Name:      "Board With Users",
	}

	// Setup expectations for the main board query
	boardRows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "name"}).
		AddRow(testBoard.Id, testBoard.CreatedAt, testBoard.UpdatedAt, testBoard.Name)

	// First query: Get the board - Note that we need to match 2 arguments (id and LIMIT)
	mock.ExpectQuery(`SELECT (.+) FROM "boards" WHERE id = \$1 ORDER BY "boards"."id" LIMIT \$2`).
		WithArgs(1, 1). // GORM sends id=1 and LIMIT 1
		WillReturnRows(boardRows)

	// Second query: Get the user_boards join records
	userBoardRows := sqlmock.NewRows([]string{"board_id", "user_id"}).
		AddRow(1, 10).
		AddRow(1, 11)

	mock.ExpectQuery(`SELECT (.+) FROM "user_boards" WHERE "user_boards"."board_id" = \$1`).
		WithArgs(1).
		WillReturnRows(userBoardRows)

	// Third query: Get the users
	userRows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "uuid", "username", "email", "password"}).
		AddRow(10, time.Now(), time.Now(), "uuid-1", "user1", "user1@example.com", "password1").
		AddRow(11, time.Now(), time.Now(), "uuid-2", "user2", "user2@example.com", "password2")

	mock.ExpectQuery(`SELECT (.+) FROM "users" WHERE "users"."id" IN \(\$1,\$2\)`).
		WithArgs(10, 11).
		WillReturnRows(userRows)

	// Call the function we're testing
	board, err := GetBoardsWithUsers(1)

	// Assert expectations
	assert.Nil(t, err)
	assert.Equal(t, testBoard.Id, board.Id)
	assert.Equal(t, testBoard.Name, board.Name)
	assert.Equal(t, 2, len(board.Users)) // Check that we have the expected number of users

	// Test not found error
	mock.ExpectQuery(`SELECT (.+) FROM "boards" WHERE id = \$1 ORDER BY "boards"."id" LIMIT \$2`).
		WithArgs(999, 1).
		WillReturnError(gorm.ErrRecordNotFound)

	_, err = GetBoardsWithUsers(999)
	assert.NotNil(t, err)
	assert.Equal(t, "board not found", err.Error())

	// Test other error
	mock.ExpectQuery(`SELECT (.+) FROM "boards" WHERE id = \$1 ORDER BY "boards"."id" LIMIT \$2`).
		WithArgs(2, 1).
		WillReturnError(errors.New("database error"))

	_, err = GetBoardsWithUsers(2)
	assert.NotNil(t, err)
}

func TestGetBoardsWithFeedbacks(t *testing.T) {
	mock, err := setupTest()
	if err != nil {
		t.Fatalf("Error setting up test: %v", err)
	}

	// Define test data
	testBoard := Board.Board{
		BaseModel: BaseModel.BaseModel{Id: 1},
		Name:      "Board With Feedbacks",
	}

	// Setup expectations - the preload function is complex to mock
	// We'll just test the basic query and result structure
	rows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "name"}).
		AddRow(testBoard.Id, testBoard.CreatedAt, testBoard.UpdatedAt, testBoard.Name)

	mock.ExpectQuery(`SELECT (.+) FROM "boards"`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(rows)

	// For the feedbacks preload, we need to match a different query
	// This is a simplification, as the actual preload query can be complex
	feedbackRows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "date", "channel", "text", "board_id"})
	// No rows returned is fine for this test

	mock.ExpectQuery(`SELECT (.+) FROM "feedbacks"`).
		WillReturnRows(feedbackRows)

	// Call the function we're testing
	board, err := GetBoardsWithFeedbacks(1)

	// Assert expectations
	assert.Nil(t, err)
	assert.Equal(t, testBoard.Id, board.Id)
	assert.Equal(t, testBoard.Name, board.Name)

	// Test not found error
	mock.ExpectQuery(`SELECT (.+) FROM "boards" WHERE (.+)`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnError(gorm.ErrRecordNotFound)

	_, err = GetBoardsWithFeedbacks(999)
	assert.NotNil(t, err)
	assert.Equal(t, "board not found", err.Error())

	// Test other error
	mock.ExpectQuery(`SELECT (.+) FROM "boards" WHERE (.+)`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnError(errors.New("database error"))

	_, err = GetBoardsWithFeedbacks(2)
	assert.NotNil(t, err)
}

func TestGetBoardsByUserID(t *testing.T) {
	mock, err := setupTest()
	if err != nil {
		t.Fatalf("Error setting up test: %v", err)
	}

	// Define expected boards that should be returned
	expectedBoards := []Board.Board{
		{
			BaseModel: BaseModel.BaseModel{Id: 1},
			Name:      "Board 1 for User",
		},
		{
			BaseModel: BaseModel.BaseModel{Id: 2},
			Name:      "Board 2 for User",
		},
	}

	// Setup mock rows for the JOIN query
	rows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "name"})
	for _, board := range expectedBoards {
		rows.AddRow(board.Id, board.CreatedAt, board.UpdatedAt, board.Name)
	}

	// Expect a JOIN query on user_boards and a WHERE condition for the user ID
	mock.ExpectQuery(`SELECT (.+) FROM "boards" JOIN user_boards ON boards.id = user_boards.board_id WHERE user_boards.user_id = \$1`).
		WithArgs(10). // User ID 10 for testing
		WillReturnRows(rows)

	// Call the function we're testing
	boards, err := GetBoardsByUserID(10)

	// Assertions
	assert.Nil(t, err)
	assert.Equal(t, len(expectedBoards), len(boards))
	if len(boards) >= 2 {
		assert.Equal(t, expectedBoards[0].Id, boards[0].Id)
		assert.Equal(t, expectedBoards[0].Name, boards[0].Name)
		assert.Equal(t, expectedBoards[1].Id, boards[1].Id)
		assert.Equal(t, expectedBoards[1].Name, boards[1].Name)
	}

	// Test database error scenario
	mock.ExpectQuery(`SELECT (.+) FROM "boards" JOIN user_boards ON boards.id = user_boards.board_id WHERE user_boards.user_id = \$1`).
		WithArgs(11). // Different user ID for error test
		WillReturnError(errors.New("database error"))

	// Call function with error condition
	_, err = GetBoardsByUserID(11)
	assert.NotNil(t, err)
}
