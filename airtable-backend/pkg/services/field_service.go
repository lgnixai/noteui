package services

import (
	"encoding/json"

	"airtable-backend/pkg/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type FieldService struct {
	db *gorm.DB
}

func NewFieldService(db *gorm.DB) *FieldService {
	return &FieldService{db: db}
}

// CreateField creates a new field
func (s *FieldService) CreateField(field *models.Field) error {
	// Marshal ValidationRule to JSON before saving
	validationJSON, err := json.Marshal(field.Validation)
	if err != nil {
		return err
	}

	// Create a temporary struct for saving
	type FieldForSave struct {
		ID          uuid.UUID `gorm:"type:uuid;primary_key"`
		TableID     uuid.UUID `gorm:"type:uuid"`
		Name        string
		Key         string
		Type        models.FieldType
		Description string
		Validation  string `gorm:"type:text"`
		Order       int
	}

	fieldForSave := FieldForSave{
		ID:          field.ID,
		TableID:     field.TableID,
		Name:        field.Name,
		Key:         field.Key,
		Type:        field.Type,
		Description: field.Description,
		Validation:  string(validationJSON),
		Order:       field.Order,
	}

	return s.db.Create(&fieldForSave).Error
}

// GetFieldByID retrieves a field by ID
func (s *FieldService) GetFieldByID(id uuid.UUID) (*models.Field, error) {
	var field models.Field
	if err := s.db.First(&field, id).Error; err != nil {
		return nil, err
	}
	return &field, nil
}

// GetFieldsByTableID retrieves all fields for a table
func (s *FieldService) GetFieldsByTableID(tableID uuid.UUID) ([]models.Field, error) {
	var fields []models.Field
	if err := s.db.Where("table_id = ?", tableID).Order("\"order\" asc").Find(&fields).Error; err != nil {
		return nil, err
	}
	return fields, nil
}

// UpdateField updates a field
func (s *FieldService) UpdateField(field *models.Field) error {
	// Marshal ValidationRule to JSON before saving
	validationJSON, err := json.Marshal(field.Validation)
	if err != nil {
		return err
	}

	// Create a temporary struct for saving
	type FieldForSave struct {
		ID          uuid.UUID `gorm:"type:uuid;primary_key"`
		TableID     uuid.UUID `gorm:"type:uuid"`
		Name        string
		Key         string
		Type        models.FieldType
		Description string
		Validation  string `gorm:"type:text"`
		Order       int
	}

	fieldForSave := FieldForSave{
		ID:          field.ID,
		TableID:     field.TableID,
		Name:        field.Name,
		Key:         field.Key,
		Type:        field.Type,
		Description: field.Description,
		Validation:  string(validationJSON),
		Order:       field.Order,
	}

	return s.db.Save(&fieldForSave).Error
}

// DeleteField deletes a field
func (s *FieldService) DeleteField(id uuid.UUID) error {
	return s.db.Delete(&models.Field{}, id).Error
}

// UpdateFieldOrder updates the order of fields
func (s *FieldService) UpdateFieldOrder(tableID uuid.UUID, fieldOrders map[uuid.UUID]int) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		for fieldID, order := range fieldOrders {
			if err := tx.Model(&models.Field{}).
				Where("id = ? AND table_id = ?", fieldID, tableID).
				Update("\"order\"", order).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// ValidateFieldValue 验证字段值
func (s *FieldService) ValidateFieldValue(fieldID uuid.UUID, value interface{}) error {
	field, err := s.GetFieldByID(fieldID)
	if err != nil {
		return err
	}
	return field.Validate(value)
}

// 辅助函数

func isValidFieldType(fieldType models.FieldType) bool {
	switch fieldType {
	case models.FieldTypeText,
		models.FieldTypeNumber,
		models.FieldTypeBoolean,
		models.FieldTypeDate,
		models.FieldTypeSelect,
		models.FieldTypeMulti,
		models.FieldTypeFile,
		models.FieldTypeLink,
		models.FieldTypeFormula,
		models.FieldTypeAuto:
		return true
	default:
		return false
	}
}

func getDefaultValueForType(fieldType models.FieldType) interface{} {
	switch fieldType {
	case models.FieldTypeText:
		return ""
	case models.FieldTypeNumber:
		return 0
	case models.FieldTypeBoolean:
		return false
	case models.FieldTypeDate:
		return nil
	case models.FieldTypeSelect:
		return ""
	case models.FieldTypeMulti:
		return []string{}
	case models.FieldTypeFile:
		return nil
	case models.FieldTypeLink:
		return nil
	case models.FieldTypeFormula:
		return ""
	case models.FieldTypeAuto:
		return 0
	default:
		return nil
	}
}
