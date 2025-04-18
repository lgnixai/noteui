package services

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"airtable-backend/pkg/models"
	"airtable-backend/pkg/query"
	"airtable-backend/pkg/redis"     // Import redis package to use redis.Publish
	"airtable-backend/pkg/websocket" // Import websocket package to access manager methods
)

// Define WebSocket message format for record updates
type RecordUpdateMessage struct {
	Type     string         `json:"type"` // e.g., "record_created", "record_updated", "record_deleted"
	TableID  uuid.UUID      `json:"tableId"`
	RecordID uuid.UUID      `json:"recordId"`
	Record   *models.Record `json:"record,omitempty"` // Full record for created/updated
}

type RecordService struct {
	DB *gorm.DB
	// Removed: RedisPub    *redis.Publisher // This field is not needed
	WSManager    *websocket.Manager // Need access to the WS Manager to broadcast
	FieldService *FieldService      // Dependency to get table fields
}

// NewRecordService initializes the RecordService.
// We don't pass the Redis publisher explicitly, as the redis.Publish function
// uses the globally initialized redis.RDB client.
func NewRecordService(db *gorm.DB, wsManager *websocket.Manager, fieldService *FieldService) *RecordService {
	return &RecordService{
		DB:           db,
		WSManager:    wsManager,
		FieldService: fieldService,
	}
}

// transformRecordData converts field IDs to field names in the record data
func (s *RecordService) transformRecordData(record *models.Record, fieldMap map[string]models.Field) error {
	var dataMap map[string]interface{}
	if err := json.Unmarshal(record.Data, &dataMap); err != nil {
		return fmt.Errorf("failed to unmarshal record data: %w", err)
	}

	// 处理嵌套的 data 结构
	if nestedData, ok := dataMap["data"].(map[string]interface{}); ok {
		transformedData := make(map[string]interface{})
		for fieldID, value := range nestedData {
			if field, exists := fieldMap[fieldID]; exists {
				transformedData[field.KeyName] = value
			} else {
				transformedData[fieldID] = value
			}
		}
		dataMap["data"] = transformedData
	}

	// 如果数据为空，返回空对象而不是 null
	if len(dataMap) == 0 {
		record.Data = []byte("{}")
		return nil
	}

	transformedJSON, err := json.Marshal(dataMap)
	if err != nil {
		return fmt.Errorf("failed to marshal transformed data: %w", err)
	}

	record.Data = transformedJSON
	return nil
}

// GetRecords retrieves records for a table with filtering, sorting, and pagination.
func (s *RecordService) GetRecords(
	tableID uuid.UUID,
	filterJSON []byte, // Raw JSON bytes for filter
	sortJSON []byte, // Raw JSON bytes for sort
	limit int,
	offset int,
) ([]models.Record, int64, error) {
	var records []models.Record
	var total int64

	// Fetch fields for the table first to build queries correctly
	fields, err := s.FieldService.GetFieldsByTableID(tableID)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get fields for table %s: %w", tableID, err)
	}
	// Create a map for easy lookup by FieldID
	fieldMap := make(map[string]models.Field)
	for _, field := range fields {
		fieldMap[field.ID.String()] = field
	}

	dbQuery := s.DB.Model(&models.Record{}).Where("table_id = ?", tableID)

	// Apply filter
	if len(filterJSON) > 0 {
		filterGroup, err := query.ParseFilterJSON(filterJSON)
		if err != nil {
			return nil, 0, fmt.Errorf("invalid filter json: %w", err)
		}
		if filterGroup != nil {
			dbQuery, err = query.BuildGormFilter(dbQuery, fieldMap, filterGroup)
			if err != nil {
				return nil, 0, fmt.Errorf("failed to build filter query: %w", err)
			}
		}
	}

	// Count total records *before* applying limit/offset
	if err := dbQuery.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count records: %w", err)
	}

	// Apply sort
	if len(sortJSON) > 0 {
		sorts, err := query.ParseSortJSON(sortJSON)
		if err != nil {
			return nil, 0, fmt.Errorf("invalid sort json: %w", err)
		}
		if len(sorts) > 0 {
			dbQuery, err = query.BuildGormSort(dbQuery, fieldMap, sorts)
			if err != nil {
				return nil, 0, fmt.Errorf("failed to build sort query: %w", err)
			}
		}
	}

	// Apply pagination
	if limit > 0 {
		dbQuery = dbQuery.Limit(limit)
	}
	if offset > 0 {
		dbQuery = dbQuery.Offset(offset)
	}

	// Execute the query with preloading
	if err := dbQuery.Preload("Table.Fields").Find(&records).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to retrieve records: %w", err)
	}

	// Transform the data for each record
	for i := range records {
		if err := s.transformRecordData(&records[i], fieldMap); err != nil {
			return nil, 0, fmt.Errorf("failed to transform record data: %w", err)
		}
	}

	return records, total, nil
}

