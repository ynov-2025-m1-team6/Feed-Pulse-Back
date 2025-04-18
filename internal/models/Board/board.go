package Board

import (
	"github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/models/BaseModel"
	"github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/models/Feedback"
	"github.com/ynov-2025-m1-team6/Feed-Pulse-Back/internal/models/User"
)

type Board struct {
	BaseModel.BaseModel
	Name string `json:"name" gorm:"not null"`

	Users     []User.User         `gorm:"many2many:user_boards;"`
	Feedbacks []Feedback.Feedback `gorm:"foreignKey:id"`
}
