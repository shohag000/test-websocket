package ws

import (
	"github.com/shohag000/test-websocket/model"
)

// Hub maintains the set of active clients and broadcasts messages to the
// clients.
type Hub struct {
	// Registered clients.
	clients map[*Client]bool

	// Inbound messages from the clients.
	// broadcast chan []byte
	broadcast chan model.Data

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client
}

// NewHub returns a new hub
func NewHub() *Hub {
	return &Hub{
		broadcast:  make(chan model.Data),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
	}
}

// Run runs the hub
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
		case iData := <-h.broadcast:
			for client := range h.clients {
				// Send message to a particular user only
				if (client.UserID != iData.UserID) && iData.DataType != model.ErrorData {
					continue
				}

				select {
				case client.send <- iData:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
		}
	}
}
