package services

import (
	"github.com/google/uuid"
	"gorm.io/gorm"

	"airtable-backend/pkg/models"
)

type BaseService struct {
	DB *gorm.DB
}

func NewBaseService(db *gorm.DB) *BaseService {
	return &BaseService{DB: db}
}

func (s *BaseService) CreateBase(base *models.Base) error {
	return s.DB.Create(base).Error
}

func (s *BaseService) GetBaseByID(id uuid.UUID) (*models.Base, error) {
	var base models.Base
	err := s.DB.First(&base, id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // Base not found
		}
		return nil, err
	}
	return &base, nil
}

func (s *BaseService) GetAllBases() ([]models.Base, error) {
	var bases []models.Base
	err := s.DB.Find(&bases).Error
	return bases, err
}

func (s *BaseService) UpdateBase(base *models.Base) error {
	return s.DB.Save(base).Error
}

func (s *BaseService) DeleteBase(id uuid.UUID) error {
	// Consider adding cascade delete for tables, fields, records in the DB schema
	// Or handle deletion of related objects here.
	return s.DB.Delete(&models.Base{}, id).Error
}
