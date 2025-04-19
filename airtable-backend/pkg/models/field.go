package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// FieldType 定义字段类型
type FieldType string

const (
	FieldTypeText    FieldType = "text"
	FieldTypeNumber  FieldType = "number"
	FieldTypeBoolean FieldType = "boolean"
	FieldTypeDate    FieldType = "date"
	FieldTypeSelect  FieldType = "select"  // 新增：单选
	FieldTypeMulti   FieldType = "multi"   // 新增：多选
	FieldTypeFile    FieldType = "file"    // 新增：文件
	FieldTypeLink    FieldType = "link"    // 新增：关联
	FieldTypeFormula FieldType = "formula" // 新增：公式
	FieldTypeAuto    FieldType = "auto"    // 新增：自动编号
)

// ValidationRule 定义字段验证规则
type ValidationRule struct {
	Required    bool        `json:"required"`
	MinLength   int         `json:"minLength,omitempty"`
	MaxLength   int         `json:"maxLength,omitempty"`
	MinValue    float64     `json:"minValue,omitempty"`
	MaxValue    float64     `json:"maxValue,omitempty"`
	Pattern     string      `json:"pattern,omitempty"`
	CustomError string      `json:"customError,omitempty"`
	Options     []string    `json:"options,omitempty"` // 用于 select 和 multi 类型
	Default     interface{} `json:"default,omitempty"` // 默认值
}

// Field 表示表格中的字段
type Field struct {
	ID          uuid.UUID      `gorm:"type:uuid;primary_key" json:"id"`
	TableID     uuid.UUID      `gorm:"type:uuid;not null" json:"tableId"`
	Name        string         `gorm:"size:255;not null" json:"name"`
	Key         string         `gorm:"size:255;not null;uniqueIndex:idx_table_key" json:"key"`
	Type        FieldType      `gorm:"size:50;not null" json:"type"`
	Description string         `gorm:"size:500" json:"description"`
	Validation  ValidationRule `gorm:"type:jsonb" json:"validation"`
	Order       int            `gorm:"not null;default:0" json:"order"` // 新增：排序字段
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// BeforeCreate 在创建字段前设置默认值
func (f *Field) BeforeCreate(tx *gorm.DB) error {
	if f.ID == uuid.Nil {
		f.ID = uuid.New()
	}
	if f.Key == "" {
		f.Key = f.Name
	}
	return nil
}

// Validate 验证字段值是否符合规则
func (f *Field) Validate(value interface{}) error {
	// 检查必填
	if f.Validation.Required && value == nil {
		return &ValidationError{Message: "Field is required"}
	}

	// 根据字段类型进行验证
	switch f.Type {
	case FieldTypeText:
		if str, ok := value.(string); ok {
			if f.Validation.MinLength > 0 && len(str) < f.Validation.MinLength {
				return &ValidationError{Message: "Text is too short"}
			}
			if f.Validation.MaxLength > 0 && len(str) > f.Validation.MaxLength {
				return &ValidationError{Message: "Text is too long"}
			}
			if f.Validation.Pattern != "" {
				// TODO: 实现正则表达式验证
			}
		}
	case FieldTypeNumber:
		if num, ok := value.(float64); ok {
			if f.Validation.MinValue != 0 && num < f.Validation.MinValue {
				return &ValidationError{Message: "Number is too small"}
			}
			if f.Validation.MaxValue != 0 && num > f.Validation.MaxValue {
				return &ValidationError{Message: "Number is too large"}
			}
		}
	case FieldTypeSelect, FieldTypeMulti:
		if options := f.Validation.Options; len(options) > 0 {
			if f.Type == FieldTypeSelect {
				if str, ok := value.(string); ok {
					valid := false
					for _, option := range options {
						if str == option {
							valid = true
							break
						}
					}
					if !valid {
						return &ValidationError{Message: "Invalid option selected"}
					}
				}
			} else if f.Type == FieldTypeMulti {
				if arr, ok := value.([]string); ok {
					for _, val := range arr {
						valid := false
						for _, option := range options {
							if val == option {
								valid = true
								break
							}
						}
						if !valid {
							return &ValidationError{Message: "Invalid option selected"}
						}
					}
				}
			}
		}
	}

	return nil
}

// ValidationError 表示验证错误
type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}
