package migrations

import (
	"log"

	"gorm.io/gorm"
)

// AddFieldKey adds the key column to the fields table
func AddFieldKey(db *gorm.DB) error {
	// Check if the column exists using PostgreSQL information_schema
	var count int
	err := db.Raw(`
		SELECT COUNT(*)
		FROM information_schema.columns
		WHERE table_name = 'fields'
		AND column_name = 'key'
	`).Scan(&count).Error
	if err != nil {
		log.Printf("Failed to check if key column exists: %v", err)
		return err
	}

	if count == 0 {
		err = db.Exec("ALTER TABLE fields ADD COLUMN key VARCHAR(255)").Error
		if err != nil {
			log.Printf("Failed to add key column: %v", err)
			return err
		}
	}

	// Step 2: Update existing records with a default value
	err = db.Exec(`UPDATE fields SET key = name WHERE key IS NULL`).Error
	if err != nil {
		return err
	}

	// Step 3: Add unique index
	err = db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS idx_table_key ON fields(table_id, key)`).Error
	if err != nil {
		return err
	}

	// Step 4: Make the column NOT NULL
	err = db.Exec(`ALTER TABLE fields ALTER COLUMN key SET NOT NULL`).Error
	if err != nil {
		return err
	}

	return nil
}
