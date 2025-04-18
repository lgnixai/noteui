package main

import (
	"log"
	"net/http"

	"airtable-backend/configs"
	"airtable-backend/pkg/api/handlers"
	"airtable-backend/pkg/api/routes"
	"airtable-backend/pkg/database"
	"airtable-backend/pkg/redis"
	"airtable-backend/pkg/services"
	"airtable-backend/pkg/websocket"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

func main() {
	// Load configuration
	cfg := configs.LoadConfig()

	// Connect to Database
	database.ConnectDB(cfg)

	// Connect to Redis
	redis.ConnectRedis(cfg)
	// Close Redis connection when main exits (optional but good practice)
	defer redis.RDB.Close()

	// Setup Redis subscriber and WebSocket Manager
	redisSubscriber := redis.NewSubscriber()
	defer redisSubscriber.Close() // Close subscriber on exit

	wsManager := websocket.NewManager(redisSubscriber)
	go wsManager.Run() // Run the WebSocket manager in a goroutine

	// Initialize Services
	baseService := services.NewBaseService(database.DB)
	tableService := services.NewTableService(database.DB)
	// Field service is needed by record service
	fieldService := services.NewFieldService(database.DB)
	recordService := services.NewRecordService(database.DB, wsManager, fieldService) // Pass WSManager and FieldService

	// Initialize Handlers
	baseHandler := handlers.NewBaseHandler(baseService)
	tableHandler := handlers.NewTableHandler(tableService, baseService)                // Pass BaseService to TableHandler
	fieldHandler := handlers.NewFieldHandler(fieldService, tableService)               // Pass TableService to FieldHandler
	recordHandler := handlers.NewRecordHandler(recordService, tableService, wsManager) // Pass TableService and WSManager
	websocketHandler := handlers.NewWebSocketHandler(wsManager)                        // Pass WSManager

	// Setup Router
	r := mux.NewRouter()
	routes.SetupRoutes(r, baseHandler, tableHandler, fieldHandler, recordHandler, websocketHandler)

	// Configure CORS
	corsHandler := cors.New(cors.Options{
		AllowedOrigins: []string{"*"}, // Allow all origins

		//AllowedOrigins:   []string{cfg.CORSOrigin}, // Use the CORS_ORIGIN from config
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	})

	// Wrap the router with the CORS handler
	handler := corsHandler.Handler(r)

	// Start Server
	log.Printf("Server starting on %s", cfg.ServerPort)
	err := http.ListenAndServe(cfg.ServerPort, handler)
	if err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