// CreateRecord creates a new record. Data should be JSON corresponding to fields.
// ... (CreateRecord function remains the same) ...
func (s *RecordService) CreateRecord(tableID uuid.UUID, data json.RawMessage) (*models.Record, error) {
	// Optional: Validate data against table fields schema here
	// (e.g., check if keys in data match field key_names, check value types)

	record := models.Record{
		TableID: tableID,
		Data:    data,
	}

	if err := s.DB.Create(&record).Error; err != nil {
		return nil, fmt.Errorf("failed to create record: %w", err)
	}

	// Publish update to Redis
	message := RecordUpdateMessage{
		Type:     "record_created",
		TableID:  tableID,
		RecordID: record.ID,
		Record:   &record,
	}
	s.publishUpdate(tableID, message) // This calls the helper

	return &record, nil
}

// GetRecordByID retrieves a single record by its ID.
func (s *RecordService) GetRecordByID(id uuid.UUID) (*models.Record, error) {
	var record models.Record
	err := s.DB.Preload("Table.Fields").First(&record, id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // Record not found
		}
		return nil, fmt.Errorf("failed to get record by ID: %w", err)
	}

	// Get fields for the table
	fields, err := s.FieldService.GetFieldsByTableID(record.TableID)
	if err != nil {
		return nil, fmt.Errorf("failed to get fields for table %s: %w", record.TableID, err)
	}

	// Create a map for easy lookup by FieldID
	fieldMap := make(map[string]models.Field)
	for _, field := range fields {
		fieldMap[field.ID.String()] = field
	}

	// Transform the data
	if err := s.transformRecordData(&record, fieldMap); err != nil {
		return nil, fmt.Errorf("failed to transform record data: %w", err)
	}

	return &record, nil
}

// UpdateRecord updates an existing record. Data should be JSON with updated fields.
// It will merge the provided data with the existing JSONB data.
// ... (UpdateRecord function remains the same) ...
func (s *RecordService) UpdateRecord(id uuid.UUID, newData json.RawMessage) (*models.Record, error) {
	existingRecord, err := s.GetRecordByID(id)
	if err != nil {
		return nil, err // Propagate not found error etc.
	}
	if existingRecord == nil {
		return nil, fmt.Errorf("record with ID %s not found", id)
	}

	// Unmarshal existing data
	var existingMap map[string]json.RawMessage
	if err := json.Unmarshal(existingRecord.Data, &existingMap); err != nil {
		return nil, fmt.Errorf("failed to unmarshal existing record data: %w", err)
	}

	// Unmarshal new data
	var newMap map[string]json.RawMessage
	if err := json.Unmarshal(newData, &newMap); err != nil {
		return nil, fmt.Errorf("failed to unmarshal new record data: %w", err)
	}

	// Optional: Validate newData keys against field key_names for the record's table
	// This prevents adding arbitrary keys to the JSONB if not defined as fields.
	// Fetch table fields...
	// fieldMap, err := s.getFieldMapByTableID(existingRecord.TableID) // Need a helper
	// if err != nil { /* handle error */ }
	// for key := range newMap {
	//    // Check if key exists in fieldMap as a KeyName
	//    found := false
	//    for _, field := range fieldMap {
	//        if field.KeyName == key {
	//            found = true
	//            break
	//        }
	//    }
	//    if !found {
	//       return nil, fmt.Errorf("invalid key '%s' in update data for table %s", key, existingRecord.TableID)
	//    }
	// }

	// Merge new data into existing data
	for key, value := range newMap {
		existingMap[key] = value
	}

	// Marshal merged data back to JSON
	mergedData, err := json.Marshal(existingMap)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal merged record data: %w", err)
	}

	// Update record in DB
	existingRecord.Data = mergedData
	if err := s.DB.Save(existingRecord).Error; err != nil {
		return nil, fmt.Errorf("failed to update record: %w", err)
	}

	// Publish update to Redis
	message := RecordUpdateMessage{
		Type:     "record_updated",
		TableID:  existingRecord.TableID,
		RecordID: existingRecord.ID,
		Record:   existingRecord, // Send updated record
	}
	s.publishUpdate(existingRecord.TableID, message) // This calls the helper

	return existingRecord, nil
}

// DeleteRecord deletes a record by its ID.
// ... (DeleteRecord function remains the same) ...
func (s *RecordService) DeleteRecord(id uuid.UUID) error {
	recordToDelete, err := s.GetRecordByID(id)
	if err != nil {
		return err // Propagate error
	}
	if recordToDelete == nil {
		return fmt.Errorf("record with ID %s not found", id)
	}

	if err := s.DB.Delete(&models.Record{}, id).Error; err != nil {
		return fmt.Errorf("failed to delete record: %w", err)
	}

	// Publish update to Redis
	message := RecordUpdateMessage{
		Type:     "record_deleted",
		TableID:  recordToDelete.TableID,
		RecordID: recordToDelete.ID,
		// Record is nil for delete message
	}
	s.publishUpdate(recordToDelete.TableID, message) // This calls the helper

	return nil
}

// publishUpdate marshals the message and publishes it to the correct Redis channel.
// This helper function correctly calls the package-level redis.Publish function.
func (s *RecordService) publishUpdate(tableID uuid.UUID, message RecordUpdateMessage) {
	channel := fmt.Sprintf("table_updates:%s", tableID.String())
	messageBytes, err := json.Marshal(message)
	if err != nil {
		log.Printf("Failed to marshal websocket message for table %s: %v", tableID, err)
		return
	}
	// Correctly call the package-level function
	redis.Publish(channel, string(messageBytes)) // Publish as string
}
