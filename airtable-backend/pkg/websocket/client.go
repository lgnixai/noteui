package websocket

import (
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

// Client is a middleman between the websocket connection and the hub.
type Client struct {
	ID      uuid.UUID
	manager *Manager
	conn    *websocket.Conn
	// Buffered channel of outbound messages.
	send chan []byte

	subscribedTables map[uuid.UUID]bool // Tables this client is interested in
}

// NewClient creates a new WebSocket client.
func NewClient(manager *Manager, conn *websocket.Conn) *Client {
	return &Client{
		ID:               uuid.New(),
		manager:          manager,
		conn:             conn,
		send:             make(chan []byte, 256),
		subscribedTables: make(map[uuid.UUID]bool),
	}
}

// readPump pumps messages from the websocket connection to the hub.
//
// The application runs readPump in a separate goroutine for each
// connection. The application ensures that there are at most a
// single reader on a connection by executing all reads from this
// goroutine.
func (c *Client) readPump() {
	defer func() {
		c.manager.unregister <- c
		c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	for {
		// This is where we could handle incoming messages from the client,
		// e.g., subscribe/unsubscribe requests.
		// For this example, we only expect pings/pongs and ignore other messages.
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
		// Process incoming message if needed
		// log.Printf("Received message from client %s: %s", c.ID, message)
		// Example: handle subscribe/unsubscribe messages
		// c.handleIncomingMessage(message)
		_ = message // Ignore for now
	}
}

// writePump pumps messages from the hub to the websocket connection.
//
// A goroutine running writePump is started for each connection. The
// application ensures that there is at most a single writer to a
// connection by executing all writes from this goroutine.
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The manager closed the channel.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued chat messages to the current websocket message.
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write(newline)
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// SendMessage sends a message to the client's send channel.
func (c *Client) SendMessage(message []byte) {
	select {
	case c.send <- message:
	default:
		// If channel is full, client is likely slow or disconnected, unregister it.
		close(c.send)
		c.manager.unregister <- c
		log.Printf("Client %s send channel is full, unregistering", c.ID)
	}
}

// SubscribeToTable marks the client as subscribed to a specific table.
func (c *Client) SubscribeToTable(tableID uuid.UUID) {
	c.subscribedTables[tableID] = true
	log.Printf("Client %s subscribed to table %s", c.ID, tableID)
	// Note: Actual Redis subscription is managed by the Manager
}

// UnsubscribeFromTable marks the client as unsubscribed from a specific table.
func (c *Client) UnsubscribeFromTable(tableID uuid.UUID) {
	delete(c.subscribedTables, tableID)
	log.Printf("Client %s unsubscribed from table %s", c.ID, tableID)
	// Note: Actual Redis unsubscription is managed by the Manager
}

// IsSubscribedToTable checks if the client is subscribed to a specific table.
func (c *Client) IsSubscribedToTable(tableID uuid.UUID) bool {
	_, ok := c.subscribedTables[tableID]
	return ok
}

// handleIncomingMessage processes messages received from the client.
// This is a placeholder; implement logic based on client message format.
// func (c *Client) handleIncomingMessage(message []byte) {
//     // Example message format: {"action": "subscribe", "tableId": "..."}
//     // var msg struct { Action string `json:"action"` TableID string `json:"tableId"` }
//     // if err := json.Unmarshal(message, &msg); err != nil {
//     //     log.Printf("Error unmarshalling client message: %v", err)
//     //     return
//     // }
//     //
//     // switch msg.Action {
//     // case "subscribe":
//     //     if tableID, err := uuid.Parse(msg.TableID); err == nil {
//     //          c.SubscribeToTable(tableID) // Update client's state
//     //          // Inform manager to manage Redis subscription if needed
//     //     }
//     // case "unsubscribe":
//     //     if tableID, err := uuid.Parse(msg.TableID); err == nil {
//     //          c.UnsubscribeFromTable(tableID) // Update client's state
//     //          // Inform manager to manage Redis subscription if needed
//     //     }
//     // // ... other actions like heartbeats, initial data requests, etc.
//     // }
// }
