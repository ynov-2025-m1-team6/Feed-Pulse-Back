package User

import (
	"github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/models/BaseModel"
)

type User struct {
	BaseModel.BaseModel
	UUID     string `json:"uuid" gorm:"type:uuid;default:uuid_generate_v4();unique;not null"`
	Username string `json:"username" gorm:"unique;not null"`
	Email    string `json:"email" gorm:"unique;not null"`
	Password string `json:"password" gorm:"not null"`
}
