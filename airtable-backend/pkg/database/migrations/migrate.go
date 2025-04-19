package migrations

import (
	"log"

	"gorm.io/gorm"
)

func RunMigrations(db *gorm.DB) error {
	// Run initial schema first
	if err := InitialSchema(db); err != nil {
		log.Printf("Error running initial schema: %v", err)
		return err
	}

	// Run subsequent migrations
	migrations := []func(*gorm.DB) error{
		AddFieldKey,
		// Add more migrations here as needed
	}

	for _, migration := range migrations {
		if err := migration(db); err != nil {
			log.Printf("Error running migration: %v", err)
			return err
		}
	}

	return nil
}
