package migrations

import (
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	// Create a temporary test database
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Run migrations
	if err := RunMigrations(db); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	return db
}

func TestInitialSchema(t *testing.T) {
	db := setupTestDB(t)

	// Verify tables were created
	var tables []struct {
		Name string
	}
	if err := db.Raw("SELECT name FROM sqlite_master WHERE type='table'").Scan(&tables).Error; err != nil {
		t.Fatalf("Failed to query tables: %v", err)
	}

	expectedTables := map[string]bool{
		"tables":  true,
		"fields":  true,
		"records": true,
	}

	for _, table := range tables {
		if !expectedTables[table.Name] {
			t.Errorf("Unexpected table found: %s", table.Name)
		}
		delete(expectedTables, table.Name)
	}

	if len(expectedTables) > 0 {
		t.Errorf("Missing tables: %v", expectedTables)
	}
}

func TestAddFieldKey(t *testing.T) {
	db := setupTestDB(t)

	// Create a test table
	table := struct {
		ID   uint
		Name string
	}{
		Name: "Test Table",
	}
	if err := db.Table("tables").Create(&table).Error; err != nil {
		t.Fatalf("Failed to create test table: %v", err)
	}

	// Create a test field without a key
	field := struct {
		ID      uint
		TableID uint
		Name    string
		Type    string
	}{
		TableID: table.ID,
		Name:    "Test Field",
		Type:    "text",
	}
	if err := db.Table("fields").Create(&field).Error; err != nil {
		t.Fatalf("Failed to create test field: %v", err)
	}

	// Run the migration
	if err := AddFieldKey(db); err != nil {
		t.Fatalf("Failed to run AddFieldKey migration: %v", err)
	}

	// Verify the field has a key
	var updatedField struct {
		Key string
	}
	if err := db.Table("fields").Where("id = ?", field.ID).First(&updatedField).Error; err != nil {
		t.Fatalf("Failed to query updated field: %v", err)
	}

	if updatedField.Key == "" {
		t.Error("Field key was not set after migration")
	}
}
