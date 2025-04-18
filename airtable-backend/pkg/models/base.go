package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Base struct {
	gorm.Model
	ID     uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	Name   string
	UserID uuid.UUID
	Tables []Table // Has Many Tables
}

func (b *Base) BeforeCreate(tx *gorm.DB) (err error) {
	if b.ID == uuid.Nil {
		b.ID = uuid.New()
	}
	return
}
