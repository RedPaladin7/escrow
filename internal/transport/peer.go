package transport

import (
	"fmt"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

type PeerConnection struct {
	ID          string
	RemoteAddr  string
	conn        *websocket.Conn
	send        chan []byte
	isOutbound  bool
	lastPing    time.Time
	mu          sync.RWMutex
}

func NewPeerConnection(id, remoteAddr string, conn *websocket.Conn, isOutbound bool) *PeerConnection {
	return &PeerConnection{
		ID:         id,
		RemoteAddr: remoteAddr,
		conn:       conn,
		send:       make(chan []byte, 256),
		isOutbound: isOutbound,
		lastPing:   time.Now(),
	}
}

func (pc *PeerConnection) IsOutbound() bool {
	pc.mu.RLock()
	defer pc.mu.RUnlock()
	return pc.isOutbound
}

func (pc *PeerConnection) LastPing() time.Time {
	pc.mu.RLock()
	defer pc.mu.RUnlock()
	return pc.lastPing
}

func (pc *PeerConnection) UpdatePing() {
	pc.mu.Lock()
	defer pc.mu.Unlock()
	pc.lastPing = time.Now()
}

func (pc *PeerConnection) Send(data []byte) error {
	select {
	case pc.send <- data:
		return nil
	case <-time.After(5 * time.Second):
		return fmt.Errorf("send timeout for peer %s", pc.ID)
	}
}

func (pc *PeerConnection) Close() error {
	close(pc.send)
	return pc.conn.Close()
}

func DialPeer(url string) (*websocket.Conn, error) {
	dialer := websocket.Dialer{
		HandshakeTimeout: 10 * time.Second,
	}

	conn, _, err := dialer.Dial(url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to dial peer: %w", err)
	}

	logrus.Infof("Successfully connected to peer: %s", url)
	return conn, nil
}
