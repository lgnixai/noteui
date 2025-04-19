package services

import (
	"testing"

	"airtable-backend/pkg/models"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// First create tables table
	err = db.Exec(`CREATE TABLE IF NOT EXISTS tables (
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

	// Then create fields table
	err = db.Exec(`CREATE TABLE IF NOT EXISTS fields (
		id TEXT PRIMARY KEY,
		table_id TEXT NOT NULL,
		name TEXT NOT NULL,
		key TEXT NOT NULL,
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

func TestFieldService_CreateField(t *testing.T) {
	db := setupTestDB(t)
	service := NewFieldService(db)
	tableService := NewTableService(db)

	// Create test table first
	table := &models.Table{
		BaseID: uuid.New(),
		Name:   "Test Table",
	}
	err := tableService.CreateTable(table)
	assert.NoError(t, err)

	// Then create field
	field := &models.Field{
		TableID: table.ID,
		Name:    "Test Field",
		Type:    models.FieldTypeText,
	}

	err = service.CreateField(field)
	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, field.ID)
	assert.Equal(t, field.Name, field.Key)

	// Test duplicate field name
	duplicateField := &models.Field{
		TableID: table.ID,
		Name:    "Test Field",
		Type:    models.FieldTypeText,
	}
	err = service.CreateField(duplicateField)
	assert.Error(t, err)
}

func TestFieldService_GetFieldByID(t *testing.T) {
	db := setupTestDB(t)
	service := NewFieldService(db)
	tableService := NewTableService(db)

	// Create test table first
	table := &models.Table{
		BaseID: uuid.New(),
		Name:   "Test Table",
	}
	err := tableService.CreateTable(table)
	assert.NoError(t, err)

	// Then create field
	field := &models.Field{
		TableID: table.ID,
		Name:    "Test Field",
		Type:    models.FieldTypeText,
	}
	service.CreateField(field)

	// Test get existing field
	result, err := service.GetFieldByID(field.ID)
	assert.NoError(t, err)
	assert.Equal(t, field.Name, result.Name)
	assert.Equal(t, field.Type, result.Type)

	// Test get non-existent field
	_, err = service.GetFieldByID(uuid.New())
	assert.Error(t, err)
}

func TestFieldService_GetFieldsByTableID(t *testing.T) {
	db := setupTestDB(t)
	service := NewFieldService(db)
	tableService := NewTableService(db)

	// Create test table first
	table := &models.Table{
		BaseID: uuid.New(),
		Name:   "Test Table",
	}
	err := tableService.CreateTable(table)
	assert.NoError(t, err)

	// Then create fields
	field1 := &models.Field{
		TableID: table.ID,
		Name:    "Field 1",
		Type:    models.FieldTypeText,
	}
	field2 := &models.Field{
		TableID: table.ID,
		Name:    "Field 2",
		Type:    models.FieldTypeNumber,
	}
	service.CreateField(field1)
	service.CreateField(field2)

	// Test get fields for table
	fields, err := service.GetFieldsByTableID(table.ID)
	assert.NoError(t, err)
	assert.Len(t, fields, 2)

	// Test get fields for non-existent table
	fields, err = service.GetFieldsByTableID(uuid.New())
	assert.NoError(t, err)
	assert.Empty(t, fields)
}

func TestFieldService_UpdateField(t *testing.T) {
	db := setupTestDB(t)
	service := NewFieldService(db)
	tableService := NewTableService(db)

	// Create test table first
	table := &models.Table{
		BaseID: uuid.New(),
		Name:   "Test Table",
	}
	err := tableService.CreateTable(table)
	assert.NoError(t, err)

	// Then create field
	field := &models.Field{
		TableID: table.ID,
		Name:    "Original Name",
		Type:    models.FieldTypeText,
	}
	service.CreateField(field)

	// Update field
	field.Name = "Updated Name"
	field.Type = models.FieldTypeNumber
	err = service.UpdateField(field)
	assert.NoError(t, err)

	// Verify update
	result, err := service.GetFieldByID(field.ID)
	assert.NoError(t, err)
	assert.Equal(t, "Updated Name", result.Name)
	assert.Equal(t, models.FieldTypeNumber, result.Type)
}

func TestFieldService_DeleteField(t *testing.T) {
	db := setupTestDB(t)
	service := NewFieldService(db)
	tableService := NewTableService(db)

	// Create test table first
	table := &models.Table{
		BaseID: uuid.New(),
		Name:   "Test Table",
	}
	err := tableService.CreateTable(table)
	assert.NoError(t, err)

	// Then create field
	field := &models.Field{
		TableID: table.ID,
		Name:    "Test Field",
		Type:    models.FieldTypeText,
	}
	service.CreateField(field)

	// Delete field
	err = service.DeleteField(field.ID)
	assert.NoError(t, err)

	// Verify deletion
	_, err = service.GetFieldByID(field.ID)
	assert.Error(t, err)
}

func TestFieldService_UpdateFieldOrder(t *testing.T) {
	db := setupTestDB(t)
	service := NewFieldService(db)
	tableService := NewTableService(db)

	// Create test table first
	table := &models.Table{
		BaseID: uuid.New(),
		Name:   "Test Table",
	}
	err := tableService.CreateTable(table)
	assert.NoError(t, err)

	// Then create fields
	field1 := &models.Field{
		TableID: table.ID,
		Name:    "Field 1",
		Type:    models.FieldTypeText,
		Order:   0,
	}
	field2 := &models.Field{
		TableID: table.ID,
		Name:    "Field 2",
		Type:    models.FieldTypeNumber,
		Order:   1,
	}
	service.CreateField(field1)
	service.CreateField(field2)

	// Update field order
	fieldOrders := map[uuid.UUID]int{
		field1.ID: 1,
		field2.ID: 0,
	}
	err = service.UpdateFieldOrder(table.ID, fieldOrders)
	assert.NoError(t, err)

	// Verify order update
	fields, err := service.GetFieldsByTableID(table.ID)
	assert.NoError(t, err)
	assert.Len(t, fields, 2)
	assert.Equal(t, field2.ID, fields[0].ID)
	assert.Equal(t, field1.ID, fields[1].ID)
}
