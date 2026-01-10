package api

import (
	"encoding/json"
	"net/http"

	"github.com/RedPaladin7/peerpoker/internal/game"
	"github.com/sirupsen/logrus"
)

type Handler struct {
	game        *game.Game
	peerManager PeerManager
	hub         Hub
}

type PeerManager interface {
	PeerCount() int
	GetAllPeerIDs() []string
}

type Hub interface {
	ClientCount() int
	GetClientIDs() []string
	Broadcast(data []byte, targets ...string)
}

func NewHandler(g *game.Game, pm PeerManager, hub Hub) *Handler {
	return &Handler{
		game:        g,
		peerManager: pm,
		hub:         hub,
	}
}

// Health check endpoint
func (h *Handler) HandleHealth(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"status":      "healthy",
		"game_status": h.game.GetStatus().String(),
		"players":     h.game.PlayerCount(),
		"peers":       h.peerManager.PeerCount(),
		"ws_clients":  h.hub.ClientCount(),
	}
	JSON(w, http.StatusOK, response)
}

// Get table state for a specific client
func (h *Handler) HandleGetTable(w http.ResponseWriter, r *http.Request) {
	clientID := r.Header.Get("X-Client-ID")
	if clientID == "" {
		http.Error(w, "Client ID required", http.StatusBadRequest)
		return
	}

	tableState := h.game.GetTableState(clientID)
	JSON(w, http.StatusOK, tableState)
}

// Get all players
func (h *Handler) HandleGetPlayers(w http.ResponseWriter, r *http.Request) {
	players := h.game.GetAllPlayers()
	response := map[string]interface{}{
		"players":        players,
		"total_players":  len(players),
		"active_players": h.game.ActivePlayerCount(),
	}
	JSON(w, http.StatusOK, response)
}

// Handle player action (fold, check, call, bet, raise)
func (h *Handler) HandlePlayerAction(w http.ResponseWriter, r *http.Request) {
	clientID := r.Header.Get("X-Client-ID")
	if clientID == "" {
		http.Error(w, "Client ID required", http.StatusBadRequest)
		return
	}

	var req struct {
		Action string `json:"action"`
		Value  int    `json:"value,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.game.HandlePlayerAction(clientID, req.Action, req.Value); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	JSON(w, http.StatusOK, map[string]string{"status": "success"})
}

// Set player ready
func (h *Handler) HandlePlayerReady(w http.ResponseWriter, r *http.Request) {
	clientID := r.Header.Get("X-Client-ID")
	if clientID == "" {
		http.Error(w, "Client ID required", http.StatusBadRequest)
		return
	}

	if err := h.game.SetPlayerReady(clientID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	JSON(w, http.StatusOK, map[string]string{
		"status": "ready",
		"player": clientID,
	})
}

// Handle fold action
func (h *Handler) HandleFold(w http.ResponseWriter, r *http.Request) {
	clientID := r.Header.Get("X-Client-ID")
	if clientID == "" {
		http.Error(w, "Client ID required", http.StatusBadRequest)
		return
	}

	if err := h.game.HandlePlayerAction(clientID, "fold", 0); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	JSON(w, http.StatusOK, map[string]string{
		"status": "fold",
		"player": clientID,
	})
}

// Handle check action
func (h *Handler) HandleCheck(w http.ResponseWriter, r *http.Request) {
	clientID := r.Header.Get("X-Client-ID")
	if clientID == "" {
		http.Error(w, "Client ID required", http.StatusBadRequest)
		return
	}

	if err := h.game.HandlePlayerAction(clientID, "check", 0); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	JSON(w, http.StatusOK, map[string]string{
		"status": "check",
		"player": clientID,
	})
}

// Handle call action
func (h *Handler) HandleCall(w http.ResponseWriter, r *http.Request) {
	clientID := r.Header.Get("X-Client-ID")
	if clientID == "" {
		http.Error(w, "Client ID required", http.StatusBadRequest)
		return
	}

	if err := h.game.HandlePlayerAction(clientID, "call", 0); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	JSON(w, http.StatusOK, map[string]string{
		"status": "call",
		"player": clientID,
	})
}

// Handle bet action
func (h *Handler) HandleBet(w http.ResponseWriter, r *http.Request) {
	clientID := r.Header.Get("X-Client-ID")
	if clientID == "" {
		http.Error(w, "Client ID required", http.StatusBadRequest)
		return
	}

	var req struct {
		Value int `json:"value"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.game.HandlePlayerAction(clientID, "bet", req.Value); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	JSON(w, http.StatusOK, map[string]interface{}{
		"status": "bet",
		"player": clientID,
		"value":  req.Value,
	})
}

// Handle raise action
func (h *Handler) HandleRaise(w http.ResponseWriter, r *http.Request) {
	clientID := r.Header.Get("X-Client-ID")
	if clientID == "" {
		http.Error(w, "Client ID required", http.StatusBadRequest)
		return
	}

	var req struct {
		Value int `json:"value"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.game.HandlePlayerAction(clientID, "raise", req.Value); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	JSON(w, http.StatusOK, map[string]interface{}{
		"status": "raise",
		"player": clientID,
		"value":  req.Value,
	})
}

// Get connected peers
func (h *Handler) HandleGetPeers(w http.ResponseWriter, r *http.Request) {
	peerIDs := h.peerManager.GetAllPeerIDs()
	response := map[string]interface{}{
		"peers": peerIDs,
		"count": len(peerIDs),
	}
	JSON(w, http.StatusOK, response)
}

// Connect to a new peer
func (h *Handler) HandleConnectPeer(w http.ResponseWriter, r *http.Request) {
	var req struct {
		PeerAddr string `json:"peer_addr"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.PeerAddr == "" {
		http.Error(w, "peer_addr is required", http.StatusBadRequest)
		return
	}

	// TODO: Implement peer connection logic
	logrus.Infof("Received request to connect to peer: %s", req.PeerAddr)

	JSON(w, http.StatusOK, map[string]string{
		"status":  "success",
		"message": "Connection request initiated",
	})
}

// JSON helper function
func JSON(w http.ResponseWriter, status int, v interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(v)
}
