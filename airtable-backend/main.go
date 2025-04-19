package main

import (
	"log"

	"airtable-backend/configs"
	"airtable-backend/pkg/api/handlers"
	"airtable-backend/pkg/api/routes"
	"airtable-backend/pkg/database"
	"airtable-backend/pkg/redis"
	"airtable-backend/pkg/services"
	"airtable-backend/pkg/websocket"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
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
	queryService := services.NewQueryService(database.DB)                            // Initialize Query Service

	// Initialize Handlers
	baseHandler := handlers.NewBaseHandler(baseService)
	tableHandler := handlers.NewTableHandler(tableService, baseService)                              // Pass BaseService to TableHandler
	fieldHandler := handlers.NewFieldHandler(fieldService, tableService)                             // Pass TableService to FieldHandler
	recordHandler := handlers.NewRecordHandler(recordService, tableService, wsManager, queryService) // Pass QueryService to RecordHandler
	websocketHandler := handlers.NewWebSocketHandler(wsManager)                                      // Pass WSManager

	// Setup Router
	r := gin.Default()

	// Configure CORS
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"Content-Type", "Authorization"}
	config.AllowCredentials = true
	config.MaxAge = 300
	r.Use(cors.New(config))

	// Setup routes
	routes.SetupRoutes(r, baseHandler, tableHandler, fieldHandler, recordHandler, websocketHandler)

	// Start Server
	log.Printf("Server starting on %s", cfg.ServerPort)
	if err := r.Run(cfg.ServerPort); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
