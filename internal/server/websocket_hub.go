package server

import (
	"context"
	"sync"

	"github.com/RedPaladin7/peerpoker/internal/protocol"
	"github.com/sirupsen/logrus"
)

type WebSocketHub struct {
	clients    map[*Client]bool
	broadcast  chan *protocol.BroadcastMessage
	Register   chan *Client
	unregister chan *Client
	mu         sync.RWMutex
	closed     bool
}

func NewWebSocketHub() *WebSocketHub {
	return &WebSocketHub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan *protocol.BroadcastMessage, 256),
		Register:   make(chan *Client, 10),
		unregister: make(chan *Client, 10),
	}
}

func (h *WebSocketHub) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			h.shutdownAllClients()
			return
			
		case client := <-h.Register:
			h.registerClient(client)
			
		case client := <-h.unregister:
			h.unregisterClient(client)
			
		case message := <-h.broadcast:
			h.broadcastMessage(message)
		}
	}
}

func (h *WebSocketHub) registerClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	
	h.clients[client] = true
	logrus.WithFields(logrus.Fields{
		"client_id": client.ID,
		"peer":      client.IsPeer,
		"total":     len(h.clients),
	}).Info("Client registered")
}

func (h *WebSocketHub) unregisterClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	
	if _, ok := h.clients[client]; ok {
		delete(h.clients, client)
		close(client.send)
		
		logrus.WithFields(logrus.Fields{
			"client_id": client.ID,
			"total":     len(h.clients),
		}).Info("Client unregistered")
	}
}

func (h *WebSocketHub) broadcastMessage(msg *protocol.BroadcastMessage) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	
	if len(msg.To) == 0 {
		// Broadcast to all clients
		for client := range h.clients {
			select {
			case client.send <- msg.Data:
			default:
				logrus.Warnf("Client %s send buffer full, dropping message", client.ID)
			}
		}
	} else {
		// Broadcast to specific targets
		for client := range h.clients {
			for _, targetID := range msg.To {
				if client.ID == targetID {
					select {
					case client.send <- msg.Data:
					default:
						logrus.Warnf("Client %s send buffer full, dropping message", client.ID)
					}
					break
				}
			}
		}
	}
}

func (h *WebSocketHub) Broadcast(data []byte, targets ...string) {
	h.mu.RLock()
	if h.closed {
		h.mu.RUnlock()
		return
	}
	h.mu.RUnlock()
	
	msg := &protocol.BroadcastMessage{
		Data: data,
		To:   targets,
	}
	
	select {
	case h.broadcast <- msg:
	default:
		logrus.Warn("Broadcast channel full, dropping message")
	}
}

func (h *WebSocketHub) ClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

func (h *WebSocketHub) GetClientIDs() []string {
	h.mu.RLock()
	defer h.mu.RUnlock()
	
	ids := make([]string, 0, len(h.clients))
	for client := range h.clients {
		ids = append(ids, client.ID)
	}
	return ids
}

func (h *WebSocketHub) shutdownAllClients() {
	h.mu.Lock()
	defer h.mu.Unlock()
	
	for client := range h.clients {
		client.Close()
	}
	h.clients = make(map[*Client]bool)
}

func (h *WebSocketHub) Close() {
	h.mu.Lock()
	defer h.mu.Unlock()
	
	if !h.closed {
		h.closed = true
		close(h.broadcast)
	}
}
