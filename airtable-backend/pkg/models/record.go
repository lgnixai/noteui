package models

import (
	"encoding/json"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Record struct {
	gorm.Model
	ID      uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	TableID uuid.UUID
	Table   Table
	Data    json.RawMessage `gorm:"type:jsonb"` // Use json.RawMessage for raw JSONB storage
}

func (r *Record) BeforeCreate(tx *gorm.DB) (err error) {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	// Initialize Data as an empty JSON object if nil
	if r.Data == nil {
		r.Data = json.RawMessage("{}")
	}
	return
}
