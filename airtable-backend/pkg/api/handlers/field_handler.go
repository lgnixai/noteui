package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"airtable-backend/pkg/models"
	"airtable-backend/pkg/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type FieldHandler struct {
	Service      *services.FieldService
	TableService *services.TableService
}

func NewFieldHandler(s *services.FieldService, ts *services.TableService) *FieldHandler {
	return &FieldHandler{Service: s, TableService: ts}
}

func (h *FieldHandler) CreateField(c *gin.Context) {
	var field models.Field
	if err := c.ShouldBindJSON(&field); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	tableID, err := uuid.Parse(c.Param("tableId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid table ID"})
		return
	}
	field.TableID = tableID

	// 如果没有提供Key，则使用Name的小写形式作为Key
	if field.Key == "" {
		field.Key = strings.ToLower(strings.ReplaceAll(field.Name, " ", "_"))
	}

	switch field.Type {
	case models.FieldTypeText, models.FieldTypeNumber, models.FieldTypeBoolean, models.FieldTypeDate:
		// Valid type
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Unsupported field type: %s", field.Type)})
		return
	}

	if err := h.Service.CreateField(&field); err != nil {
		if strings.Contains(err.Error(), "unique constraint") {
			c.JSON(http.StatusConflict, gin.H{"error": "Field with this name or key already exists in this table"})
			return
		}
		if err.Error() == "table not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Table not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create field"})
		return
	}

	c.JSON(http.StatusCreated, field)
}

func (h *FieldHandler) GetField(c *gin.Context) {
	fieldID, err := uuid.Parse(c.Param("fieldId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid field ID"})
		return
	}

	field, err := h.Service.GetFieldByID(fieldID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Field not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get field"})
		return
	}
	if field == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Field not found"})
		return
	}

	c.JSON(http.StatusOK, field)
}

func (h *FieldHandler) GetFieldsByTable(c *gin.Context) {
	tableID, err := uuid.Parse(c.Param("tableId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid table ID"})
		return
	}

	table, err := h.TableService.GetTableByID(tableID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check if table exists"})
		return
	}
	if table == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Table not found"})
		return
	}

	fields, err := h.Service.GetFieldsByTableID(tableID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get fields for table"})
		return
	}

	c.JSON(http.StatusOK, fields)
}

func (h *FieldHandler) UpdateField(c *gin.Context) {
	fieldID, err := uuid.Parse(c.Param("fieldId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid field ID"})
		return
	}

	var field models.Field
	if err := c.ShouldBindJSON(&field); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	field.ID = fieldID

	existingField, err := h.Service.GetFieldByID(fieldID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Field not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check existing field"})
		return
	}
	if existingField == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Field not found"})
		return
	}

	switch field.Type {
	case models.FieldTypeText, models.FieldTypeNumber, models.FieldTypeBoolean, models.FieldTypeDate:
		// Valid type
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Unsupported field type: %s", field.Type)})
		return
	}

	// 如果没有提供Key，则使用Name的小写形式作为Key
	if field.Key == "" {
		field.Key = strings.ToLower(strings.ReplaceAll(field.Name, " ", "_"))
	}

	existingField.Name = field.Name
	existingField.Key = field.Key
	existingField.Type = field.Type

	if err := h.Service.UpdateField(existingField); err != nil {
		if strings.Contains(err.Error(), "unique constraint") {
			c.JSON(http.StatusConflict, gin.H{"error": "Field with this name or key already exists in this table"})
			return
		}
		if err.Error() == "table not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Table not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update field"})
		return
	}

	c.JSON(http.StatusOK, existingField)
}

func (h *FieldHandler) DeleteField(c *gin.Context) {
	fieldID, err := uuid.Parse(c.Param("fieldId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid field ID"})
		return
	}

	// 先检查字段是否存在
	existingField, err := h.Service.GetFieldByID(fieldID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Field not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check existing field"})
		return
	}
	if existingField == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Field not found"})
		return
	}

	if err := h.Service.DeleteField(fieldID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete field"})
		return
	}

	c.Status(http.StatusOK)
}

func (h *FieldHandler) UpdateFieldOrder(c *gin.Context) {
	tableID, err := uuid.Parse(c.Param("tableId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid table ID"})
		return
	}

	// 先检查表是否存在
	table, err := h.TableService.GetTableByID(tableID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check if table exists"})
		return
	}
	if table == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Table not found"})
		return
	}

	var fieldOrders map[uuid.UUID]int
	if err := c.ShouldBindJSON(&fieldOrders); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// 验证所有字段是否都属于同一个表
	for fieldID := range fieldOrders {
		field, err := h.Service.GetFieldByID(fieldID)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("Field %s not found", fieldID)})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check field"})
			return
		}
		if field.TableID != tableID {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Field %s does not belong to table %s", fieldID, tableID)})
			return
		}
	}

	if err := h.Service.UpdateFieldOrder(tableID, fieldOrders); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusOK)
}

func (h *FieldHandler) ValidateFieldValue(c *gin.Context) {
	fieldID, err := uuid.Parse(c.Param("fieldId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid field ID"})
		return
	}

	var value interface{}
	if err := c.ShouldBindJSON(&value); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if err := h.Service.ValidateFieldValue(fieldID, value); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusOK)
}
