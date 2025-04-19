package migrations

import (
	"encoding/json"
	"testing"

	"airtable-backend/pkg/models"

	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// First create tables table
	err = db.Exec(`CREATE TABLE tables (
		id TEXT PRIMARY KEY,
		base_id TEXT NOT NULL,
		name TEXT NOT NULL,
		description TEXT,
		created_at DATETIME,
		updated_at DATETIME,
		deleted_at DATETIME
	)`).Error
	if err != nil {
		t.Fatalf("Failed to create tables table: %v", err)
	}

	// Then create fields table without the key column
	err = db.Exec(`CREATE TABLE fields (
		id TEXT PRIMARY KEY,
		table_id TEXT NOT NULL,
		name TEXT NOT NULL,
		type TEXT NOT NULL,
		description TEXT,
		validation TEXT,
		"order" INTEGER NOT NULL DEFAULT 0,
		created_at DATETIME,
		updated_at DATETIME,
		deleted_at DATETIME,
		FOREIGN KEY (table_id) REFERENCES tables(id)
	)`).Error
	if err != nil {
		t.Fatalf("Failed to create fields table: %v", err)
	}

	return db
}

func TestAddFieldKey(t *testing.T) {
	db := setupTestDB(t)

	// Create test table first
	table := models.Table{
		ID:     uuid.New(),
		BaseID: uuid.New(),
		Name:   "Test Table",
	}
	err := db.Exec("INSERT INTO tables (id, base_id, name) VALUES (?, ?, ?)",
		table.ID, table.BaseID, table.Name).Error
	if err != nil {
		t.Fatalf("Failed to create test table: %v", err)
	}

	// Then create test field
	validation := models.ValidationRule{}
	validationJSON, _ := json.Marshal(validation)
	field := models.Field{
		ID:         uuid.New(),
		TableID:    table.ID,
		Name:       "Test Field",
		Type:       models.FieldTypeText,
		Validation: validation,
	}
	err = db.Exec("INSERT INTO fields (id, table_id, name, type, validation) VALUES (?, ?, ?, ?, ?)",
		field.ID, field.TableID, field.Name, field.Type, validationJSON).Error
	if err != nil {
		t.Fatalf("Failed to create test field: %v", err)
	}

	// Run migration
	err = AddFieldKey(db)
	if err != nil {
		t.Fatalf("Migration failed: %v", err)
	}

	// Verify results
	var result models.Field
	err = db.First(&result, field.ID).Error
	if err != nil {
		t.Fatalf("Failed to fetch field: %v", err)
	}

	// Check if key was set correctly
	if result.Key != field.Name {
		t.Errorf("Expected key to be %s, got %s", field.Name, result.Key)
	}

	// Try to create a field with duplicate key
	duplicateField := models.Field{
		ID:         uuid.New(),
		TableID:    field.TableID,
		Name:       field.Name,
		Key:        field.Name,
		Type:       models.FieldTypeText,
		Validation: validation,
	}
	err = db.Exec("INSERT INTO fields (id, table_id, name, key, type, validation) VALUES (?, ?, ?, ?, ?, ?)",
		duplicateField.ID, duplicateField.TableID, duplicateField.Name, duplicateField.Key, duplicateField.Type, validationJSON).Error
	if err == nil {
		t.Error("Expected error when creating field with duplicate key")
	}
}
