package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"airtable-backend/pkg/models"
	"airtable-backend/pkg/services"

	"github.com/gin-gonic/gin"
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

func setupTestRouter(t *testing.T) (*gin.Engine, *FieldHandler) {
	db := setupTestDB(t)
	fieldService := services.NewFieldService(db)
	tableService := services.NewTableService(db)
	handler := NewFieldHandler(fieldService, tableService)

	router := gin.Default()
	router.POST("/tables/:tableId/fields", handler.CreateField)
	router.GET("/tables/:tableId/fields", handler.GetFieldsByTable)
	router.GET("/fields/:fieldId", handler.GetField)
	router.PUT("/fields/:fieldId", handler.UpdateField)
	router.DELETE("/fields/:fieldId", handler.DeleteField)
	router.PUT("/tables/:tableId/fields/order", handler.UpdateFieldOrder)

	return router, handler
}

func TestCreateField(t *testing.T) {
	router, handler := setupTestRouter(t)

	// Create test table first
	table := models.Table{
		BaseID: uuid.New(),
		Name:   "Test Table",
	}
	err := handler.TableService.CreateTable(&table)
	assert.NoError(t, err)

	// Then create field
	field := models.Field{
		TableID: table.ID,
		Name:    "Test Field",
		Key:     "test_field",
		Type:    models.FieldTypeText,
	}

	jsonData, _ := json.Marshal(field)
	req, _ := http.NewRequest("POST", "/tables/"+table.ID.String()+"/fields", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response models.Field
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, field.Name, response.Name)
	assert.Equal(t, field.Type, response.Type)
}

