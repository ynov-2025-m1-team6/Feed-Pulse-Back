package BaseModel

import "time"

type BaseModel struct {
	Id        int `json:"id" gorm:"primaryKey;autoIncrement"`
	CreatedAt time.Time
	UpdatedAt time.Time
}
