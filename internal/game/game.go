package game

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"sort"
	"sync"

	"github.com/RedPaladin7/peerpoker/internal/blockchain"
	"github.com/RedPaladin7/peerpoker/internal/crypto"
	"github.com/RedPaladin7/peerpoker/internal/deck"
	"github.com/RedPaladin7/peerpoker/internal/protocol"
	"github.com/ethereum/go-ethereum/common"
	"github.com/sirupsen/logrus"
)

const (
	SmallBlind = 10
	BigBlind   = 20
)

type Game struct {
	lock          sync.RWMutex
	listenAddr    string
	broadcastFunc BroadcastFunc
	playerStates  map[string]*PlayerState
	rotationMap   map[int]string
	nextRotationID     int
	currentDealerID    int
	currentPlayerTurn  int
	currentStatus      GameStatus
	currentPot         int
	highestBet         int
	lastRaiserID       int
	lastRaiseAmount    int

	// Deck and cards
	deckKeys         *crypto.CardKeys
	foldedPlayerKeys map[string]*crypto.CardKeys
	revealedKeys     map[string]*crypto.CardKeys
	currentDeck      [][]byte
	myHand           []deck.Card
	communityCards   []deck.Card

	// Side pots
	sidePots []SidePot

	// Blockchain integration
	blockchain        *blockchain.BlockchainClient
	blockchainGameID  [32]byte
	blockchainEnabled bool

	// NEW: Disconnect handling
	DisconnectHandler *DisconnectHandler
}

type BroadcastFunc func(data []byte, targets ...string)

type SidePot struct {
	Amount          int
	Cap             int
	EligiblePlayers []string
}

func NewGame(addr string, broadcast BroadcastFunc, bc *blockchain.BlockchainClient) *Game {
	keys, _ := crypto.GenerateCardKeys()
	g := &Game{
		listenAddr:       addr,
		broadcastFunc:    broadcast,
		playerStates:     make(map[string]*PlayerState),
		rotationMap:      make(map[int]string),
		currentStatus:    GameStatusWaiting,
		deckKeys:         keys,
		foldedPlayerKeys: make(map[string]*crypto.CardKeys),
		revealedKeys:     make(map[string]*crypto.CardKeys),
		myHand:           make([]deck.Card, 0, 2),
		communityCards:   make([]deck.Card, 0, 5),
		sidePots:         []SidePot{},
		blockchain:       bc,
		blockchainEnabled: bc != nil,
	}

	// NEW: Initialize disconnect handler
	g.DisconnectHandler = NewDisconnectHandler(g)

	go g.loop()
	return g
}

func (g *Game) loop() {
	// Background processing if needed
	// Can be used for timeouts, periodic state sync, etc.
}

// GetStatus returns the current game status
func (g *Game) GetStatus() GameStatus {
	g.lock.RLock()
	defer g.lock.RUnlock()
	return g.currentStatus
}

func (g *Game) setStatus(status GameStatus) {
	g.currentStatus = status
	logrus.Infof("Game status changed to: %s", status.String())
}

// PlayerCount returns the number of players
func (g *Game) PlayerCount() int {
	g.lock.RLock()
	defer g.lock.RUnlock()
	return len(g.playerStates)
}

// ActivePlayerCount returns the number of active players
func (g *Game) ActivePlayerCount() int {
	g.lock.RLock()
	defer g.lock.RUnlock()
	count := 0
	for _, state := range g.playerStates {
		if state.IsActive {
			count++
		}
	}
	return count
}

