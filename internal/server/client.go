package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/RedPaladin7/peerpoker/internal/game"
	"github.com/RedPaladin7/peerpoker/internal/protocol"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512 * 1024
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for development
	},
}

type Client struct {
	ID     string
	conn   *websocket.Conn
	hub    *WebSocketHub
	game   *game.Game
	send   chan []byte
	IsPeer bool
}

func NewClientFromHTTP(w http.ResponseWriter, r *http.Request, hub *WebSocketHub, g *game.Game, isPeer bool) (*Client, error) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return nil, err
	}

	clientID := r.Header.Get("X-Client-ID")
	if clientID == "" {
		clientID = r.RemoteAddr + "-" + time.Now().Format("20060102150405")
	}

	client := &Client{
		ID:     clientID,
		conn:   conn,
		hub:    hub,
		game:   g,
		send:   make(chan []byte, 256),
		IsPeer: isPeer,
	}

	return client, nil
}

func (c *Client) ReadPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
		
		// Remove from game
		c.game.RemovePlayer(c.ID)
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logrus.Errorf("WebSocket error: %v", err)
			}
			break
		}

		if err := c.handleMessage(message); err != nil {
			logrus.Errorf("Message handling error: %v", err)
		}
	}
}

func (c *Client) WritePump() {
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

			// Add queued messages to current websocket message
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

func (c *Client) handleMessage(data []byte) error {
	var msg protocol.Message
	if err := json.Unmarshal(data, &msg); err != nil {
		return err
	}

	logrus.WithFields(logrus.Fields{
		"from":    c.ID,
		"type":    msg.Type,
		"payload": len(msg.Payload),
	}).Debug("Received message")

	return c.game.HandleMessage(c.ID, &msg)
}

func (c *Client) Send(data []byte) error {
	select {
	case c.send <- data:
		return nil
	default:
		return fmt.Errorf("send buffer full")
	}
}

func (c *Client) Close() {
	c.conn.Close()
}
