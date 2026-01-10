package game

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	// DisconnectTimeout is how long to wait before declaring player abandoned
	DisconnectTimeout = 5 * time.Minute
)

// DisconnectHandler manages simple disconnect detection and timeout
type DisconnectHandler struct {
	game              *Game
	disconnectTimers  map[string]*time.Timer
	reconnectChannels map[string]chan bool
	mu                sync.RWMutex
	logger            *logrus.Logger
}

// NewDisconnectHandler creates a new disconnect handler
func NewDisconnectHandler(game *Game) *DisconnectHandler {
	return &DisconnectHandler{
		game:              game,
		disconnectTimers:  make(map[string]*time.Timer),
		reconnectChannels: make(map[string]chan bool),
		logger:            game.Logger,
	}
}

// HandleDisconnect handles a player disconnection with timeout
func (dh *DisconnectHandler) HandleDisconnect(ctx context.Context, playerID string) error {
	dh.mu.Lock()

	// Check if already handling this disconnect
	if _, exists := dh.disconnectTimers[playerID]; exists {
		dh.mu.Unlock()
		return fmt.Errorf("disconnect already being handled for player %s", playerID)
	}

	// Create reconnect channel
	dh.reconnectChannels[playerID] = make(chan bool, 1)
	dh.mu.Unlock()

	dh.logger.Warnf("⚠️  Player %s disconnected. Starting %v timeout...", playerID, DisconnectTimeout)

	// Start timeout timer
	timer := time.NewTimer(DisconnectTimeout)
	dh.mu.Lock()
	dh.disconnectTimers[playerID] = timer
	dh.mu.Unlock()

	select {
	case <-timer.C:
		// Timeout reached - player abandoned game
		dh.logger.Errorf("❌ Player %s abandoned game (timeout reached)", playerID)
		return dh.handleAbandon(playerID)

	case <-dh.reconnectChannels[playerID]:
		// Player reconnected in time
		timer.Stop()
		dh.mu.Lock()
		delete(dh.disconnectTimers, playerID)
		delete(dh.reconnectChannels, playerID)
		dh.mu.Unlock()

		dh.logger.Infof("✅ Player %s reconnected successfully", playerID)
		return nil

	case <-ctx.Done():
		// Context cancelled
		timer.Stop()
		return ctx.Err()
	}
}

// HandleReconnect handles a player reconnection
func (dh *DisconnectHandler) HandleReconnect(playerID string) error {
	dh.mu.Lock()
	defer dh.mu.Unlock()

	ch, exists := dh.reconnectChannels[playerID]
	if !exists {
		return fmt.Errorf("no disconnect timer for player %s", playerID)
	}

	// Signal reconnection
	select {
	case ch <- true:
		dh.logger.Infof("Signaled reconnection for player %s", playerID)
	default:
		// Channel already closed or full
	}

	return nil
}

// handleAbandon handles when a player abandons (timeout reached)
func (dh *DisconnectHandler) handleAbandon(playerID string) error {
	// Clean up timers
	dh.mu.Lock()
	delete(dh.disconnectTimers, playerID)
	delete(dh.reconnectChannels, playerID)
	dh.mu.Unlock()

	// Mark player as abandoned
	player := dh.game.GetPlayer(playerID)
	if player == nil {
		return fmt.Errorf("player %s not found", playerID)
	}
	player.Status = PlayerAbandoned

	dh.logger.Warnf("💀 Player %s abandoned. Aborting game and applying penalty...", playerID)

	// Abort game and distribute penalty
	return dh.abortGameWithPenalty(playerID)
}

// abortGameWithPenalty aborts the game and penalizes the abandoned player
func (dh *DisconnectHandler) abortGameWithPenalty(abandonedPlayerID string) error {
	dh.logger.Warnf("🚫 Aborting game. Applying penalty to %s", abandonedPlayerID)

	abandonedPlayer := dh.game.GetPlayer(abandonedPlayerID)
	if abandonedPlayer == nil {
		return fmt.Errorf("player %s not found", abandonedPlayerID)
	}

	// Get remaining active players
	remainingPlayers := make([]*Player, 0)
	for _, p := range dh.game.Players {
		if p.ID != abandonedPlayerID && p.Status == PlayerActive {
			remainingPlayers = append(remainingPlayers, p)
		}
	}

	if len(remainingPlayers) == 0 {
		return fmt.Errorf("no remaining players to distribute penalty")
	}

	// Calculate penalty distribution (split abandoned player's buy-in)
	penaltyPerPlayer := abandonedPlayer.BuyIn / len(remainingPlayers)
	remainder := abandonedPlayer.BuyIn % len(remainingPlayers)

	dh.logger.Infof("💰 Distributing %d chips penalty from %s to %d remaining players",
		abandonedPlayer.BuyIn, abandonedPlayerID, len(remainingPlayers))

	// Each remaining player gets their buy-in back + share of penalty
	for i, p := range remainingPlayers {
		// Refund original buy-in + penalty share
		p.Stack = p.BuyIn + penaltyPerPlayer
		if i == 0 {
			p.Stack += remainder // Give remainder to first player
		}
		dh.logger.Infof("  → Player %s: %d chips (buy-in) + %d (penalty) = %d total",
			p.ID, p.BuyIn, penaltyPerPlayer, p.Stack)
	}

	// Abandoned player gets nothing (loses entire buy-in)
	abandonedPlayer.Stack = 0
	dh.logger.Errorf("  → Player %s: 0 chips (PENALTY - lost %d)", abandonedPlayerID, abandonedPlayer.BuyIn)

	// Mark game as aborted
	dh.game.Status = GameAborted

	// Submit to blockchain
	if dh.game.BlockchainClient != nil {
		dh.logger.Info("📝 Submitting penalty to blockchain...")
		return dh.game.EndGameWithPenalty(abandonedPlayerID, remainingPlayers)
	}

	dh.logger.Info("✅ Game aborted successfully. Penalty applied.")
	return nil
}

// GetStatus returns current disconnect handler status
func (dh *DisconnectHandler) GetStatus() map[string]interface{} {
	dh.mu.RLock()
	defer dh.mu.RUnlock()

	activeTimers := make([]string, 0)
	for playerID := range dh.disconnectTimers {
		activeTimers = append(activeTimers, playerID)
	}

	return map[string]interface{}{
		"active_disconnect_timers": activeTimers,
		"num_timers":               len(activeTimers),
	}
}
