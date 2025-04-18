package handlers

import (
	"encoding/json"
	"log"
	"net/http"
)

// JSONResponse is a helper to write JSON responses.
func JSONResponse(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		if err := json.NewEncoder(w).Encode(data); err != nil {
			log.Printf("Error encoding JSON response: %v", err)
			// Optionally write a generic error response here
		}
	}
}

// ErrorResponse is a helper to write JSON error responses.
func ErrorResponse(w http.ResponseWriter, status int, message string) {
	JSONResponse(w, status, map[string]string{"error": message})
}
