package websocket

import (
	"stock-tracker/internal/models"
	"stock-tracker/pkg/logger"
	"sync"

	"github.com/gorilla/websocket"
)

type Message struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

type Hub struct {
	clients    map[*Client]bool
	broadcast  chan *Message
	register   chan *Client
	unregister chan *Client
	mu         sync.RWMutex
}

type Client struct {
	hub  *Hub
	conn *websocket.Conn
	send chan *Message
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan *Message, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

func NewClient(hub *Hub, conn *websocket.Conn) *Client {
	return &Client{
		hub:  hub,
		conn: conn,
		send: make(chan *Message, 256),
	}
}
func (h *Hub) RegisterClient(client *Client) {
	h.register <- client
}
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()

			logger.Info().
				Int("total_clients", len(h.clients)).
				Msg("Client connected to WebSocket")

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			h.mu.Unlock()

			logger.Info().
				Int("total_clients", len(h.clients)).
				Msg("Client disconnected from WebSocket")

		case message := <-h.broadcast:
			h.mu.RLock()
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
			h.mu.RUnlock()
		}
	}
}

func (h *Hub) BroadcastStockUpdate(stock *models.Stock) {
	msg := &Message{
		Type:    "stock_update",
		Payload: stock,
	}

	select {
	case h.broadcast <- msg:
	default:
		logger.Warn().Msg("Broadcast channel full, dropping message")
	}
}

func (h *Hub) BroadcastAlert(alert *models.Alert) {
	msg := &Message{
		Type:    "alert",
		Payload: alert,
	}

	select {
	case h.broadcast <- msg:
	default:
		logger.Warn().Msg("Broadcast channel full, dropping alert message")
	}
}

func (c *Client) WritePump() {
	defer func() {
		c.conn.Close()
	}()

	for message := range c.send {
		err := c.conn.WriteJSON(message)
		if err != nil {
			logger.Error().Err(err).Msg("Error writing to WebSocket")
			return
		}
	}
}

func (c *Client) ReadPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	for {
		var msg Message
		err := c.conn.ReadJSON(&msg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logger.Error().Err(err).Msg("WebSocket error")
			}
			break
		}

		logger.Debug().
			Str("type", msg.Type).
			Msg("Received message from client")
	}
}
