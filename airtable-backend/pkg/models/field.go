package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Define supported field types
const (
	FieldTypeText string = "text"
	FieldNumber   string = "number"
	TypeBoolean   string = "boolean"
	TypeDate      string = "date"
	// Add more types: singleSelect, multipleSelect, attachment, linkToAnotherRecord, etc.
)

type Field struct {
	gorm.Model
	ID      uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	TableID uuid.UUID
	Table   Table
	Name    string `gorm:"uniqueIndex:idx_field_table_name"` // Field name must be unique within a table
	Type    string // e.g., "text", "number", "boolean", "date"
	KeyName string `gorm:"uniqueIndex:idx_field_table_keyname"` // Key used in JSONB, maybe a short UUID or slug
}

func (f *Field) BeforeCreate(tx *gorm.DB) (err error) {
	if f.ID == uuid.Nil {
		f.ID = uuid.New()
	}
	// Generate a unique KeyName if not provided. Using UUID short string is good.
	if f.KeyName == "" {
		f.KeyName = uuid.New().String()[:8] // Use a short identifier
	}
	return
}
