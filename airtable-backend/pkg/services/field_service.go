package services

import (
	"airtable-backend/pkg/models"

	`fmt`

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type FieldService struct {
	DB *gorm.DB
}

func NewFieldService(db *gorm.DB) *FieldService {
	return &FieldService{DB: db}
}

func (s *FieldService) CreateField(field *models.Field) error {
	// Check if a field with the same name or key_name already exists in this table
	var existingField models.Field
	err := s.DB.Where("table_id = ? AND (name = ? OR key_name = ?)", field.TableID, field.Name, field.KeyName).First(&existingField).Error
	if err == nil {
		// Record found, name or key_name is not unique in this table
		return fmt.Errorf("field with name '%s' or key_name '%s' already exists in this table", field.Name, field.KeyName)
	}
	if err != gorm.ErrRecordNotFound {
		// Some other database error
		return fmt.Errorf("database error checking for existing field: %w", err)
	}
	// No existing field with same name or key_name, proceed to create
	return s.DB.Create(field).Error
}

func (s *FieldService) GetFieldByID(id uuid.UUID) (*models.Field, error) {
	var field models.Field
	err := s.DB.First(&field, id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // Field not found
		}
		return nil, err
	}
	return &field, nil
}

func (s *FieldService) GetFieldsByTableID(tableID uuid.UUID) ([]models.Field, error) {
	var fields []models.Field
	err := s.DB.Where("table_id = ?", tableID).Find(&fields).Error
	return fields, err
}

func (s *FieldService) UpdateField(field *models.Field) error {
	// When updating a field, changing the KeyName would break existing records.
	// Changing Type is also complex as it affects interpretation of existing JSONB data.
	// A real system would need careful migration logic here.
	// For this example, we allow updating Name and Type, but not KeyName directly via this method.
	// GORM's Save will update all fields including KeyName if present in struct,
	// so be cautious. It's better to fetch the existing field and only update allowed fields.

	existingField := &models.Field{}
	err := s.DB.First(existingField, field.ID).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("field with ID %s not found", field.ID)
		}
		return fmt.Errorf("error fetching field for update: %w", err)
	}

	// Check if new name or keyname conflicts (only if name/keyname is actually being changed)
	if existingField.TableID != field.TableID || existingField.Name != field.Name || existingField.KeyName != field.KeyName {
		var conflictField models.Field
		query := s.DB.Where("table_id = ?", field.TableID).Where("id != ?", field.ID)
		query = query.Where("name = ? OR key_name = ?", field.Name, field.KeyName)
		err = query.First(&conflictField).Error

		if err == nil {
			// Conflict found
			return fmt.Errorf("field with name '%s' or key_name '%s' already exists in this table", field.Name, field.KeyName)
		}
		if err != gorm.ErrRecordNotFound {
			return fmt.Errorf("database error checking for update conflict: %w", err)
		}
	}

	// Update only allowed fields (Name, Type)
	// If you want to update more, add them here, but be careful with KeyName.
	existingField.Name = field.Name
	existingField.Type = field.Type
	// existingField.KeyName should NOT be updated via Save unless specific logic is added

	return s.DB.Save(existingField).Error
}

func (s *FieldService) DeleteField(id uuid.UUID) error {
	// Deleting a field means records will still have the data in JSONB, but it should be ignored.
	// A more robust solution might remove the key from all records' JSONB, which is expensive.
	// For this example, we just delete the field definition. The query logic needs to be robust
	// enough to handle querying with field IDs that no longer exist (or ignore them).

	return s.DB.Delete(&models.Field{}, id).Error
}
