package handlers

import (
	"log"
	"net/http"

	"airtable-backend/pkg/websocket" // Import local websocket package (used for Client, Manager types)
	// Removed the conflicting import: "github.com/gorilla/websocket" // This line is removed

	"github.com/google/uuid"

	// Add gorilla/websocket with an alias to avoid name collision
	gowebsocket "github.com/gorilla/websocket" // Alias the external package
)

// Use the aliased Upgrader from the gorilla package
var upgrader = gowebsocket.Upgrader{ // Use gowebsocket.Upgrader
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// Allow connections from any origin (for development)
		// In production, restrict to your frontend domain
		return true
	},
}

type WebSocketHandler struct {
	Manager *websocket.Manager // Manager type comes from local websocket package
}

func NewWebSocketHandler(manager *websocket.Manager) *WebSocketHandler {
	return &WebSocketHandler{Manager: manager}
}

func (h *WebSocketHandler) ServeWS(w http.ResponseWriter, r *http.Request) {
	// Optional: Check user authentication here before upgrading

	// Use the aliased Upgrader's Upgrade method
	conn, err := upgrader.Upgrade(w, r, nil) // Use upgrader (which is gowebsocket.Upgrader)
	if err != nil {
		log.Println("WebSocket upgrade failed:", err)
		return
	}

	// NewClient function comes from your local websocket package, so use websocket.NewClient
	client := websocket.NewClient(h.Manager, conn)

	// FIX: Use the new exported method to register the client
	h.Manager.RegisterClient(client) // Call the exported method

	log.Printf("Client %s connected via WebSocket", client.ID)

	// ---- Auto-subscribe client to tables based on query params (Example) ----
	queryValues := r.URL.Query()
	tableIDStr := queryValues.Get("tableId")
	if tableIDStr != "" {
		if tableID, parseErr := uuid.Parse(tableIDStr); parseErr == nil {
			log.Printf("Client %s attempting to auto-subscribe to table %s from WS URL", client.ID, tableID)
			// Use the Manager's exported method to subscribe the client
			h.Manager.SubscribeClientToTable(client, tableID) // Call the exported method
			log.Printf("Client %s auto-subscribed to table %s", client.ID, tableID)

		} else {
			log.Printf("Client %s provided invalid tableId in WS URL: %s", client.ID, tableIDStr)
		}
	} else {
		log.Printf("Client %s connected via WS without specifying tableId in URL. No auto-subscription.", client.ID)
	}

	// Client's readPump and writePump are started by the Manager when the client is registered.
	// The readPump will handle unregistering the client on disconnect.
}
