package routes

import (
	"airtable-backend/pkg/api/handlers"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(
	r *gin.Engine,
	baseHandler *handlers.BaseHandler,
	tableHandler *handlers.TableHandler,
	fieldHandler *handlers.FieldHandler,
	recordHandler *handlers.RecordHandler,
	websocketHandler *handlers.WebSocketHandler,
) {
	api := r.Group("/api/v1")

	// Base routes
	api.POST("/bases", baseHandler.CreateBase)
	api.GET("/bases", baseHandler.GetAllBases)
	api.GET("/bases/:baseId", baseHandler.GetBase)
	api.PUT("/bases/:baseId", baseHandler.UpdateBase)
	api.DELETE("/bases/:baseId", baseHandler.DeleteBase)

	// Table routes (nested under base)
	api.POST("/bases/:baseId/tables", tableHandler.CreateTable)
	api.GET("/bases/:baseId/tables", tableHandler.GetTablesByBase)
	api.GET("/bases/:baseId/tables/:tableId", tableHandler.GetTable)
	api.PUT("/bases/:baseId/tables/:tableId", tableHandler.UpdateTable)
	api.DELETE("/bases/:baseId/tables/:tableId", tableHandler.DeleteTable)

	// Field routes (nested under table)
	api.POST("/bases/:baseId/tables/:tableId/fields", fieldHandler.CreateField)
	api.GET("/bases/:baseId/tables/:tableId/fields", fieldHandler.GetFieldsByTable)
	api.GET("/bases/:baseId/tables/:tableId/fields/:fieldId", fieldHandler.GetField)
	api.PUT("/bases/:baseId/tables/:tableId/fields/:fieldId", fieldHandler.UpdateField)
	api.DELETE("/bases/:baseId/tables/:tableId/fields/:fieldId", fieldHandler.DeleteField)
	api.PUT("/bases/:baseId/tables/:tableId/fields/order", fieldHandler.UpdateFieldOrder)
	api.POST("/bases/:baseId/tables/:tableId/fields/:fieldId/validate", fieldHandler.ValidateFieldValue)

	// Record routes (nested under table)
	api.POST("/bases/:baseId/tables/:tableId/records", recordHandler.CreateRecord)
	api.GET("/bases/:baseId/tables/:tableId/records", recordHandler.GetRecords)
	api.GET("/bases/:baseId/tables/:tableId/records/:recordId", recordHandler.GetRecord)
	api.PUT("/bases/:baseId/tables/:tableId/records/:recordId", recordHandler.UpdateRecord)
	api.DELETE("/bases/:baseId/tables/:tableId/records/:recordId", recordHandler.DeleteRecord)

	// WebSocket endpoint
	r.GET("/ws", websocketHandler.ServeWS)

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.String(200, "OK")
	})
}
