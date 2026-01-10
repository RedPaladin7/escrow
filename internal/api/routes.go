package api

import (
	"net/http"

	"github.com/gorilla/mux"
)

func (h *Handler) Routes() http.Handler {
	r := mux.NewRouter()

	// Apply middleware
	r.Use(CORSMiddleware)
	r.Use(LoggingMiddleware)
	r.Use(RecoveryMiddleware)

	// Health check
	r.HandleFunc("/api/health", h.HandleHealth).Methods("GET", "OPTIONS")

	// Game state endpoints
	r.HandleFunc("/api/table", h.HandleGetTable).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/players", h.HandleGetPlayers).Methods("GET", "OPTIONS")

	// Player actions
	r.HandleFunc("/api/ready", h.HandlePlayerReady).Methods("POST", "OPTIONS")
	r.HandleFunc("/api/fold", h.HandleFold).Methods("POST", "OPTIONS")
	r.HandleFunc("/api/check", h.HandleCheck).Methods("POST", "OPTIONS")
	r.HandleFunc("/api/call", h.HandleCall).Methods("POST", "OPTIONS")
	r.HandleFunc("/api/bet", h.HandleBet).Methods("POST", "OPTIONS")
	r.HandleFunc("/api/raise", h.HandleRaise).Methods("POST", "OPTIONS")
	r.HandleFunc("/api/action", h.HandlePlayerAction).Methods("POST", "OPTIONS")

	// Peer management
	r.HandleFunc("/api/peers", h.HandleGetPeers).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/peers/connect", h.HandleConnectPeer).Methods("POST", "OPTIONS")

	return r
}
