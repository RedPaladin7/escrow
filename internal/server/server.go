package server

import (
	"context"
	"net/http"
	"sync"

	"github.com/RedPaladin7/peerpoker/internal/api"
	"github.com/RedPaladin7/peerpoker/internal/config"
	"github.com/RedPaladin7/peerpoker/internal/game"
	"github.com/sirupsen/logrus"
)

type Server struct {
	config      *config.Config
	hub         *WebSocketHub
	peerManager *PeerManager
	game        *game.Game
	apiServer   *http.Server
	wsServer    *http.Server
	wg          sync.WaitGroup
}

func NewServer(cfg *config.Config) (*Server, error) {
	hub := NewWebSocketHub()
	peerManager := NewPeerManager(cfg.MaxPlayers)
	
	gameInstance := game.NewGame(cfg.WSPort, hub.Broadcast)

	s := &Server{
		config:      cfg,
		hub:         hub,
		peerManager: peerManager,
		game:        gameInstance,
	}

	apiHandler := api.NewHandler(s.game, s.peerManager, s.hub)
	
	s.apiServer = &http.Server{
		Addr:    cfg.GetAPIAddr(),
		Handler: apiHandler.Routes(),
	}

	wsHandler := s.createWSHandler()
	s.wsServer = &http.Server{
		Addr:    cfg.GetWSAddr(),
		Handler: wsHandler,
	}

	return s, nil
}

func (s *Server) Start(ctx context.Context) error {
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		s.hub.Run(ctx)
	}()

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		logrus.Infof("API server listening on %s", s.apiServer.Addr)
		if err := s.apiServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logrus.Errorf("API server error: %v", err)
		}
	}()

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		logrus.Infof("WebSocket server listening on %s", s.wsServer.Addr)
		if err := s.wsServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logrus.Errorf("WebSocket server error: %v", err)
		}
	}()

	if s.config.InitialPeer != "" {
		s.wg.Add(1)
		go func() {
			defer s.wg.Done()
			if err := s.peerManager.ConnectToPeer(s.config.InitialPeer, s.hub); err != nil {
				logrus.Errorf("Failed to connect to initial peer: %v", err)
			}
		}()
	}

	<-ctx.Done()
	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	logrus.Info("Shutting down servers...")

	errChan := make(chan error, 2)
	
	go func() {
		errChan <- s.apiServer.Shutdown(ctx)
	}()
	
	go func() {
		errChan <- s.wsServer.Shutdown(ctx)
	}()

	var lastErr error
	for i := 0; i < 2; i++ {
		if err := <-errChan; err != nil {
			logrus.Errorf("Shutdown error: %v", err)
			lastErr = err
		}
	}

	s.hub.Close()
	s.wg.Wait()

	return lastErr
}

func (s *Server) createWSHandler() http.Handler {
	mux := http.NewServeMux()
	
	mux.HandleFunc("/ws", s.handleWebSocket)
	mux.HandleFunc("/p2p", s.handleP2PConnection)
	
	return api.CORSMiddleware(mux)
}

func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	client, err := NewClientFromHTTP(w, r, s.hub, s.game, false)
	if err != nil {
		logrus.Errorf("WebSocket upgrade failed: %v", err)
		http.Error(w, "WebSocket upgrade failed", http.StatusBadRequest)
		return
	}

	s.hub.Register <- client
	
	go client.WritePump()
	go client.ReadPump()
}

func (s *Server) handleP2PConnection(w http.ResponseWriter, r *http.Request) {
	if s.peerManager.PeerCount() >= s.config.MaxPlayers {
		http.Error(w, "Maximum peers reached", http.StatusServiceUnavailable)
		return
	}

	client, err := NewClientFromHTTP(w, r, s.hub, s.game, true)
	if err != nil {
		logrus.Errorf("P2P WebSocket upgrade failed: %v", err)
		http.Error(w, "WebSocket upgrade failed", http.StatusBadRequest)
		return
	}

	if err := s.peerManager.AddPeer(client); err != nil {
		logrus.Errorf("Failed to add peer: %v", err)
		client.Close()
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	s.hub.Register <- client
	
	go client.WritePump()
	go client.ReadPump()
}
