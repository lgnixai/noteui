package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Table struct {
	gorm.Model
	ID     uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	Name   string
	BaseID uuid.UUID
	Base   Base
	Fields []Field // Has Many Fields
	// Records []Record // Has Many Records - GORM doesn't map JSONB field directly like this relation
}

func (t *Table) BeforeCreate(tx *gorm.DB) (err error) {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return
}
