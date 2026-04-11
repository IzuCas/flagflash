package websocket

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/IzuCas/flagflash/internal/domain/entity"
	redisinfra "github.com/IzuCas/flagflash/internal/infrastructure/redis"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Configure appropriately in production
	},
}

// MessageType represents the type of WebSocket message
type MessageType string

const (
	MessageTypeFlagUpdate  MessageType = "flag_update"
	MessageTypeFlagDelete  MessageType = "flag_delete"
	MessageTypeFlagsSync   MessageType = "flags_sync"
	MessageTypeSubscribe   MessageType = "subscribe"
	MessageTypeUnsubscribe MessageType = "unsubscribe"
	MessageTypePing        MessageType = "ping"
	MessageTypePong        MessageType = "pong"
	MessageTypeError       MessageType = "error"
)

// Message represents a WebSocket message
type Message struct {
	Type          MessageType           `json:"type"`
	EnvironmentID uuid.UUID             `json:"environment_id,omitempty"`
	Flag          *entity.FeatureFlag   `json:"flag,omitempty"`
	Flags         []*entity.FeatureFlag `json:"flags,omitempty"`
	FlagKey       string                `json:"flag_key,omitempty"`
	Error         string                `json:"error,omitempty"`
	Timestamp     time.Time             `json:"timestamp"`
}

// Client represents a WebSocket client
type Client struct {
	hub      *Hub
	conn     *websocket.Conn
	send     chan []byte
	envs     map[uuid.UUID]bool
	tenantID uuid.UUID
	apiKeyID uuid.UUID
	mu       sync.RWMutex
}

// Hub manages WebSocket clients and broadcasts
type Hub struct {
	clients    map[*Client]bool
	envClients map[uuid.UUID]map[*Client]bool
	broadcast  chan *Message
	register   chan *Client
	unregister chan *Client
	mu         sync.RWMutex
	pubsub     *redisinfra.PubSub
}

// NewHub creates a new WebSocket hub
func NewHub(pubsub *redisinfra.PubSub) *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		envClients: make(map[uuid.UUID]map[*Client]bool),
		broadcast:  make(chan *Message, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		pubsub:     pubsub,
	}
}

// Run starts the hub
func (h *Hub) Run(ctx context.Context) {
	// Subscribe to Redis pub/sub for flag updates
	if h.pubsub != nil {
		go h.subscribeToRedis(ctx)
	}

	for {
		select {
		case <-ctx.Done():
			return
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
		case client := <-h.unregister:
			h.removeClient(client)
		case message := <-h.broadcast:
			h.broadcastMessage(message)
		}
	}
}

func (h *Hub) subscribeToRedis(ctx context.Context) {
	msgChan, cleanup, err := h.pubsub.SubscribeAll(ctx)
	if err != nil {
		log.Printf("Failed to subscribe to Redis: %v", err)
		return
	}
	defer cleanup()

	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-msgChan:
			if !ok {
				return
			}
			// Convert Redis message to WebSocket message
			wsMsg := &Message{
				EnvironmentID: msg.EnvironmentID,
				Timestamp:     time.Now(),
			}
			switch msg.Type {
			case "update":
				wsMsg.Type = MessageTypeFlagUpdate
				wsMsg.Flag = msg.Flag
			case "delete":
				wsMsg.Type = MessageTypeFlagDelete
				wsMsg.FlagKey = msg.FlagKey
			}
			h.broadcast <- wsMsg
		}
	}
}

func (h *Hub) removeClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.clients[client]; ok {
		delete(h.clients, client)
		close(client.send)

		// Remove from environment subscriptions
		client.mu.RLock()
		for envID := range client.envs {
			if envClients, ok := h.envClients[envID]; ok {
				delete(envClients, client)
				if len(envClients) == 0 {
					delete(h.envClients, envID)
				}
			}
		}
		client.mu.RUnlock()
	}
}

func (h *Hub) broadcastMessage(msg *Message) {
	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Failed to marshal message: %v", err)
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	// If message has environment ID, only send to subscribed clients
	if msg.EnvironmentID != uuid.Nil {
		if clients, ok := h.envClients[msg.EnvironmentID]; ok {
			for client := range clients {
				select {
				case client.send <- data:
				default:
					// Client buffer full, handled in write pump
				}
			}
		}
		return
	}

	// Broadcast to all clients
	for client := range h.clients {
		select {
		case client.send <- data:
		default:
			// Client buffer full
		}
	}
}

// SubscribeClient subscribes a client to an environment
func (h *Hub) SubscribeClient(client *Client, environmentID uuid.UUID) {
	h.mu.Lock()
	defer h.mu.Unlock()

	client.mu.Lock()
	client.envs[environmentID] = true
	client.mu.Unlock()

	if h.envClients[environmentID] == nil {
		h.envClients[environmentID] = make(map[*Client]bool)
	}
	h.envClients[environmentID][client] = true
}

