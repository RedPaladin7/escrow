package transport

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

const (
	WriteWait      = 10 * time.Second
	PongWait       = 60 * time.Second
	PingPeriod     = (PongWait * 9) / 10
	MaxMessageSize = 512 * 1024
)

var DefaultUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins
	},
}

type WSConnection struct {
	conn *websocket.Conn
	send chan []byte
}

func NewWSConnection(conn *websocket.Conn) *WSConnection {
	return &WSConnection{
		conn: conn,
		send: make(chan []byte, 256),
	}
}

func (wsc *WSConnection) ReadLoop(onMessage func([]byte) error, onClose func()) {
	defer func() {
		wsc.conn.Close()
		if onClose != nil {
			onClose()
		}
	}()

	wsc.conn.SetReadLimit(MaxMessageSize)
	wsc.conn.SetReadDeadline(time.Now().Add(PongWait))
	wsc.conn.SetPongHandler(func(string) error {
		wsc.conn.SetReadDeadline(time.Now().Add(PongWait))
		return nil
	})

	for {
		_, message, err := wsc.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logrus.Errorf("WebSocket read error: %v", err)
			}
			break
		}

		if onMessage != nil {
			if err := onMessage(message); err != nil {
				logrus.Errorf("Message handler error: %v", err)
			}
		}
	}
}

func (wsc *WSConnection) WriteLoop() {
	ticker := time.NewTicker(PingPeriod)
	defer func() {
		ticker.Stop()
		wsc.conn.Close()
	}()

	for {
		select {
		case message, ok := <-wsc.send:
			wsc.conn.SetWriteDeadline(time.Now().Add(WriteWait))
			if !ok {
				wsc.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := wsc.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued messages
			n := len(wsc.send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-wsc.send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			wsc.conn.SetWriteDeadline(time.Now().Add(WriteWait))
			if err := wsc.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (wsc *WSConnection) Send(data []byte) error {
	select {
	case wsc.send <- data:
		return nil
	default:
		return fmt.Errorf("send buffer full")
	}
}

func (wsc *WSConnection) SendJSON(v interface{}) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return wsc.Send(data)
}

func (wsc *WSConnection) Close() error {
	close(wsc.send)
	return wsc.conn.Close()
}
