package server

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"sync"

	"antiochus/internal/models"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		// Allow empty origin (non-browser clients) and localhost variants
		if origin == "" {
			return true
		}
		for _, allowed := range []string{
			"http://127.0.0.1", "http://localhost",
			"https://127.0.0.1", "https://localhost",
		} {
			if origin == allowed || strings.HasPrefix(origin, allowed+":") {
				return true
			}
		}
		return false
	},
}

type WSClient struct {
	conn *websocket.Conn
	send chan []byte
}

type WSHub struct {
	clients    map[*WSClient]bool
	broadcast  chan models.WSEvent
	register   chan *WSClient
	unregister chan *WSClient
	mu         sync.Mutex

	// EncryptFunc encrypts a JSON-serializable event for WS transport.
	// Returns the encrypted string payload. If nil, events are sent as plaintext JSON.
	EncryptFunc func(data interface{}) (string, error)
}

func NewWSHub() *WSHub {
	return &WSHub{
		clients:    make(map[*WSClient]bool),
		broadcast:  make(chan models.WSEvent, 256),
		register:   make(chan *WSClient),
		unregister: make(chan *WSClient),
	}
}

func (h *WSHub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			count := len(h.clients)
			h.mu.Unlock()
			log.Printf("[ws] client registered, total clients=%d", count)

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			count := len(h.clients)
			h.mu.Unlock()
			log.Printf("[ws] client unregistered, total clients=%d", count)

		case event := <-h.broadcast:
			h.mu.Lock()
			clientCount := len(h.clients)
			h.mu.Unlock()
			log.Printf("[ws] broadcasting event type=%s friend=%s to %d client(s)", event.Type, event.Room, clientCount)
			var data []byte
			var err error

			if h.EncryptFunc != nil {
				// Encrypt the event — send as {"encrypted": "iv.ciphertext"}
				encrypted, encErr := h.EncryptFunc(event)
				if encErr != nil {
					log.Printf("ws encrypt error: %v", encErr)
					// Fall back to plaintext if encryption fails
					data, _ = json.Marshal(event)
				} else {
					data, err = json.Marshal(map[string]string{"encrypted": encrypted})
					if err != nil {
						log.Printf("ws marshal error: %v", err)
						continue
					}
				}
			} else {
				data, err = json.Marshal(event)
				if err != nil {
					log.Printf("ws marshal error: %v", err)
					continue
				}
			}

			h.mu.Lock()
			for client := range h.clients {
				select {
				case client.send <- data:
				default:
					delete(h.clients, client)
					close(client.send)
				}
			}
			h.mu.Unlock()
		}
	}
}

func (h *WSHub) HandleUpgrade(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("ws upgrade error: %v", err)
		return
	}

	client := &WSClient{
		conn: conn,
		send: make(chan []byte, 256),
	}

	h.register <- client

	go client.writePump()
	go client.readPump(h)
}

func (c *WSClient) writePump() {
	defer c.conn.Close()
	for msg := range c.send {
		if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
			return
		}
	}
}

func (c *WSClient) readPump(h *WSHub) {
	defer func() {
		h.unregister <- c
		c.conn.Close()
	}()
	for {
		_, _, err := c.conn.ReadMessage()
		if err != nil {
			return
		}
	}
}

func (h *WSHub) Broadcast(event models.WSEvent) {
	h.broadcast <- event
}
