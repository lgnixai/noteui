package routes

import (
	"airtable-backend/pkg/api/handlers"
	"net/http"

	"github.com/gorilla/mux"
)

func SetupRoutes(
	r *mux.Router,
	baseHandler *handlers.BaseHandler,
	tableHandler *handlers.TableHandler,
	fieldHandler *handlers.FieldHandler,
	recordHandler *handlers.RecordHandler,
	websocketHandler *handlers.WebSocketHandler,
) {
	api := r.PathPrefix("/api/v1").Subrouter()

	// Base routes
	api.HandleFunc("/bases", baseHandler.CreateBase).Methods("POST")
	api.HandleFunc("/bases", baseHandler.GetAllBases).Methods("GET")
	api.HandleFunc("/bases/{baseId}", baseHandler.GetBase).Methods("GET")
	api.HandleFunc("/bases/{baseId}", baseHandler.UpdateBase).Methods("PUT")
	api.HandleFunc("/bases/{baseId}", baseHandler.DeleteBase).Methods("DELETE")

	// Table routes (nested under base)
	api.HandleFunc("/bases/{baseId}/tables", tableHandler.CreateTable).Methods("POST")
	api.HandleFunc("/bases/{baseId}/tables", tableHandler.GetTablesByBase).Methods("GET")
	api.HandleFunc("/bases/{baseId}/tables/{tableId}", tableHandler.GetTable).Methods("GET") // Can also get table directly
	api.HandleFunc("/bases/{baseId}/tables/{tableId}", tableHandler.UpdateTable).Methods("PUT")
	api.HandleFunc("/bases/{baseId}/tables/{tableId}", tableHandler.DeleteTable).Methods("DELETE")

	// Field routes (nested under table)
	api.HandleFunc("/bases/{baseId}/tables/{tableId}/fields", fieldHandler.CreateField).Methods("POST")
	api.HandleFunc("/bases/{baseId}/tables/{tableId}/fields", fieldHandler.GetFieldsByTable).Methods("GET")
	api.HandleFunc("/bases/{baseId}/tables/{tableId}/fields/{fieldId}", fieldHandler.GetField).Methods("GET") // Can also get field directly
	api.HandleFunc("/bases/{baseId}/tables/{tableId}/fields/{fieldId}", fieldHandler.UpdateField).Methods("PUT")
	api.HandleFunc("/bases/{baseId}/tables/{tableId}/fields/{fieldId}", fieldHandler.DeleteField).Methods("DELETE")

	// Record routes (nested under table)
	api.HandleFunc("/bases/{baseId}/tables/{tableId}/records", recordHandler.CreateRecord).Methods("POST")
	api.HandleFunc("/bases/{baseId}/tables/{tableId}/records", recordHandler.GetRecords).Methods("GET")           // The main query endpoint
	api.HandleFunc("/bases/{baseId}/tables/{tableId}/records/{recordId}", recordHandler.GetRecord).Methods("GET") // Get single record
	api.HandleFunc("/bases/{baseId}/tables/{tableId}/records/{recordId}", recordHandler.UpdateRecord).Methods("PUT")
	api.HandleFunc("/bases/{baseId}/tables/{tableId}/records/{recordId}", recordHandler.DeleteRecord).Methods("DELETE")

	// WebSocket endpoint (not nested under base/table typically)
	r.HandleFunc("/ws", websocketHandler.ServeWS)

	// Optional: Basic health check
	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
}