// GetAllPlayers returns all player states
func (g *Game) GetAllPlayers() []PlayerStateResponse {
	g.lock.RLock()
	defer g.lock.RUnlock()

	players := make([]PlayerStateResponse, 0)
	for i := 0; i < g.nextRotationID; i++ {
		addr, ok := g.rotationMap[i]
		if !ok {
			continue
		}

		state, ok := g.playerStates[addr]
		if !ok {
			continue
		}

		players = append(players, PlayerStateResponse{
			PlayerID:      state.ListenAddr,
			RotationID:    state.RotationID,
			Stack:         state.Stack,
			CurrentBet:    state.CurrentRoundBet,
			IsActive:      state.IsActive,
			IsFolded:      state.IsFolded,
			IsAllIn:       state.IsAllIn,
			IsReady:       state.IsReady,
			IsDealer:      state.RotationID == g.currentDealerID,
			IsCurrentTurn: state.RotationID == g.currentPlayerTurn,
		})
	}
	return players
}

// GetTableState returns the table state for a specific client
func (g *Game) GetTableState(clientID string) TableStateResponse {
	g.lock.RLock()
	defer g.lock.RUnlock()

	myState, exists := g.playerStates[clientID]
	if !exists {
		return TableStateResponse{
			Status: g.currentStatus.String(),
		}
	}

	validActions := g.getValidActions(clientID)
	actionStrings := make([]string, len(validActions))
	for i, action := range validActions {
		actionStrings[i] = action.String()
	}

	myHandResp := make([]CardResponse, 0)
	if len(g.myHand) > 0 {
		myHandResp = make([]CardResponse, len(g.myHand))
		for i, card := range g.myHand {
			myHandResp[i] = CardResponse{
				Suit:    card.Suit.String(),
				Value:   card.Value,
				Display: card.String(),
			}
		}
	}

	communityCardResp := make([]CardResponse, len(g.communityCards))
	for i, card := range g.communityCards {
		communityCardResp[i] = CardResponse{
			Suit:    card.Suit.String(),
			Value:   card.Value,
			Display: card.String(),
		}
	}

	minRaise := g.highestBet + g.lastRaiseAmount
	if g.highestBet == 0 {
		minRaise = BigBlind
	}

	return TableStateResponse{
		Status:          g.currentStatus.String(),
		MyHand:          myHandResp,
		CommunityCards:  communityCardResp,
		Pot:             g.currentPot,
		HighestBet:      g.highestBet,
		MinRaise:        minRaise,
		ValidActions:    actionStrings,
		IsMyTurn:        myState.RotationID == g.currentPlayerTurn,
		MyStack:         myState.Stack,
		CurrentTurnID:   g.currentPlayerTurn,
		MyPlayerID:      myState.RotationID,
		DealerID:        g.currentDealerID,
		SmallBlind:      SmallBlind,
		BigBlind:        BigBlind,
	}
}

// HandleMessage processes incoming messages
func (g *Game) HandleMessage(from string, msg *protocol.Message) error {
	switch msg.Type {
	case protocol.TypePlayerReady:
		return g.handleMessageReady(from)
	case protocol.TypePlayerAction:
		var payload protocol.PlayerActionPayload
		if err := json.Unmarshal(msg.Payload, &payload); err != nil {
			return err
		}
		return g.handleMessagePlayerAction(from, payload)
	case protocol.TypePeerList:
		// Handle peer discovery
		return nil
	case protocol.TypeGameState:
		// Handle game state sync
		return nil
	default:
		logrus.Warnf("Unhandled message type: %s from %s", msg.Type, from)
	}
	return nil
}

func (g *Game) handleMessageReady(from string) error {
	logrus.Infof("Player %s is ready", from)
	return g.SetPlayerReady(from)
}

func (g *Game) handleMessagePlayerAction(from string, payload protocol.PlayerActionPayload) error {
	logrus.WithFields(logrus.Fields{
		"from":   from,
		"action": payload.Action,
		"value":  payload.Value,
	}).Info("Received player action")
	return g.HandlePlayerAction(from, payload.Action, payload.Value)
}

// Broadcast sends data to specified targets
func (g *Game) broadcast(data []byte, targets ...string) {
	if g.broadcastFunc != nil {
		g.broadcastFunc(data, targets...)
	}
}

