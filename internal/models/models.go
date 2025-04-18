package models

import (
	"github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/database"
	"github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/models/Analysis"
	"github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/models/Board"
	"github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/models/Feedback"
	"github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/models/User"
)

func InitModels() error {
	// Initialize models here if needed
	err := database.DB.AutoMigrate(
		&User.User{},
		&Board.Board{},
		&Feedback.Feedback{},
		&Analysis.Analysis{},
	)
	return err
}
