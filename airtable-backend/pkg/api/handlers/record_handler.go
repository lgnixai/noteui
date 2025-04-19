package handlers

import (
	"airtable-backend/pkg/models"
	"airtable-backend/pkg/services"
	"airtable-backend/pkg/websocket" // Need WS Manager to subscribe clients initially
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type RecordHandler struct {
	Service      *services.RecordService
	TableService *services.TableService // Need TableService to check if table exists
	WSManager    *websocket.Manager     // To subscribe client to table updates on initial GET
	QueryService *services.QueryService
}

func NewRecordHandler(s *services.RecordService, ts *services.TableService, wsManager *websocket.Manager, qs *services.QueryService) *RecordHandler {
	return &RecordHandler{Service: s, TableService: ts, WSManager: wsManager, QueryService: qs}
}

func (h *RecordHandler) CreateRecord(c *gin.Context) {
	tableID, err := uuid.Parse(c.Param("tableId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid table ID"})
		return
	}

	var data map[string]interface{}
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// 将map转换为json.RawMessage
	rawData, err := json.Marshal(data)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to marshal data"})
		return
	}

	record, err := h.Service.CreateRecord(tableID, rawData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, record)
}

func (h *RecordHandler) GetRecords(c *gin.Context) {
	tableID, err := uuid.Parse(c.Param("tableId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid table ID"})
		return
	}

	var params models.QueryParams
	if err := c.ShouldBindJSON(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid query parameters"})
		return
	}

	result, err := h.QueryService.QueryRecords(tableID, params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *RecordHandler) GetRecord(c *gin.Context) {
	recordID, err := uuid.Parse(c.Param("recordId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid record ID"})
		return
	}

	record, err := h.Service.GetRecordByID(recordID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if record == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Record not found"})
		return
	}

	c.JSON(http.StatusOK, record)
}

func (h *RecordHandler) UpdateRecord(c *gin.Context) {
	recordID, err := uuid.Parse(c.Param("recordId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid record ID"})
		return
	}

	var data map[string]interface{}
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// 将map转换为json.RawMessage
	rawData, err := json.Marshal(data)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to marshal data"})
		return
	}

	record, err := h.Service.UpdateRecord(recordID, rawData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, record)
}

func (h *RecordHandler) DeleteRecord(c *gin.Context) {
	recordID, err := uuid.Parse(c.Param("recordId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid record ID"})
		return
	}

	if err := h.Service.DeleteRecord(recordID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}
