package database

import (
	"testing"

	"airtable-backend/configs"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}
	return db
}

func TestConnectDB(t *testing.T) {
	// Create test config
	cfg := &configs.Config{
		DatabaseURL: "file::memory:?cache=shared",
	}

	// Test connection
	ConnectDB(cfg)

	// Verify connection
	if DB == nil {
		t.Error("Expected DB to be initialized")
	}

	// Test if tables were created
	var tables []string
	err := DB.Raw("SELECT name FROM sqlite_master WHERE type='table'").Scan(&tables).Error
	if err != nil {
		t.Fatalf("Failed to query tables: %v", err)
	}

	expectedTables := []string{"users", "bases", "tables", "fields", "records"}
	for _, expected := range expectedTables {
		found := false
		for _, table := range tables {
			if table == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected table %s to be created", expected)
		}
	}
}

func TestGINIndex(t *testing.T) {
	// Create test config
	cfg := &configs.Config{
		DatabaseURL: "file::memory:?cache=shared",
	}

	// Test connection
	ConnectDB(cfg)

	// Verify GIN index
	var count int64
	err := DB.Raw("SELECT count(*) FROM sqlite_master WHERE type='index' AND name='idx_records_data_gin'").Scan(&count).Error
	if err != nil {
		t.Fatalf("Failed to query indexes: %v", err)
	}

	if count == 0 {
		t.Error("Expected GIN index to be created")
	}
}