func TestGetFieldsByTable(t *testing.T) {
	router, handler := setupTestRouter(t)

	// Create test table first
	table := models.Table{
		BaseID: uuid.New(),
		Name:   "Test Table",
	}
	err := handler.TableService.CreateTable(&table)
	assert.NoError(t, err)

	// Then create fields
	field1 := models.Field{
		TableID: table.ID,
		Name:    "Field 1",
		Key:     "field_1",
		Type:    models.FieldTypeText,
	}
	field2 := models.Field{
		TableID: table.ID,
		Name:    "Field 2",
		Key:     "field_2",
		Type:    models.FieldTypeNumber,
	}

	handler.Service.CreateField(&field1)
	handler.Service.CreateField(&field2)

	req, _ := http.NewRequest("GET", "/tables/"+table.ID.String()+"/fields", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response []models.Field
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Len(t, response, 2)
}

func TestGetField(t *testing.T) {
	router, handler := setupTestRouter(t)

	// Create test table first
	table := models.Table{
		BaseID: uuid.New(),
		Name:   "Test Table",
	}
	err := handler.TableService.CreateTable(&table)
	assert.NoError(t, err)

	// Then create test field
	field := models.Field{
		TableID: table.ID,
		Name:    "Test Field",
		Key:     "test_field",
		Type:    models.FieldTypeText,
	}
	handler.Service.CreateField(&field)

	req, _ := http.NewRequest("GET", "/fields/"+field.ID.String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.Field
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, field.Name, response.Name)
	assert.Equal(t, field.Type, response.Type)
}

func TestUpdateField(t *testing.T) {
	router, handler := setupTestRouter(t)

	// Create test table first
	table := models.Table{
		BaseID: uuid.New(),
		Name:   "Test Table",
	}
	err := handler.TableService.CreateTable(&table)
	assert.NoError(t, err)

	// Then create test field
	field := models.Field{
		TableID: table.ID,
		Name:    "Original Name",
		Key:     "original_name",
		Type:    models.FieldTypeText,
	}
	handler.Service.CreateField(&field)

	// Update field
	updatedField := models.Field{
		Name: "Updated Name",
		Key:  "updated_name",
		Type: models.FieldTypeNumber,
	}
	jsonData, _ := json.Marshal(updatedField)
	req, _ := http.NewRequest("PUT", "/fields/"+field.ID.String(), bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.Field
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, updatedField.Name, response.Name)
	assert.Equal(t, updatedField.Type, response.Type)
}

func TestDeleteField(t *testing.T) {
	router, handler := setupTestRouter(t)

	// Create test table first
	table := models.Table{
		BaseID: uuid.New(),
		Name:   "Test Table",
	}
	err := handler.TableService.CreateTable(&table)
	assert.NoError(t, err)

	// Then create test field
	field := models.Field{
		TableID: table.ID,
		Name:    "Test Field",
		Key:     "test_field",
		Type:    models.FieldTypeText,
	}
	handler.Service.CreateField(&field)

	req, _ := http.NewRequest("DELETE", "/fields/"+field.ID.String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify field is deleted
	req, _ = http.NewRequest("GET", "/fields/"+field.ID.String(), nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestCreateField_WithoutKey(t *testing.T) {
	router, handler := setupTestRouter(t)

	// Create test table first
	table := models.Table{
		BaseID: uuid.New(),
		Name:   "Test Table",
	}
	err := handler.TableService.CreateTable(&table)
	assert.NoError(t, err)

	// Create field without key
	field := models.Field{
		TableID: table.ID,
		Name:    "Test Field",
		Type:    models.FieldTypeText,
	}

	jsonData, _ := json.Marshal(field)
	req, _ := http.NewRequest("POST", "/tables/"+table.ID.String()+"/fields", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response models.Field
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, field.Name, response.Name)
	assert.Equal(t, "test_field", response.Key)
}

func TestCreateField_InvalidTable(t *testing.T) {
	router, _ := setupTestRouter(t)

	// Try to create field with non-existent table
	field := models.Field{
		TableID: uuid.New(),
		Name:    "Test Field",
		Type:    models.FieldTypeText,
	}

	jsonData, _ := json.Marshal(field)
	req, _ := http.NewRequest("POST", "/tables/"+field.TableID.String()+"/fields", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestUpdateField_WithoutKey(t *testing.T) {
	router, handler := setupTestRouter(t)

	// Create test table first
	table := models.Table{
		BaseID: uuid.New(),
		Name:   "Test Table",
	}
	err := handler.TableService.CreateTable(&table)
	assert.NoError(t, err)

	// Create test field
	field := models.Field{
		TableID: table.ID,
		Name:    "Original Name",
		Key:     "original_name",
		Type:    models.FieldTypeText,
	}
	handler.Service.CreateField(&field)

	// Update field without key
	updatedField := models.Field{
		Name: "Updated Name",
		Type: models.FieldTypeNumber,
	}
	jsonData, _ := json.Marshal(updatedField)
	req, _ := http.NewRequest("PUT", "/fields/"+field.ID.String(), bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.Field
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, updatedField.Name, response.Name)
	assert.Equal(t, "updated_name", response.Key)
}

func TestUpdateFieldOrder_InvalidFields(t *testing.T) {
	router, handler := setupTestRouter(t)

	// Create test tables
	table1 := models.Table{
		BaseID: uuid.New(),
		Name:   "Table 1",
	}
	err := handler.TableService.CreateTable(&table1)
	assert.NoError(t, err)

	table2 := models.Table{
		BaseID: uuid.New(),
		Name:   "Table 2",
	}
	err = handler.TableService.CreateTable(&table2)
	assert.NoError(t, err)

	// Create fields in different tables
	field1 := models.Field{
		TableID: table1.ID,
		Name:    "Field 1",
		Key:     "field_1",
		Type:    models.FieldTypeText,
	}
	handler.Service.CreateField(&field1)

	field2 := models.Field{
		TableID: table2.ID,
		Name:    "Field 2",
		Key:     "field_2",
		Type:    models.FieldTypeText,
	}
	handler.Service.CreateField(&field2)

	// Try to update order with fields from different tables
	fieldOrders := map[uuid.UUID]int{
		field1.ID: 1,
		field2.ID: 0,
	}
	jsonData, _ := json.Marshal(fieldOrders)
	req, _ := http.NewRequest("PUT", "/tables/"+table1.ID.String()+"/fields/order", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDeleteField_NotFound(t *testing.T) {
	router, _ := setupTestRouter(t)

	// Try to delete non-existent field
	fieldID := uuid.New()
	req, _ := http.NewRequest("DELETE", "/fields/"+fieldID.String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestUpdateFieldOrder_Success(t *testing.T) {
	router, handler := setupTestRouter(t)

	// Create test table
	table := models.Table{
		BaseID: uuid.New(),
		Name:   "Test Table",
	}
	err := handler.TableService.CreateTable(&table)
	assert.NoError(t, err)

	// Create fields
	field1 := models.Field{
		TableID: table.ID,
		Name:    "Field 1",
		Key:     "field_1",
		Type:    models.FieldTypeText,
	}
	handler.Service.CreateField(&field1)

	field2 := models.Field{
		TableID: table.ID,
		Name:    "Field 2",
		Key:     "field_2",
		Type:    models.FieldTypeText,
	}
	handler.Service.CreateField(&field2)

	// Update field order
	fieldOrders := map[uuid.UUID]int{
		field1.ID: 1,
		field2.ID: 0,
	}
	jsonData, _ := json.Marshal(fieldOrders)
	req, _ := http.NewRequest("PUT", "/tables/"+table.ID.String()+"/fields/order", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify order update
	fields, err := handler.Service.GetFieldsByTableID(table.ID)
	assert.NoError(t, err)
	assert.Len(t, fields, 2)
	assert.Equal(t, field2.ID, fields[0].ID)
	assert.Equal(t, field1.ID, fields[1].ID)
}
