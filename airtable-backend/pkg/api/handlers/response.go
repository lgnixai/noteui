package handlers

import (
	"github.com/gin-gonic/gin"
)

// JSONResponse is a helper to write JSON responses.
func JSONResponse(c *gin.Context, status int, data interface{}) {
	if data != nil {
		c.JSON(status, data)
	} else {
		c.Status(status)
	}
}

// ErrorResponse is a helper to write JSON error responses.
func ErrorResponse(c *gin.Context, status int, message string) {
	c.JSON(status, gin.H{"error": message})
}
