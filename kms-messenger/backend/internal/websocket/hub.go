package websocket

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"gorm.io/gorm"

	"kms-backend/internal/models"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for development
	},
}

type Client struct {
	ID     uint
	Conn   *websocket.Conn
	Hub    *Hub
	mu     sync.Mutex
}

type Hub struct {
	clients    map[uint]*Client
	broadcast  chan Message
	register   chan *Client
	unregister chan *Client
	mu         sync.RWMutex
	DB         *gorm.DB
}

type Message struct {
	Type      string          `json:"type"`
	ChatID    uint            `json:"chat_id,omitempty"`
	SenderID  uint            `json:"sender_id,omitempty"`
	Content   json.RawMessage `json:"content,omitempty"`
}

type ChatMessage struct {
	ID          uint        `json:"id"`
	ChatID      uint        `json:"chat_id"`
	SenderID    uint        `json:"sender_id"`
	Content     string      `json:"content"`
	MessageType string      `json:"message_type"`
	CreatedAt   string      `json:"created_at"`
	Sender      UserSummary `json:"sender"`
}

type UserSummary struct {
	ID        uint   `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	AvatarURL string `json:"avatar_url,omitempty"`
}

type PresenceMessage struct {
	Type     string `json:"type"`
	UserID   uint   `json:"user_id"`
	IsOnline bool   `json:"is_online"`
}

func NewHub(db *gorm.DB) *Hub {
	return &Hub{
		clients:    make(map[uint]*Client),
		broadcast:  make(chan Message, 100),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		DB:         db,
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client.ID] = client
			h.mu.Unlock()
			
			// Update user online status
			h.DB.Model(&models.User{}).Where("id = ?", client.ID).Update("is_online", true)
			
			// Broadcast presence
			h.broadcastPresence(client.ID, true)
			
		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client.ID]; ok {
				delete(h.clients, client.ID)
				client.Conn.Close()
			}
			h.mu.Unlock()
			
			// Update user online status
			h.DB.Model(&models.User{}).Where("id = ?", client.ID).Update("is_online", false)
			
			// Broadcast presence
			h.broadcastPresence(client.ID, false)
			
		case message := <-h.broadcast:
			h.mu.RLock()
			for _, client := range h.clients {
				// Send to all members of the chat
				if message.ChatID > 0 {
					// Check if client is member of this chat
					var count int64
					h.DB.Model(&models.ChatMember{}).Where("chat_id = ? AND user_id = ?", message.ChatID, client.ID).Count(&count)
					if count > 0 {
						select {
						case client.Conn.WriteJSON(message):
						default:
							// Client connection closed
							go func(c *Client) {
								h.unregister <- c
							}(client)
						}
					}
				} else if message.Type == "presence" {
					// Send presence to all
					select {
					case client.Conn.WriteJSON(message):
					default:
					}
				}
			}
			h.mu.RUnlock()
		}
	}
}

func (h *Hub) broadcastPresence(userID uint, isOnline bool) {
	msg := Message{
		Type:    "presence",
		Content: mustMarshal(PresenceMessage{Type: "presence", UserID: userID, IsOnline: isOnline}),
	}
	h.broadcast <- msg
}

func mustMarshal(v interface{}) json.RawMessage {
	data, err := json.Marshal(v)
	if err != nil {
		log.Printf("Failed to marshal: %v", err)
		return nil
	}
	return data
}

func WSHandler(hub *Hub) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			log.Printf("Failed to upgrade connection: %v", err)
			return
		}

		client := &Client{
			ID:   userID.(uint),
			Conn: conn,
			Hub:  hub,
		}

		hub.register <- client

		go client.readPump()
		go client.writePump()
	}
}

func (c *Client) readPump() {
	defer func() {
		c.Hub.unregister <- c
		c.Conn.Close()
	}()

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		var msg Message
		if err := json.Unmarshal(message, &msg); err != nil {
			log.Printf("Failed to unmarshal message: %v", err)
			continue
		}

		// Handle different message types
		switch msg.Type {
		case "chat_message":
			// Message will be saved via REST API, this is just for real-time delivery
			c.Hub.broadcast <- msg
		}
	}
}

func (c *Client) writePump() {
	defer func() {
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Conn.WriteJSON():
			if !ok {
				c.Hub.unregister <- c
				return
			}
			// Message already sent in broadcast
		}
	}
}
