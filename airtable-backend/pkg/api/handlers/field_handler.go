package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	`github.com/davecgh/go-spew/spew`

	"airtable-backend/pkg/models"
	"airtable-backend/pkg/services"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

type FieldHandler struct {
	Service      *services.FieldService
	TableService *services.TableService // Need TableService to check if table exists
}

func NewFieldHandler(s *services.FieldService, ts *services.TableService) *FieldHandler {
	return &FieldHandler{Service: s, TableService: ts}
}

func (h *FieldHandler) CreateField(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tableIDStr := vars["tableId"]
	tableID, err := uuid.Parse(tableIDStr)
	if err != nil {
		ErrorResponse(w, http.StatusBadRequest, "Invalid table ID format")
		return
	}

	// Check if the table exists
	table, err := h.TableService.GetTableByID(tableID)
	if err != nil {
		ErrorResponse(w, http.StatusInternalServerError, "Failed to check if table exists")
		return
	}
	if table == nil {
		ErrorResponse(w, http.StatusNotFound, "Table not found")
		return
	}

	var field models.Field
	if err := json.NewDecoder(r.Body).Decode(&field); err != nil {
		ErrorResponse(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	field.TableID = tableID // Associate field with the table from the URL

	fmt.Println("field.Type")
	spew.Dump(field)

	// Validate field type (optional but good practice)
	switch field.Type {
	case models.FieldTypeText, models.FieldNumber, models.TypeBoolean, models.TypeDate:
		// Valid type
	default:
		ErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("Unsupported field type: %s", field.Type))
		return
	}

	if err := h.Service.CreateField(&field); err != nil {
		// Check for unique constraint error specifically if needed
		if strings.Contains(err.Error(), "unique constraint") {
			ErrorResponse(w, http.StatusConflict, "Field with this name or key already exists in this table")
			return
		}
		ErrorResponse(w, http.StatusInternalServerError, "Failed to create field")
		return
	}

	JSONResponse(w, http.StatusCreated, field)
}

func (h *FieldHandler) GetField(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	fieldIDStr := vars["fieldId"]
	fieldID, err := uuid.Parse(fieldIDStr)
	if err != nil {
		ErrorResponse(w, http.StatusBadRequest, "Invalid field ID format")
		return
	}

	field, err := h.Service.GetFieldByID(fieldID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			ErrorResponse(w, http.StatusNotFound, "Field not found")
			return
		}
		ErrorResponse(w, http.StatusInternalServerError, "Failed to get field")
		return
	}
	if field == nil {
		ErrorResponse(w, http.StatusNotFound, "Field not found")
		return
	}

	// Optional: Validate tableId/baseId from URL if needed

	JSONResponse(w, http.StatusOK, field)
}

func (h *FieldHandler) GetFieldsByTable(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tableIDStr := vars["tableId"]
	tableID, err := uuid.Parse(tableIDStr)
	if err != nil {
		ErrorResponse(w, http.StatusBadRequest, "Invalid table ID format")
		return
	}

	// Check if the table exists
	table, err := h.TableService.GetTableByID(tableID)
	if err != nil {
		ErrorResponse(w, http.StatusInternalServerError, "Failed to check if table exists")
		return
	}
	if table == nil {
		ErrorResponse(w, http.StatusNotFound, "Table not found")
		return
	}

	fields, err := h.Service.GetFieldsByTableID(tableID)
	if err != nil {
		ErrorResponse(w, http.StatusInternalServerError, "Failed to get fields for table")
		return
	}

	JSONResponse(w, http.StatusOK, fields)
}

func (h *FieldHandler) UpdateField(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	fieldIDStr := vars["fieldId"]
	fieldID, err := uuid.Parse(fieldIDStr)
	if err != nil {
		ErrorResponse(w, http.StatusBadRequest, "Invalid field ID format")
		return
	}

	var field models.Field
	if err := json.NewDecoder(r.Body).Decode(&field); err != nil {
		ErrorResponse(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	field.ID = fieldID // Ensure ID from URL is used

	// Optional: Check if field exists
	existingField, err := h.Service.GetFieldByID(fieldID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			ErrorResponse(w, http.StatusNotFound, "Field not found")
			return
		}
		ErrorResponse(w, http.StatusInternalServerError, "Failed to check existing field")
		return
	}
	if existingField == nil {
		ErrorResponse(w, http.StatusNotFound, "Field not found")
		return
	}

	// Validate field type
	switch field.Type {
	case models.FieldTypeText, models.FieldNumber, models.TypeBoolean, models.TypeDate:
		// Valid type
	default:
		ErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("Unsupported field type: %s", field.Type))
		return
	}

	// Update allowed fields (Name, Type). Do NOT update KeyName via this method.
	existingField.Name = field.Name
	existingField.Type = field.Type

	if err := h.Service.UpdateField(existingField); err != nil {
		if strings.Contains(err.Error(), "unique constraint") {
			ErrorResponse(w, http.StatusConflict, "Field with this name or key already exists in this table")
			return
		}
		ErrorResponse(w, http.StatusInternalServerError, "Failed to update field")
		return
	}

	JSONResponse(w, http.StatusOK, existingField)
}

func (h *FieldHandler) DeleteField(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	fieldIDStr := vars["fieldId"]
	fieldID, err := uuid.Parse(fieldIDStr)
	if err != nil {
		ErrorResponse(w, http.StatusBadRequest, "Invalid field ID format")
		return
	}

	// Optional: Check if field exists before delete
	existingField, err := h.Service.GetFieldByID(fieldID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			ErrorResponse(w, http.StatusNotFound, "Field not found")
			return
		}
		ErrorResponse(w, http.StatusInternalServerError, "Failed to check existing field")
		return
	}
	if existingField == nil {
		ErrorResponse(w, http.StatusNotFound, "Field not found")
		return
	}

	if err := h.Service.DeleteField(fieldID); err != nil {
		ErrorResponse(w, http.StatusInternalServerError, "Failed to delete field")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
