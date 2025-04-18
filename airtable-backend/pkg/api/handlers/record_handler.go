package handlers

import (
	"airtable-backend/pkg/models"
	"airtable-backend/pkg/services"
	"airtable-backend/pkg/websocket" // Need WS Manager to subscribe clients initially
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

type RecordHandler struct {
	Service      *services.RecordService
	TableService *services.TableService // Need TableService to check if table exists
	WSManager    *websocket.Manager     // To subscribe client to table updates on initial GET
}

func NewRecordHandler(s *services.RecordService, ts *services.TableService, wsManager *websocket.Manager) *RecordHandler {
	return &RecordHandler{Service: s, TableService: ts, WSManager: wsManager}
}

func (h *RecordHandler) CreateRecord(w http.ResponseWriter, r *http.Request) {
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

	var data json.RawMessage // Accept raw JSON for the data field
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		ErrorResponse(w, http.StatusBadRequest, "Invalid request payload for record data")
		return
	}

	record, err := h.Service.CreateRecord(tableID, data)
	if err != nil {
		ErrorResponse(w, http.StatusInternalServerError, "Failed to create record")
		return
	}

	JSONResponse(w, http.StatusCreated, record)
}

func (h *RecordHandler) GetRecords(w http.ResponseWriter, r *http.Request) {
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

	// Parse query parameters
	queryValues := r.URL.Query()

	limit, _ := strconv.Atoi(queryValues.Get("limit"))
	offset, _ := strconv.Atoi(queryValues.Get("offset"))

	// Filter and Sort parameters are expected as JSON strings (might need encoding)
	// In a real scenario, these might be base64 encoded or passed in a POST body for complex queries
	// For simplicity here, we'll assume they are URL-encoded JSON strings
	filterJSONStr := queryValues.Get("filter")
	sortJSONStr := queryValues.Get("sort")

	// Decode URL-encoded JSON strings if necessary, or assume plain JSON string value in query param
	// Let's assume plain JSON string value for simplicity in this example.
	// In a real app, you'd use something like:
	// filterBytes, err := url.QueryUnescape(filterJSONStr)
	// sortBytes, err := url.QueryUnescape(sortJSONStr)

	filterBytes := []byte(filterJSONStr)
	sortBytes := []byte(sortJSONStr)

	records, total, err := h.Service.GetRecords(tableID, filterBytes, sortBytes, limit, offset)
	if err != nil {
		// Check specific errors from query building or DB
		if strings.Contains(err.Error(), "invalid filter json") || strings.Contains(err.Error(), "invalid sort json") ||
			strings.Contains(err.Error(), "failed to build filter query") || strings.Contains(err.Error(), "failed to build sort query") ||
			strings.Contains(err.Error(), "unsupported operator") || strings.Contains(err.Error(), "invalid value format") ||
			strings.Contains(err.Error(), "field with ID") {
			ErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("Query parameter error: %v", err))
			return
		}
		ErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("Failed to retrieve records: %v", err))
		return
	}

	// Response structure including records and total count
	response := struct {
		Records []models.Record `json:"records"`
		Total   int64           `json:"total"`
	}{
		Records: records,
		Total:   total,
	}

	JSONResponse(w, http.StatusOK, response)

	// ---- WebSocket Subscription Handling (Optional, but useful for initial view load) ----
	// If this is a request coming from a WebSocket client that just connected and loaded initial data,
	// we can associate this client with the table ID here. This requires passing the client ID
	// from the WebSocket connection context to the HTTP request context, which is advanced.
	// A simpler approach: The WebSocket connection handler immediately subscribes the client
	// to any table ID provided in the WebSocket upgrade request URL (e.g., /ws?tableId=...).
	// Let's handle subscription in the websocket handler for simplicity.
}

func (h *RecordHandler) GetRecord(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	recordIDStr := vars["recordId"]
	recordID, err := uuid.Parse(recordIDStr)
	if err != nil {
		ErrorResponse(w, http.StatusBadRequest, "Invalid record ID format")
		return
	}

	record, err := h.Service.GetRecordByID(recordID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			ErrorResponse(w, http.StatusNotFound, "Record not found")
			return
		}
		ErrorResponse(w, http.StatusInternalServerError, "Failed to get record")
		return
	}
	if record == nil {
		ErrorResponse(w, http.StatusNotFound, "Record not found")
		return
	}

	// Optional: Validate baseId/tableId from URL if needed

	JSONResponse(w, http.StatusOK, record)
}

func (h *RecordHandler) UpdateRecord(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	recordIDStr := vars["recordId"]
	recordID, err := uuid.Parse(recordIDStr)
	if err != nil {
		ErrorResponse(w, http.StatusBadRequest, "Invalid record ID format")
		return
	}

	var newData json.RawMessage // Accept raw JSON for the data field
	if err := json.NewDecoder(r.Body).Decode(&newData); err != nil {
		ErrorResponse(w, http.StatusBadRequest, "Invalid request payload for record data")
		return
	}

	updatedRecord, err := h.Service.UpdateRecord(recordID, newData)
	if err != nil {
		// Check if it's a "not found" error from the service
		if strings.Contains(err.Error(), "record with ID") && strings.Contains(err.Error(), "not found") {
			ErrorResponse(w, http.StatusNotFound, err.Error()) // Propagate "not found" message
			return
		}
		ErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("Failed to update record: %v", err))
		return
	}

	JSONResponse(w, http.StatusOK, updatedRecord)
}

func (h *RecordHandler) DeleteRecord(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	recordIDStr := vars["recordId"]
	recordID, err := uuid.Parse(recordIDStr)
	if err != nil {
		ErrorResponse(w, http.StatusBadRequest, "Invalid record ID format")
		return
	}

	// Optional: Check if record exists before delete (service does this, but handler can too)
	existingRecord, err := h.Service.GetRecordByID(recordID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			ErrorResponse(w, http.StatusNotFound, "Record not found")
			return
		}
		ErrorResponse(w, http.StatusInternalServerError, "Failed to check existing record")
		return
	}
	if existingRecord == nil {
		ErrorResponse(w, http.StatusNotFound, "Record not found")
		return
	}

	if err := h.Service.DeleteRecord(recordID); err != nil {
		ErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("Failed to delete record: %v", err))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
