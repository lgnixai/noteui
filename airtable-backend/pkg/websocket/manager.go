package websocket

import (
	"fmt"
	"log"
	"sync"

	"airtable-backend/pkg/redis"

	"github.com/google/uuid"
)

// Manager maintains the set of active clients and broadcasts messages to the
// clients.
type Manager struct {
	// Registered clients. Protected by clientsMutex.
	clients      map[*Client]bool
	clientsMutex sync.Mutex

	// Inbound messages from the clients. (Not used in this pub/sub model, but good to keep)
	// broadcast chan []byte

	// Register requests from the clients. This is an UNEXPORTED field.
	register chan *Client // This field remains unexported

	// Unregister requests from clients. This is an UNEXPORTED field.
	unregister chan *Client // This field remains unexported

	// Redis Subscriber for receiving updates. The PubSub instance handles the connection.
	redisSubscriber *redis.Subscriber

	// Manually track channels the manager is subscribed to in Redis. Protected by channelMutex.
	// Keys are Redis channel names (e.g., "table_updates:{tableId}").
	activeRedisSubscriptions map[string]bool
	channelMutex             sync.Mutex // Mutex for protecting activeRedisSubscriptions

	// Channel to signal the manager to process redis subscriptions updates
	updateSubscriptions chan struct{} // Buffered channel to coalesce signals
}

// NewManager creates a new Manager.
func NewManager(redisSub *redis.Subscriber) *Manager {
	m := &Manager{
		clients:                  make(map[*Client]bool),
		register:                 make(chan *Client), // Initialized the unexported channel
		unregister:               make(chan *Client), // Initialized the unexported channel
		redisSubscriber:          redisSub,
		activeRedisSubscriptions: make(map[string]bool),  // Initialize the tracking map
		updateSubscriptions:      make(chan struct{}, 1), // Buffered channel
	}

	return m
}

// Run starts the manager's goroutines.
func (m *Manager) Run() {
	// Start the goroutine to listen to Redis Pub/Sub messages
	go m.listenRedis()

	// Start the main goroutine to manage client registrations and subscription updates
	for {
		select {
		case client := <-m.register:
			// This is the internal handling of registration messages sent via RegisterClient method
			m.clientsMutex.Lock()
			m.clients[client] = true
			m.clientsMutex.Unlock()
			log.Printf("Client %s registered. Total clients: %d", client.ID, len(m.clients))
			// Start the client's read and write pumps here
			go client.writePump()
			go client.readPump() // readPump will send client to unregister on disconnect

			// Trigger subscription update check after a new client registers
			m.triggerSubscriptionUpdate()

		case client := <-m.unregister:
			// This is the internal handling of unregistration messages sent by readPump
			m.clientsMutex.Lock()
			if _, ok := m.clients[client]; ok {
				delete(m.clients, client)
				// Close the client's send channel to signal its writePump to exit
				close(client.send)
				log.Printf("Client %s unregistered. Total clients: %d", client.ID, len(m.clients))

				// Trigger subscription update check because this client's subscriptions are gone
				m.triggerSubscriptionUpdate()
			}
			m.clientsMutex.Unlock()

		case <-m.updateSubscriptions:
			// Process updates to which channels need Redis subscriptions
			m.updateRedisSubscriptions()
		}
	}
}

// FIX: Add an exported method to register clients
// RegisterClient registers a new client with the manager.
// Called by the WebSocket handler when a new connection is established.
func (m *Manager) RegisterClient(client *Client) {
	log.Printf("Sending client %s to manager register channel", client.ID)
	// Send the client to the internal register channel.
	// The Run method handles receiving from this channel.
	m.register <- client
}

// BroadcastMessage remains the same...
func (m *Manager) BroadcastMessage(channel string, message []byte) {
	// ... (same implementation)
}

// listenRedis remains the same...
func (m *Manager) listenRedis() {
	// ... (same implementation)
}

// SubscribeClientToTable remains the same...
// This method is already exported and correctly accesses the client's internal state.
func (m *Manager) SubscribeClientToTable(client *Client, tableID uuid.UUID) {
	// ... (same implementation)
}

// triggerSubscriptionUpdate remains the same...
func (m *Manager) triggerSubscriptionUpdate() {
	// ... (same implementation)
}

// updateRedisSubscriptions remains the same...
func (m *Manager) updateRedisSubscriptions() {
	// ... (same implementation, including the fix for c.subscribedTables from previous step if not already applied)
	log.Println("Processing Redis subscription update...")

	m.clientsMutex.Lock()
	desiredChannelsSet := make(map[string]bool)
	for client := range m.clients {
		if client == nil { // Add nil check for safety
			log.Println("Warning: nil client found in manager.clients map during subscription update")
			continue
		}
		for tableID := range client.subscribedTables { // Access client.subscribedTables (assuming it's a map)
			channel := fmt.Sprintf("table_updates:%s", tableID.String())
			desiredChannelsSet[channel] = true
		}
	}
	m.clientsMutex.Unlock()

	m.channelMutex.Lock()
	defer m.channelMutex.Unlock()
	// ... rest of updateRedisSubscriptions remains the same ...
	// (channelsToSubscribe, channelsToUnsubscribe logic and Redis Subscribe/Unsubscribe calls)
}

// GetSubscribedTableIDsForClient remains the same after the previous fix...
func (m *Manager) GetSubscribedTableIDsForClient(client *Client) []uuid.UUID {
	// ... (same implementation using 'client.subscribedTables')
	m.clientsMutex.Lock()
	defer m.clientsMutex.Unlock()

	if _, ok := m.clients[client]; ok {
		tableIDs := make([]uuid.UUID, 0, len(client.subscribedTables))
		for id := range client.subscribedTables {
			tableIDs = append(tableIDs, id)
		}
		return tableIDs
	}
	return []uuid.UUID{}
}
