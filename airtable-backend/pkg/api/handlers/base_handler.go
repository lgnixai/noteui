package handlers

import (
	"airtable-backend/pkg/models"
	"airtable-backend/pkg/services"
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"gorm.io/gorm" // Check for gorm.ErrRecordNotFound
)

type BaseHandler struct {
	Service *services.BaseService
}

func NewBaseHandler(s *services.BaseService) *BaseHandler {
	return &BaseHandler{Service: s}
}

func (h *BaseHandler) CreateBase(w http.ResponseWriter, r *http.Request) {
	var base models.Base
	if err := json.NewDecoder(r.Body).Decode(&base); err != nil {
		ErrorResponse(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// In a real app, get UserID from authenticated user
	// For this example, let's assign a dummy UserID or omit if not strictly necessary for Base creation
	// base.UserID = ...

	if err := h.Service.CreateBase(&base); err != nil {
		ErrorResponse(w, http.StatusInternalServerError, "Failed to create base")
		return
	}

	JSONResponse(w, http.StatusCreated, base)
}

func (h *BaseHandler) GetBase(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["baseId"]
	id, err := uuid.Parse(idStr)
	if err != nil {
		ErrorResponse(w, http.StatusBadRequest, "Invalid base ID format")
		return
	}

	base, err := h.Service.GetBaseByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			ErrorResponse(w, http.StatusNotFound, "Base not found")
			return
		}
		ErrorResponse(w, http.StatusInternalServerError, "Failed to get base")
		return
	}
	if base == nil { // Service might return nil for not found without gorm.ErrRecordNotFound
		ErrorResponse(w, http.StatusNotFound, "Base not found")
		return
	}

	JSONResponse(w, http.StatusOK, base)
}

func (h *BaseHandler) GetAllBases(w http.ResponseWriter, r *http.Request) {
	bases, err := h.Service.GetAllBases()
	if err != nil {
		ErrorResponse(w, http.StatusInternalServerError, "Failed to get bases")
		return
	}

	JSONResponse(w, http.StatusOK, bases)
}

func (h *BaseHandler) UpdateBase(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["baseId"]
	id, err := uuid.Parse(idStr)
	if err != nil {
		ErrorResponse(w, http.StatusBadRequest, "Invalid base ID format")
		return
	}

	var base models.Base
	if err := json.NewDecoder(r.Body).Decode(&base); err != nil {
		ErrorResponse(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	base.ID = id // Ensure the ID from the URL is used

	// Optional: Check if base exists before attempting update
	existingBase, err := h.Service.GetBaseByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			ErrorResponse(w, http.StatusNotFound, "Base not found")
			return
		}
		ErrorResponse(w, http.StatusInternalServerError, "Failed to check existing base")
		return
	}
	if existingBase == nil {
		ErrorResponse(w, http.StatusNotFound, "Base not found")
		return
	}

	// Update only allowed fields
	existingBase.Name = base.Name
	// existingBase.UserID = ... // Don't allow changing ownership here

	if err := h.Service.UpdateBase(existingBase); err != nil {
		ErrorResponse(w, http.StatusInternalServerError, "Failed to update base")
		return
	}

	JSONResponse(w, http.StatusOK, existingBase)
}

func (h *BaseHandler) DeleteBase(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["baseId"]
	id, err := uuid.Parse(idStr)
	if err != nil {
		ErrorResponse(w, http.StatusBadRequest, "Invalid base ID format")
		return
	}

	// Optional: Check if base exists before attempting delete
	existingBase, err := h.Service.GetBaseByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			ErrorResponse(w, http.StatusNotFound, "Base not found")
			return
		}
		ErrorResponse(w, http.StatusInternalServerError, "Failed to check existing base")
		return
	}
	if existingBase == nil {
		ErrorResponse(w, http.StatusNotFound, "Base not found")
		return
	}

	if err := h.Service.DeleteBase(id); err != nil {
		ErrorResponse(w, http.StatusInternalServerError, "Failed to delete base")
		return
	}

	w.WriteHeader(http.StatusNoContent) // 204 No Content on successful deletion
}
