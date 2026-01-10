package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/RedPaladin7/peerpoker/internal/api"
	"github.com/RedPaladin7/peerpoker/internal/blockchain"
	"github.com/RedPaladin7/peerpoker/internal/config"
	"github.com/RedPaladin7/peerpoker/internal/game"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

type Server struct {
	listenAddr  string
	apiPort     string
	config      *config.Config
	hub         *WebSocketHub
	peerManager *PeerManager
	game        *game.Game
	blockchain  *blockchain.BlockchainClient
	mu          sync.RWMutex
	running     bool
}

func NewServer(cfg *config.Config) *Server {
	// Initialize blockchain client if enabled
	var bc *blockchain.BlockchainClient
	if os.Getenv("BLOCKCHAIN_ENABLED") == "true" {
		logrus.Info("Blockchain integration enabled, initializing client...")

		bcConfig := &blockchain.Config{
			RPCURL:                 os.Getenv("BLOCKCHAIN_RPC_URL"),
			PrivateKey:             os.Getenv("BLOCKCHAIN_PRIVATE_KEY"),
			PokerTableAddress:      os.Getenv("CONTRACT_POKER_TABLE"),
			PotManagerAddress:      os.Getenv("CONTRACT_POT_MANAGER"),
			PlayerRegistryAddress:  os.Getenv("CONTRACT_PLAYER_REGISTRY"),
			DisputeResolverAddress: os.Getenv("CONTRACT_DISPUTE_RESOLVER"),
		}

		var err error
		bc, err = blockchain.NewBlockchainClient(bcConfig)
		if err != nil {
			logrus.Warnf("Failed to initialize blockchain client: %v", err)
			logrus.Warn("Continuing without blockchain integration")
			bc = nil
		} else {
			logrus.Info("✅ Blockchain client initialized successfully")

			// Log blockchain info
			balance, err := bc.GetMyBalance()
			if err == nil {
				logrus.WithField("balance", blockchain.ConvertFromWei(balance)).Info("Wallet balance")
			}
		}
	} else {
		logrus.Info("Blockchain integration disabled")
	}

	s := &Server{
		listenAddr: cfg.ListenAddr,
		apiPort:    cfg.APIPort,
		config:     cfg,
		blockchain: bc,
	}

	s.hub = NewWebSocketHub(s)
	s.peerManager = NewPeerManager(s)

	// Pass blockchain client to game
	s.game = game.NewGame(cfg.ListenAddr, s.broadcastToPlayers, bc)

	return s
}

func (s *Server) Start() error {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return fmt.Errorf("server already running")
	}
	s.running = true
	s.mu.Unlock()

	logrus.WithFields(logrus.Fields{
		"ws_addr":  s.listenAddr,
		"api_port": s.apiPort,
	}).Info("Starting poker server")

	// Start WebSocket hub
	go s.hub.Run()

	// Start peer manager
	go s.peerManager.Run()

	// Start WebSocket server
	go s.startWebSocketServer()

	// Start HTTP API server
	return s.startAPIServer()
}

func (s *Server) startWebSocketServer() {
	router := mux.NewRouter()

	// WebSocket endpoint for clients
	router.HandleFunc("/ws", s.handleWebSocket)

	// WebSocket endpoint for peers
	router.HandleFunc("/p2p", s.handlePeerConnection)

	addr := s.listenAddr
	logrus.Infof("WebSocket server listening on %s", addr)

	server := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  60 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	if err := server.ListenAndServe(); err != nil {
		logrus.Fatalf("WebSocket server failed: %v", err)
	}
}

func (s *Server) startAPIServer() error {
	router := mux.NewRouter()

	// Create API handler
	apiHandler := api.NewHandler(s.game)

	// Setup routes
	api.SetupRoutes(router, apiHandler)

	// Add middleware
	router.Use(api.LoggingMiddleware)
	router.Use(api.CORSMiddleware)

	addr := fmt.Sprintf(":%s", s.apiPort)
	logrus.Infof("HTTP API server listening on %s", addr)

	server := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  60 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	return server.ListenAndServe()
}

func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	client, err := NewClient(s.hub, w, r)
	if err != nil {
		logrus.Errorf("Failed to create client: %v", err)
		return
	}

	s.hub.register <- client

	go client.WritePump()
	go client.ReadPump()
}

func (s *Server) handlePeerConnection(w http.ResponseWriter, r *http.Request) {
	peer, err := s.peerManager.HandleIncomingPeer(w, r)
	if err != nil {
		logrus.Errorf("Failed to handle peer connection: %v", err)
		return
	}

	go peer.ReadPump()
	go peer.WritePump()
}

func (s *Server) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return
	}

	logrus.Info("Stopping server...")

	// Close blockchain client
	if s.blockchain != nil {
		logrus.Info("Closing blockchain client...")
		s.blockchain.Close()
		logrus.Info("Blockchain client closed")
	}

	s.running = false
	logrus.Info("Server stopped")
}

func (s *Server) ConnectToPeer(peerAddr string) error {
	return s.peerManager.ConnectToPeer(peerAddr)
}

func (s *Server) broadcastToPlayers(data []byte, targets ...string) {
	if len(targets) == 0 {
		// Broadcast to all clients
		s.hub.broadcast <- data
	} else {
		// Send to specific targets
		for _, target := range targets {
			s.hub.sendToClient(target, data)
		}
	}
}

func (s *Server) GetGame() *game.Game {
	return s.game
}

func (s *Server) GetPeerManager() *PeerManager {
	return s.peerManager
}

func (s *Server) GetHub() *WebSocketHub {
	return s.hub
}

// GetBlockchainClient returns the blockchain client (can be nil)
func (s *Server) GetBlockchainClient() *blockchain.BlockchainClient {
	return s.blockchain
}

// IsBlockchainEnabled returns whether blockchain integration is active
func (s *Server) IsBlockchainEnabled() bool {
	return s.blockchain != nil
}
