package server

import (
	"fmt"
	"sync"

	"github.com/sirupsen/logrus"
)

type PeerManager struct {
	peers      map[string]*Client
	maxPeers   int
	mu         sync.RWMutex
}

func NewPeerManager(maxPeers int) *PeerManager {
	return &PeerManager{
		peers:    make(map[string]*Client),
		maxPeers: maxPeers,
	}
}

func (pm *PeerManager) AddPeer(client *Client) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	
	if len(pm.peers) >= pm.maxPeers {
		return fmt.Errorf("maximum peers (%d) reached", pm.maxPeers)
	}
	
	if _, exists := pm.peers[client.ID]; exists {
		return fmt.Errorf("peer %s already exists", client.ID)
	}
	
	pm.peers[client.ID] = client
	logrus.Infof("Peer added: %s (total: %d)", client.ID, len(pm.peers))
	return nil
}

func (pm *PeerManager) RemovePeer(clientID string) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	
	if _, exists := pm.peers[clientID]; exists {
		delete(pm.peers, clientID)
		logrus.Infof("Peer removed: %s (total: %d)", clientID, len(pm.peers))
	}
}

func (pm *PeerManager) GetPeer(clientID string) (*Client, bool) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	
	peer, exists := pm.peers[clientID]
	return peer, exists
}

func (pm *PeerManager) PeerCount() int {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return len(pm.peers)
}

func (pm *PeerManager) GetAllPeerIDs() []string {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	
	ids := make([]string, 0, len(pm.peers))
	for id := range pm.peers {
		ids = append(ids, id)
	}
	return ids
}

func (pm *PeerManager) ConnectToPeer(peerAddr string, hub *WebSocketHub) error {
	// TODO: Implement outbound WebSocket connection to peer
	// This will be used for P2P mesh networking
	logrus.Infof("Attempting to connect to peer: %s", peerAddr)
	
	// For now, we'll use passive connections only (peers connect to us)
	// Future enhancement: dial out to peer WebSocket endpoints
	
	return nil
}
