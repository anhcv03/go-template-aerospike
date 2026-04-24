package websocket

import (
	"sync"

	"github.com/gorilla/websocket"
)

// Hub manages websocket clients and broadcasts messages to them.
type Hub struct {
	clients    map[*websocket.Conn]struct{}
	register   chan *websocket.Conn
	unregister chan *websocket.Conn
	broadcast  chan []byte
	once       sync.Once
}

func NewHub() *Hub {
	h := &Hub{
		clients:    make(map[*websocket.Conn]struct{}),
		register:   make(chan *websocket.Conn, 16),
		unregister: make(chan *websocket.Conn, 16),
		broadcast:  make(chan []byte, 64),
	}
	h.once.Do(func() { go h.run() })
	return h
}

func (h *Hub) run() {
	for {
		select {
		case conn := <-h.register:
			h.clients[conn] = struct{}{}
		case conn := <-h.unregister:
			if _, ok := h.clients[conn]; ok {
				delete(h.clients, conn)
				conn.Close()
			}
		case msg := <-h.broadcast:
			for conn := range h.clients {
				if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
					delete(h.clients, conn)
					conn.Close()
				}
			}
		}
	}
}

// Register adds a websocket connection to the hub.
func (h *Hub) Register(conn *websocket.Conn) {
	h.register <- conn
}

// Unregister removes a websocket connection from the hub.
func (h *Hub) Unregister(conn *websocket.Conn) {
	h.unregister <- conn
}

// Broadcast sends a message to all connected clients.
// Non-blocking to avoid backpressure on request handlers.
func (h *Hub) Broadcast(msg []byte) {
	select {
	case h.broadcast <- msg:
	default:
	}
}

