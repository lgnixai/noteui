package handlers

import (
	"airtable-backend/pkg/models"
	"airtable-backend/pkg/services"
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

type TableHandler struct {
	Service     *services.TableService
	BaseService *services.BaseService // Need BaseService to check if base exists
}

func NewTableHandler(s *services.TableService, bs *services.BaseService) *TableHandler {
	return &TableHandler{Service: s, BaseService: bs}
}

func (h *TableHandler) CreateTable(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	baseIDStr := vars["baseId"]
	baseID, err := uuid.Parse(baseIDStr)
	if err != nil {
		ErrorResponse(w, http.StatusBadRequest, "Invalid base ID format")
		return
	}

	// Check if the base exists
	base, err := h.BaseService.GetBaseByID(baseID)
	if err != nil {
		ErrorResponse(w, http.StatusInternalServerError, "Failed to check if base exists")
		return
	}
	if base == nil {
		ErrorResponse(w, http.StatusNotFound, "Base not found")
		return
	}

	var table models.Table
	if err := json.NewDecoder(r.Body).Decode(&table); err != nil {
		ErrorResponse(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	table.BaseID = baseID // Associate table with the base from the URL

	if err := h.Service.CreateTable(&table); err != nil {
		ErrorResponse(w, http.StatusInternalServerError, "Failed to create table")
		return
	}

	JSONResponse(w, http.StatusCreated, table)
}

func (h *TableHandler) GetTable(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tableIDStr := vars["tableId"]
	tableID, err := uuid.Parse(tableIDStr)
	if err != nil {
		ErrorResponse(w, http.StatusBadRequest, "Invalid table ID format")
		return
	}

	table, err := h.Service.GetTableByID(tableID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			ErrorResponse(w, http.StatusNotFound, "Table not found")
			return
		}
		ErrorResponse(w, http.StatusInternalServerError, "Failed to get table")
		return
	}
	if table == nil {
		ErrorResponse(w, http.StatusNotFound, "Table not found")
		return
	}

	// Optional: Validate baseId from URL if needed, e.g., vars["baseId"] vs table.BaseID
	// For simplicity, assuming tableId is globally unique enough for this example.

	JSONResponse(w, http.StatusOK, table)
}

func (h *TableHandler) GetTablesByBase(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	baseIDStr := vars["baseId"]
	baseID, err := uuid.Parse(baseIDStr)
	if err != nil {
		ErrorResponse(w, http.StatusBadRequest, "Invalid base ID format")
		return
	}

	// Check if the base exists
	base, err := h.BaseService.GetBaseByID(baseID)
	if err != nil {
		ErrorResponse(w, http.StatusInternalServerError, "Failed to check if base exists")
		return
	}
	if base == nil {
		ErrorResponse(w, http.StatusNotFound, "Base not found")
		return
	}

	tables, err := h.Service.GetTablesByBaseID(baseID)
	if err != nil {
		ErrorResponse(w, http.StatusInternalServerError, "Failed to get tables for base")
		return
	}

	JSONResponse(w, http.StatusOK, tables)
}

func (h *TableHandler) UpdateTable(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tableIDStr := vars["tableId"]
	tableID, err := uuid.Parse(tableIDStr)
	if err != nil {
		ErrorResponse(w, http.StatusBadRequest, "Invalid table ID format")
		return
	}

	var table models.Table
	if err := json.NewDecoder(r.Body).Decode(&table); err != nil {
		ErrorResponse(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	table.ID = tableID // Ensure the ID from the URL is used

	// Optional: Check if table exists
	existingTable, err := h.Service.GetTableByID(tableID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			ErrorResponse(w, http.StatusNotFound, "Table not found")
			return
		}
		ErrorResponse(w, http.StatusInternalServerError, "Failed to check existing table")
		return
	}
	if existingTable == nil {
		ErrorResponse(w, http.StatusNotFound, "Table not found")
		return
	}

	// Update allowed fields
	existingTable.Name = table.Name
	// existingTable.BaseID = ... // Don't allow moving tables between bases here

	if err := h.Service.UpdateTable(existingTable); err != nil {
		ErrorResponse(w, http.StatusInternalServerError, "Failed to update table")
		return
	}

	JSONResponse(w, http.StatusOK, existingTable)
}

func (h *TableHandler) DeleteTable(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tableIDStr := vars["tableId"]
	tableID, err := uuid.Parse(tableIDStr)
	if err != nil {
		ErrorResponse(w, http.StatusBadRequest, "Invalid table ID format")
		return
	}

	// Optional: Check if table exists before delete
	existingTable, err := h.Service.GetTableByID(tableID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			ErrorResponse(w, http.StatusNotFound, "Table not found")
			return
		}
		ErrorResponse(w, http.StatusInternalServerError, "Failed to check existing table")
		return
	}
	if existingTable == nil {
		ErrorResponse(w, http.StatusNotFound, "Table not found")
		return
	}

	if err := h.Service.DeleteTable(tableID); err != nil {
		ErrorResponse(w, http.StatusInternalServerError, "Failed to delete table")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
