package models

import (
	"github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/database"
	"github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/models/User"
)

func InitModels() error {
	// Initialize models here if needed
	err := database.DB.AutoMigrate(
		&User.User{},
	)
	return err
}
