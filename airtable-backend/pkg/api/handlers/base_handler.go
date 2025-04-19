package handlers

import (
	"airtable-backend/pkg/models"
	"airtable-backend/pkg/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type BaseHandler struct {
	Service *services.BaseService
}

func NewBaseHandler(s *services.BaseService) *BaseHandler {
	return &BaseHandler{Service: s}
}

func (h *BaseHandler) CreateBase(c *gin.Context) {
	var base models.Base
	if err := c.ShouldBindJSON(&base); err != nil {
		ErrorResponse(c, 400, "Invalid request payload")
		return
	}

	if err := h.Service.CreateBase(&base); err != nil {
		ErrorResponse(c, 500, "Failed to create base")
		return
	}

	JSONResponse(c, 201, base)
}

func (h *BaseHandler) GetBase(c *gin.Context) {
	idStr := c.Param("baseId")
	id, err := uuid.Parse(idStr)
	if err != nil {
		ErrorResponse(c, 400, "Invalid base ID format")
		return
	}

	base, err := h.Service.GetBaseByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			ErrorResponse(c, 404, "Base not found")
			return
		}
		ErrorResponse(c, 500, "Failed to get base")
		return
	}
	if base == nil {
		ErrorResponse(c, 404, "Base not found")
		return
	}

	JSONResponse(c, 200, base)
}

func (h *BaseHandler) GetAllBases(c *gin.Context) {
	bases, err := h.Service.GetAllBases()
	if err != nil {
		ErrorResponse(c, 500, "Failed to get bases")
		return
	}

	JSONResponse(c, 200, bases)
}

func (h *BaseHandler) UpdateBase(c *gin.Context) {
	idStr := c.Param("baseId")
	id, err := uuid.Parse(idStr)
	if err != nil {
		ErrorResponse(c, 400, "Invalid base ID format")
		return
	}

	var base models.Base
	if err := c.ShouldBindJSON(&base); err != nil {
		ErrorResponse(c, 400, "Invalid request payload")
		return
	}
	base.ID = id

	existingBase, err := h.Service.GetBaseByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			ErrorResponse(c, 404, "Base not found")
			return
		}
		ErrorResponse(c, 500, "Failed to check existing base")
		return
	}
	if existingBase == nil {
		ErrorResponse(c, 404, "Base not found")
		return
	}

	existingBase.Name = base.Name

	if err := h.Service.UpdateBase(existingBase); err != nil {
		ErrorResponse(c, 500, "Failed to update base")
		return
	}

	JSONResponse(c, 200, existingBase)
}

func (h *BaseHandler) DeleteBase(c *gin.Context) {
	idStr := c.Param("baseId")
	id, err := uuid.Parse(idStr)
	if err != nil {
		ErrorResponse(c, 400, "Invalid base ID format")
		return
	}

	existingBase, err := h.Service.GetBaseByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			ErrorResponse(c, 404, "Base not found")
			return
		}
		ErrorResponse(c, 500, "Failed to check existing base")
		return
	}
	if existingBase == nil {
		ErrorResponse(c, 404, "Base not found")
		return
	}

	if err := h.Service.DeleteBase(id); err != nil {
		ErrorResponse(c, 500, "Failed to delete base")
		return
	}

	c.Status(204)
}
