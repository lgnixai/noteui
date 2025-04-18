package services

import (
	"github.com/google/uuid"
	"gorm.io/gorm"

	"airtable-backend/pkg/models"
)

type TableService struct {
	DB *gorm.DB
}

func NewTableService(db *gorm.DB) *TableService {
	return &TableService{DB: db}
}

func (s *TableService) CreateTable(table *models.Table) error {
	return s.DB.Create(table).Error
}

func (s *TableService) GetTableByID(id uuid.UUID) (*models.Table, error) {
	var table models.Table
	err := s.DB.Preload("Fields").First(&table, id).Error // Preload fields for easier access later
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // Table not found
		}
		return nil, err
	}
	return &table, nil
}

func (s *TableService) GetTablesByBaseID(baseID uuid.UUID) ([]models.Table, error) {
	var tables []models.Table
	err := s.DB.Where("base_id = ?", baseID).Find(&tables).Error
	return tables, err
}

func (s *TableService) UpdateTable(table *models.Table) error {
	return s.DB.Save(table).Error
}

func (s *TableService) DeleteTable(id uuid.UUID) error {
	// Consider cascade delete for fields and records
	// Or handle deletion of related objects here (Fields and Records first).
	// Delete associated fields
	if err := s.DB.Where("table_id = ?", id).Delete(&models.Field{}).Error; err != nil {
		return err
	}
	// Delete associated records
	if err := s.DB.Where("table_id = ?", id).Delete(&models.Record{}).Error; err != nil {
		return err
	}
	return s.DB.Delete(&models.Table{}, id).Error
}
