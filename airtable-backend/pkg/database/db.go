package database

import (
	"log"
	"strings"

	"airtable-backend/configs"
	"airtable-backend/pkg/database/migrations"
	"airtable-backend/pkg/models"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDB(cfg *configs.Config) {
	var err error
	var dialector gorm.Dialector

	// Use SQLite for testing
	if strings.HasPrefix(cfg.DatabaseURL, "file::memory:") {
		dialector = sqlite.Open(cfg.DatabaseURL)
	} else {
		dialector = postgres.Open(cfg.DatabaseURL)
	}

	DB, err = gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	log.Println("Database connected successfully")

	// Run migrations
	err = migrations.AddFieldKey(DB)
	if err != nil {
		log.Fatalf("Failed to run field key migration: %v", err)
	}

	// AutoMigrate models
	err = DB.AutoMigrate(&models.User{}, &models.Base{}, &models.Table{}, &models.Field{}, &models.Record{})
	if err != nil {
		log.Fatalf("Failed to auto migrate database: %v", err)
	}
	log.Println("Database migration completed")

	// Only create GIN index for PostgreSQL
	if !strings.HasPrefix(cfg.DatabaseURL, "file::memory:") {
		// Manually add GIN index for JSONB data field
		// This is crucial for performance on querying JSONB
		// Check if index exists first (optional but good practice)
		var count int64
		DB.Raw("SELECT count(*) FROM pg_indexes WHERE tablename = 'records' AND indexname = 'idx_records_data_gin'").Scan(&count)
		if count == 0 {
			err = DB.Exec("CREATE INDEX idx_records_data_gin ON records USING GIN (data)").Error
			if err != nil {
				log.Fatalf("Failed to create GIN index on records.data: %v", err)
			}
			log.Println("GIN index on records.data created")
		} else {
			log.Println("GIN index on records.data already exists")
		}
	}
}
