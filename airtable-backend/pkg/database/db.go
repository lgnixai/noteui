package database

import (
	"log"

	"airtable-backend/configs"
	"airtable-backend/pkg/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDB(cfg *configs.Config) {
	var err error
	DB, err = gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	log.Println("Database connected successfully")

	// AutoMigrate models
	err = DB.AutoMigrate(&models.User{}, &models.Base{}, &models.Table{}, &models.Field{}, &models.Record{})
	if err != nil {
		log.Fatalf("Failed to auto migrate database: %v", err)
	}
	log.Println("Database migration completed")

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
