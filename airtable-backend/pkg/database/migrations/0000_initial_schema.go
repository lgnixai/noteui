package migrations

import (
	"log"

	"gorm.io/gorm"
)

func InitialSchema(db *gorm.DB) error {
	// Create tables table first
	if err := db.Exec(`
		CREATE TABLE IF NOT EXISTS tables (
			id UUID PRIMARY KEY,
			name TEXT NOT NULL,
			description TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`).Error; err != nil {
		log.Printf("Error creating tables table: %v", err)
		return err
	}

	// Create fields table with foreign key constraint
	if err := db.Exec(`
		CREATE TABLE IF NOT EXISTS fields (
			id UUID PRIMARY KEY,
			table_id UUID NOT NULL,
			name TEXT NOT NULL,
			key TEXT NOT NULL,
			type TEXT NOT NULL,
			description TEXT,
			validation TEXT,
			"order" INTEGER NOT NULL DEFAULT 0,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (table_id) REFERENCES tables(id) ON DELETE CASCADE
		)
	`).Error; err != nil {
		log.Printf("Error creating fields table: %v", err)
		return err
	}

	// Create records table with foreign key constraint
	if err := db.Exec(`
		CREATE TABLE IF NOT EXISTS records (
			id UUID PRIMARY KEY,
			table_id UUID NOT NULL,
			data JSONB NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (table_id) REFERENCES tables(id) ON DELETE CASCADE
		)
	`).Error; err != nil {
		log.Printf("Error creating records table: %v", err)
		return err
	}

	return nil
}