// Send message to other players
func (g *Game) sendToPlayers(msgType protocol.MessageType, payload interface{}, targets ...string) error {
	msg, err := protocol.NewMessage(g.listenAddr, msgType, payload)
	if err != nil {
		return err
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	g.broadcast(data, targets...)
	return nil
}

// Get other players (excluding self)
func (g *Game) getOtherPlayers() []string {
	others := make([]string, 0)
	for addr := range g.playerStates {
		if addr != g.listenAddr {
			others = append(others, addr)
		}
	}
	return others
}

// Get ready players
func (g *Game) getReadyPlayers() []string {
	ready := make([]string, 0)
	for addr, state := range g.playerStates {
		if state.IsReady {
			ready = append(ready, addr)
		}
	}
	return ready
}

// Get ready and active players
func (g *Game) getReadyActivePlayers() []string {
	ready := make([]string, 0)
	for addr, state := range g.playerStates {
		if state.IsReady && state.IsActive {
			ready = append(ready, addr)
		}
	}
	return ready
}

// Get next player ID in rotation
func (g *Game) getNextPlayerID(currentID int) int {
	if g.nextRotationID == 0 {
		return 0
	}
	return (currentID + 1) % g.nextRotationID
}

// Get next active player ID
func (g *Game) getNextActivePlayerID(currentID int) int {
	startID := currentID
	for {
		nextID := g.getNextPlayerID(startID)
		addr, ok := g.rotationMap[nextID]
		if ok {
			state := g.playerStates[addr]
			if state.IsActive && !state.IsFolded {
				return nextID
			}
		}
		startID = nextID
		if startID == currentID {
			break
		}
	}
	return currentID
}

// Advance dealer button
func (g *Game) advanceDealer() {
	if g.nextRotationID == 0 {
		return
	}

	startID := g.currentDealerID
	for {
		nextID := (startID + 1) % g.nextRotationID
		addr, ok := g.rotationMap[nextID]
		if ok && g.playerStates[addr].IsActive {
			g.currentDealerID = nextID
			return
		}
		startID = nextID
		if startID == g.currentDealerID {
			break
		}
	}
}

// StartNewHand starts a new poker hand
func (g *Game) StartNewHand() {
	activeReadyPlayers := g.getReadyActivePlayers()
	if len(activeReadyPlayers) < 2 {
		g.setStatus(GameStatusWaiting)
		logrus.Warn("Not enough players to start a hand")
		return
	}

	// Blockchain: Create game on-chain
	if g.blockchainEnabled && g.blockchainGameID == [32]byte{} {
		buyIn := big.NewInt(int64(1000)) // Default 1000 wei buy-in
		smallBlind := big.NewInt(int64(SmallBlind))
		bigBlind := big.NewInt(int64(BigBlind))
		gameID, err := g.blockchain.CreateGame(buyIn, smallBlind, bigBlind, uint8(len(activeReadyPlayers)))
		if err != nil {
			logrus.Errorf("Failed to create game on blockchain: %v", err)
			// Continue without blockchain if it fails
		} else {
			g.blockchainGameID = gameID
			logrus.WithField("game_id", fmt.Sprintf("0x%x", gameID)).Info("Blockchain game created")
		}
	}

	// Blockchain: Verify all players have locked buy-ins
	if g.blockchainEnabled && g.blockchainGameID != [32]byte{} {
		allVerified := true
		for _, playerAddr := range activeReadyPlayers {
			addr := common.HexToAddress(playerAddr)
			verified, err := g.blockchain.VerifyBuyIn(g.blockchainGameID, addr)
			if err != nil || !verified {
				logrus.Warnf("Player %s buy-in not verified: %v", playerAddr, err)
				allVerified = false
			}
		}
		if !allVerified {
			logrus.Warn("Not all players have verified buy-ins, but continuing game...")
			// In production, you might want to reject game start here
		}
	}

	logrus.Info("=== Starting new hand ===")

	// Reset state
	g.rotationMap = make(map[int]string)
	g.nextRotationID = 0
	g.myHand = make([]deck.Card, 0, 2)
	g.communityCards = make([]deck.Card, 0, 5)
	g.lastRaiseAmount = BigBlind
	g.currentPot = 0
	g.highestBet = 0
	g.sidePots = []SidePot{}
	g.revealedKeys = make(map[string]*crypto.CardKeys)
	g.foldedPlayerKeys = make(map[string]*crypto.CardKeys)

	// Assign rotation IDs
	sort.Strings(activeReadyPlayers)
	for _, addr := range activeReadyPlayers {
		state := g.playerStates[addr]
		state.RotationID = g.nextRotationID
		state.IsFolded = false
		state.CurrentRoundBet = 0
		state.TotalBetThisHand = 0
		state.IsAllIn = false
		g.rotationMap[state.RotationID] = addr
		g.nextRotationID++
	}

	// Advance dealer
	g.advanceDealer()

	// Post blinds
	g.postBlinds()

	// Blockchain: Start game on-chain
	if g.blockchainEnabled && g.blockchainGameID != [32]byte{} {
		err := g.blockchain.StartGame(g.blockchainGameID)
		if err != nil {
			logrus.Errorf("Failed to start game on blockchain: %v", err)
		} else {
			logrus.Info("Game started on blockchain")
		}
	}

	// Set status
	g.setStatus(GameStatusDealing)

	// Start shuffle and deal
	g.InitiateShuffleAndDeal()
}

// Post blinds
func (g *Game) postBlinds() {
	activeCount := len(g.getReadyActivePlayers())
	if activeCount == 2 {
		// Heads-up: dealer posts small blind
		sbID := g.currentDealerID
		sbAddr := g.rotationMap[sbID]
		g.updatePlayerState(sbAddr, PlayerActionBet, SmallBlind)
		logrus.Infof("Player %s (dealer) posted small blind: %d", sbAddr, SmallBlind)

		bbID := g.getNextPlayerID(sbID)
		bbAddr := g.rotationMap[bbID]
		g.updatePlayerState(bbAddr, PlayerActionBet, BigBlind)
		logrus.Infof("Player %s posted big blind: %d", bbAddr, BigBlind)

		g.currentPlayerTurn = sbID
		g.lastRaiserID = bbID
	} else {
		// Multi-way: small blind is left of dealer
		sbID := g.getNextActivePlayerID(g.currentDealerID)
		sbAddr := g.rotationMap[sbID]
		g.updatePlayerState(sbAddr, PlayerActionBet, SmallBlind)
		logrus.Infof("Player %s posted small blind: %d", sbAddr, SmallBlind)

		bbID := g.getNextActivePlayerID(sbID)
		bbAddr := g.rotationMap[bbID]
		g.updatePlayerState(bbAddr, PlayerActionBet, BigBlind)
		logrus.Infof("Player %s posted big blind: %d", bbAddr, BigBlind)

		g.currentPlayerTurn = g.getNextActivePlayerID(bbID)
		g.lastRaiserID = bbID
	}

	g.lastRaiseAmount = BigBlind
}

// Update player state based on action
func (g *Game) updatePlayerState(addr string, action PlayerAction, value int) {
	state := g.playerStates[addr]

	switch action {
	case PlayerActionFold:
		state.IsFolded = true

	case PlayerActionBet, PlayerActionRaise:
		actualBet := value
		if actualBet > state.Stack {
			actualBet = state.Stack
			state.IsAllIn = true
			logrus.Infof("Player %s is ALL-IN!", addr)
		}

		amountToAdd := actualBet - state.CurrentRoundBet
		state.CurrentRoundBet = actualBet
		state.TotalBetThisHand += amountToAdd
		g.currentPot += amountToAdd
		state.Stack -= amountToAdd

		if state.CurrentRoundBet > g.highestBet {
			g.highestBet = state.CurrentRoundBet
			g.lastRaiserID = state.RotationID
		}

	case PlayerActionCall:
		amountNeeded := g.highestBet - state.CurrentRoundBet
		actualCall := amountNeeded
		if actualCall > state.Stack {
			actualCall = state.Stack
			state.IsAllIn = true
			logrus.Infof("Player %s is ALL-IN!", addr)
		}

		state.CurrentRoundBet += actualCall
		state.TotalBetThisHand += actualCall
		g.currentPot += actualCall
		state.Stack -= actualCall

	case PlayerActionCheck:
		// No state change
	}
}

// NEW: MonitorPlayerConnection monitors a player's connection
func (g *Game) MonitorPlayerConnection(playerID string) {
	g.lock.Lock()
	defer g.lock.Unlock()

	logrus.Warnf("⚠️  Monitoring disconnect for player %s", playerID)

	// Check if player exists
	state, exists := g.playerStates[playerID]
	if !exists {
		logrus.Warnf("Player %s not found in game", playerID)
		return
	}

	// Only handle disconnect if game is active
	if g.currentStatus != GameStatusInProgress && g.currentStatus != GameStatusDealing {
		logrus.Infof("Game not active, ignoring disconnect for %s", playerID)
		return
	}

	// Mark player as potentially disconnected
	state.IsActive = false

	// Run disconnect handler in goroutine
	go func() {
		ctx := context.Background()
		if err := g.DisconnectHandler.HandleDisconnect(ctx, playerID); err != nil {
			logrus.Errorf("Error handling disconnect for player %s: %v", playerID, err)
		}
	}()
}

// NEW: NotifyPlayerReconnected notifies that a player reconnected
func (g *Game) NotifyPlayerReconnected(playerID string) error {
	g.lock.Lock()
	defer g.lock.Unlock()

	logrus.Infof("✅ Player %s reconnected", playerID)

	// Restore player to active state
	state, exists := g.playerStates[playerID]
	if exists {
		state.IsActive = true
	}

	return g.DisconnectHandler.HandleReconnect(playerID)
}

// NEW: GetPlayer returns a player state by address
func (g *Game) GetPlayer(playerID string) *PlayerState {
	g.lock.RLock()
	defer g.lock.RUnlock()
	return g.playerStates[playerID]
}

// NEW: EndGameWithPenalty ends game with penalty to abandoned player
func (g *Game) EndGameWithPenalty(abandonedPlayerID string, remainingPlayers []*PlayerState) error {
	logrus.Warnf("💀 Ending game with penalty. Abandoned player: %s", abandonedPlayerID)

	abandonedPlayer := g.GetPlayer(abandonedPlayerID)
	if abandonedPlayer == nil {
		return fmt.Errorf("abandoned player %s not found", abandonedPlayerID)
	}

	// Prepare data for blockchain
	winners := make([]common.Address, 0)
	amounts := make([]*big.Int, 0)

	for _, player := range remainingPlayers {
		if player.Stack > 0 {
			// Convert player address string to common.Address
			winners = append(winners, common.HexToAddress(player.ListenAddr))
			
			// Convert chips to wei (assuming 1 chip = 0.001 ETH = 10^15 wei)
			amountWei := big.NewInt(int64(player.Stack))
			amountWei.Mul(amountWei, big.NewInt(1000000000000000)) // multiply by 10^15 wei
			amounts = append(amounts, amountWei)
		}
	}

	// Submit to blockchain if enabled
	if g.blockchainEnabled && g.blockchain != nil {
		logrus.Info("📝 Submitting penalty transaction to blockchain...")

		gameIDStr := fmt.Sprintf("%x", g.blockchainGameID[:])
		
		err := g.blockchain.EndGameWithPenalty(
			gameIDStr,
			common.HexToAddress(abandonedPlayer.ListenAddr),
			winners,
			amounts,
		)

		if err != nil {
			logrus.Errorf("Blockchain penalty submission failed: %v", err)
			return fmt.Errorf("blockchain penalty submission failed: %w", err)
		}

		logrus.Info("✅ Blockchain penalty transaction successful")
	}

	// Update game status
	g.setStatus(GameStatusFinished)

	return nil
}
