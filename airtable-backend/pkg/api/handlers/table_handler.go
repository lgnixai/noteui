package handlers

import (
	"airtable-backend/pkg/models"
	"airtable-backend/pkg/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TableHandler struct {
	Service     *services.TableService
	BaseService *services.BaseService
}

func NewTableHandler(s *services.TableService, bs *services.BaseService) *TableHandler {
	return &TableHandler{Service: s, BaseService: bs}
}

func (h *TableHandler) CreateTable(c *gin.Context) {
	baseIDStr := c.Param("baseId")
	baseID, err := uuid.Parse(baseIDStr)
	if err != nil {
		ErrorResponse(c, 400, "Invalid base ID format")
		return
	}

	base, err := h.BaseService.GetBaseByID(baseID)
	if err != nil {
		ErrorResponse(c, 500, "Failed to check if base exists")
		return
	}
	if base == nil {
		ErrorResponse(c, 404, "Base not found")
		return
	}

	var table models.Table
	if err := c.ShouldBindJSON(&table); err != nil {
		ErrorResponse(c, 400, "Invalid request payload")
		return
	}
	table.BaseID = baseID

	if err := h.Service.CreateTable(&table); err != nil {
		ErrorResponse(c, 500, "Failed to create table")
		return
	}

	JSONResponse(c, 201, table)
}

func (h *TableHandler) GetTable(c *gin.Context) {
	tableIDStr := c.Param("tableId")
	tableID, err := uuid.Parse(tableIDStr)
	if err != nil {
		ErrorResponse(c, 400, "Invalid table ID format")
		return
	}

	table, err := h.Service.GetTableByID(tableID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			ErrorResponse(c, 404, "Table not found")
			return
		}
		ErrorResponse(c, 500, "Failed to get table")
		return
	}
	if table == nil {
		ErrorResponse(c, 404, "Table not found")
		return
	}

	JSONResponse(c, 200, table)
}

func (h *TableHandler) GetTablesByBase(c *gin.Context) {
	baseIDStr := c.Param("baseId")
	baseID, err := uuid.Parse(baseIDStr)
	if err != nil {
		ErrorResponse(c, 400, "Invalid base ID format")
		return
	}

	base, err := h.BaseService.GetBaseByID(baseID)
	if err != nil {
		ErrorResponse(c, 500, "Failed to check if base exists")
		return
	}
	if base == nil {
		ErrorResponse(c, 404, "Base not found")
		return
	}

	tables, err := h.Service.GetTablesByBaseID(baseID)
	if err != nil {
		ErrorResponse(c, 500, "Failed to get tables for base")
		return
	}

	JSONResponse(c, 200, tables)
}

func (h *TableHandler) UpdateTable(c *gin.Context) {
	tableIDStr := c.Param("tableId")
	tableID, err := uuid.Parse(tableIDStr)
	if err != nil {
		ErrorResponse(c, 400, "Invalid table ID format")
		return
	}

	var table models.Table
	if err := c.ShouldBindJSON(&table); err != nil {
		ErrorResponse(c, 400, "Invalid request payload")
		return
	}
	table.ID = tableID

	existingTable, err := h.Service.GetTableByID(tableID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			ErrorResponse(c, 404, "Table not found")
			return
		}
		ErrorResponse(c, 500, "Failed to check existing table")
		return
	}
	if existingTable == nil {
		ErrorResponse(c, 404, "Table not found")
		return
	}

	existingTable.Name = table.Name

	if err := h.Service.UpdateTable(existingTable); err != nil {
		ErrorResponse(c, 500, "Failed to update table")
		return
	}

	JSONResponse(c, 200, existingTable)
}

func (h *TableHandler) DeleteTable(c *gin.Context) {
	tableIDStr := c.Param("tableId")
	tableID, err := uuid.Parse(tableIDStr)
	if err != nil {
		ErrorResponse(c, 400, "Invalid table ID format")
		return
	}

	existingTable, err := h.Service.GetTableByID(tableID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			ErrorResponse(c, 404, "Table not found")
			return
		}
		ErrorResponse(c, 500, "Failed to check existing table")
		return
	}
	if existingTable == nil {
		ErrorResponse(c, 404, "Table not found")
		return
	}

	if err := h.Service.DeleteTable(tableID); err != nil {
		ErrorResponse(c, 500, "Failed to delete table")
		return
	}

	c.Status(204)
}
