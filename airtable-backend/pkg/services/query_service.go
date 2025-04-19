package services

import (
	"airtable-backend/pkg/models"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type QueryService struct {
	db *gorm.DB
}

func NewQueryService(db *gorm.DB) *QueryService {
	return &QueryService{db: db}
}

// QueryRecords 执行高级查询
func (s *QueryService) QueryRecords(tableID uuid.UUID, params models.QueryParams) (*models.QueryResult, error) {
	var result models.QueryResult
	var total int64

	// 构建基础查询
	query := s.db.Model(&models.Record{}).Where("table_id = ?", tableID)

	// 应用过滤条件
	if err := s.applyFilters(query, params.Filters); err != nil {
		return nil, err
	}

	// 计算总数
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	// 应用排序
	if err := s.applySort(query, params.Sort); err != nil {
		return nil, err
	}

	// 应用分页
	if params.Page > 0 && params.PageSize > 0 {
		offset := (params.Page - 1) * params.PageSize
		query = query.Offset(offset).Limit(params.PageSize)
	}

	// 执行查询
	var records []models.Record
	if err := query.Find(&records).Error; err != nil {
		return nil, err
	}

	// 计算聚合结果
	aggregates, err := s.calculateAggregates(tableID, params)
	if err != nil {
		return nil, err
	}

	result = models.QueryResult{
		Records:    records,
		Total:      total,
		Page:       params.Page,
		PageSize:   params.PageSize,
		Aggregates: aggregates,
	}

	return &result, nil
}

// applyFilters 应用过滤条件
func (s *QueryService) applyFilters(query *gorm.DB, filters []models.FilterCondition) error {
	for _, filter := range filters {
		switch filter.Operator {
		case models.FilterEqual:
			query = query.Where(fmt.Sprintf("data->>'%s' = ?", filter.Field), filter.Value)
		case models.FilterNotEqual:
			query = query.Where(fmt.Sprintf("data->>'%s' != ?", filter.Field), filter.Value)
		case models.FilterGreaterThan:
			query = query.Where(fmt.Sprintf("data->>'%s' > ?", filter.Field), filter.Value)
		case models.FilterLessThan:
			query = query.Where(fmt.Sprintf("data->>'%s' < ?", filter.Field), filter.Value)
		case models.FilterGreaterEqual:
			query = query.Where(fmt.Sprintf("data->>'%s' >= ?", filter.Field), filter.Value)
		case models.FilterLessEqual:
			query = query.Where(fmt.Sprintf("data->>'%s' <= ?", filter.Field), filter.Value)
		case models.FilterContains:
			query = query.Where(fmt.Sprintf("data->>'%s' LIKE ?", filter.Field), fmt.Sprintf("%%%s%%", filter.Value))
		case models.FilterIn:
			query = query.Where(fmt.Sprintf("data->>'%s' IN ?", filter.Field), filter.Value)
		case models.FilterNotIn:
			query = query.Where(fmt.Sprintf("data->>'%s' NOT IN ?", filter.Field), filter.Value)
		case models.FilterIsNull:
			query = query.Where(fmt.Sprintf("data->>'%s' IS NULL", filter.Field))
		case models.FilterIsNotNull:
			query = query.Where(fmt.Sprintf("data->>'%s' IS NOT NULL", filter.Field))
		default:
			return fmt.Errorf("unsupported filter operator: %s", filter.Operator)
		}
	}
	return nil
}

// applySort 应用排序条件
func (s *QueryService) applySort(query *gorm.DB, sort []models.SortCondition) error {
	for _, s := range sort {
		direction := "ASC"
		if s.Direction == models.SortDesc {
			direction = "DESC"
		}
		query = query.Order(fmt.Sprintf("data->>'%s' %s", s.Field, direction))
	}
	return nil
}

// calculateAggregates 计算聚合结果
func (s *QueryService) calculateAggregates(tableID uuid.UUID, params models.QueryParams) (map[string]interface{}, error) {
	aggregates := make(map[string]interface{})

	for _, agg := range params.Aggregates {
		parts := strings.Split(agg, ":")
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid aggregate format: %s", agg)
		}

		function := parts[0]
		field := parts[1]

		var result interface{}
		var err error

		switch function {
		case string(models.AggregateCount):
			result, err = s.calculateCount(tableID, field)
		case string(models.AggregateSum):
			result, err = s.calculateSum(tableID, field)
		case string(models.AggregateAvg):
			result, err = s.calculateAvg(tableID, field)
		case string(models.AggregateMin):
			result, err = s.calculateMin(tableID, field)
		case string(models.AggregateMax):
			result, err = s.calculateMax(tableID, field)
		default:
			return nil, fmt.Errorf("unsupported aggregate function: %s", function)
		}

		if err != nil {
			return nil, err
		}

		aggregates[agg] = result
	}

	return aggregates, nil
}

// 聚合函数实现
func (s *QueryService) calculateCount(tableID uuid.UUID, field string) (int64, error) {
	var count int64
	err := s.db.Model(&models.Record{}).
		Where("table_id = ? AND data->>? IS NOT NULL", tableID, field).
		Count(&count).Error
	return count, err
}

func (s *QueryService) calculateSum(tableID uuid.UUID, field string) (float64, error) {
	var sum float64
	err := s.db.Model(&models.Record{}).
		Where("table_id = ?", tableID).
		Select("COALESCE(SUM((data->>?)::float), 0)", field).
		Scan(&sum).Error
	return sum, err
}

func (s *QueryService) calculateAvg(tableID uuid.UUID, field string) (float64, error) {
	var avg float64
	err := s.db.Model(&models.Record{}).
		Where("table_id = ?", tableID).
		Select("COALESCE(AVG((data->>?)::float), 0)", field).
		Scan(&avg).Error
	return avg, err
}

func (s *QueryService) calculateMin(tableID uuid.UUID, field string) (float64, error) {
	var min float64
	err := s.db.Model(&models.Record{}).
		Where("table_id = ?", tableID).
		Select("COALESCE(MIN((data->>?)::float), 0)", field).
		Scan(&min).Error
	return min, err
}

func (s *QueryService) calculateMax(tableID uuid.UUID, field string) (float64, error) {
	var max float64
	err := s.db.Model(&models.Record{}).
		Where("table_id = ?", tableID).
		Select("COALESCE(MAX((data->>?)::float), 0)", field).
		Scan(&max).Error
	return max, err
}