// UnsubscribeClient unsubscribes a client from an environment
func (h *Hub) UnsubscribeClient(client *Client, environmentID uuid.UUID) {
	h.mu.Lock()
	defer h.mu.Unlock()

	client.mu.Lock()
	delete(client.envs, environmentID)
	client.mu.Unlock()

	if clients, ok := h.envClients[environmentID]; ok {
		delete(clients, client)
		if len(clients) == 0 {
			delete(h.envClients, environmentID)
		}
	}
}

// BroadcastFlagUpdate broadcasts a flag update
func (h *Hub) BroadcastFlagUpdate(environmentID uuid.UUID, flag *entity.FeatureFlag) {
	h.broadcast <- &Message{
		Type:          MessageTypeFlagUpdate,
		EnvironmentID: environmentID,
		Flag:          flag,
		Timestamp:     time.Now(),
	}
}

// BroadcastFlagDelete broadcasts a flag deletion
func (h *Hub) BroadcastFlagDelete(environmentID uuid.UUID, flagKey string) {
	h.broadcast <- &Message{
		Type:          MessageTypeFlagDelete,
		EnvironmentID: environmentID,
		FlagKey:       flagKey,
		Timestamp:     time.Now(),
	}
}

// ServeWs handles WebSocket upgrade and client connection.
// environmentID is the environment the client should be automatically subscribed to.
func (h *Hub) ServeWs(w http.ResponseWriter, r *http.Request, tenantID, apiKeyID, environmentID uuid.UUID) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}

	client := &Client{
		hub:      h,
		conn:     conn,
		send:     make(chan []byte, 256),
		envs:     make(map[uuid.UUID]bool),
		tenantID: tenantID,
		apiKeyID: apiKeyID,
	}

	h.register <- client

	// Auto-subscribe to the environment derived from the API key.
	if environmentID != uuid.Nil {
		h.SubscribeClient(client, environmentID)
	}

	go client.writePump()
	go client.readPump()
}

// NewSDKHandler returns an http.Handler that upgrades the connection to WebSocket
// and auto-subscribes the client to their environment.
// It MUST be placed behind the APIKeyAuth middleware so that the context already
// contains api_key_id, tenant_id and environment_id values.
func (h *Hub) NewSDKHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tenantID, _ := r.Context().Value("tenant_id").(uuid.UUID)
		apiKeyID, _ := r.Context().Value("api_key_id").(uuid.UUID)
		environmentID, _ := r.Context().Value("environment_id").(uuid.UUID)
		h.ServeWs(w, r, tenantID, apiKeyID, environmentID)
	})
}

func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, data, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		var msg Message
		if err := json.Unmarshal(data, &msg); err != nil {
			c.sendError("Invalid message format")
			continue
		}

		c.handleMessage(&msg)
	}
}

func (c *Client) handleMessage(msg *Message) {
	switch msg.Type {
	case MessageTypeSubscribe:
		if msg.EnvironmentID != uuid.Nil {
			c.hub.SubscribeClient(c, msg.EnvironmentID)
		}
	case MessageTypeUnsubscribe:
		if msg.EnvironmentID != uuid.Nil {
			c.hub.UnsubscribeClient(c, msg.EnvironmentID)
		}
	case MessageTypePing:
		c.sendPong()
	}
}

func (c *Client) sendError(errMsg string) {
	msg := Message{
		Type:      MessageTypeError,
		Error:     errMsg,
		Timestamp: time.Now(),
	}
	data, _ := json.Marshal(msg)
	select {
	case c.send <- data:
	default:
	}
}

func (c *Client) sendPong() {
	msg := Message{
		Type:      MessageTypePong,
		Timestamp: time.Now(),
	}
	data, _ := json.Marshal(msg)
	select {
	case c.send <- data:
	default:
	}
}

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
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Batch pending messages
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
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

// PublishFlagUpdate implements the FlagPublisher interface
func (h *Hub) PublishFlagUpdate(ctx context.Context, environmentID uuid.UUID, flag *entity.FeatureFlag) error {
	h.BroadcastFlagUpdate(environmentID, flag)
	if h.pubsub != nil {
		return h.pubsub.PublishFlagUpdate(ctx, environmentID, flag)
	}
	return nil
}

// PublishFlagDelete implements the FlagPublisher interface
func (h *Hub) PublishFlagDelete(ctx context.Context, environmentID uuid.UUID, flagKey string) error {
	h.BroadcastFlagDelete(environmentID, flagKey)
	if h.pubsub != nil {
		return h.pubsub.PublishFlagDelete(ctx, environmentID, flagKey)
	}
	return nil
}
