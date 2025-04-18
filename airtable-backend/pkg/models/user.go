package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	ID   uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	Name string
	// Add auth fields later (e.g., Email, PasswordHash)
}

func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return
}
