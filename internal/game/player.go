package game

import (
	"fmt"
	"sort"

	"github.com/sirupsen/logrus"
)

type PlayerState struct {
	ListenAddr       string
	RotationID       int
	IsReady          bool
	IsActive         bool
	IsFolded         bool
	CurrentRoundBet  int
	IsAllIn          bool
	Stack            int
	TotalBetThisHand int
}

type PlayerStateResponse struct {
	PlayerID      string `json:"player_id"`
	RotationID    int    `json:"rotation_id"`
	Stack         int    `json:"stack"`
	CurrentBet    int    `json:"current_bet"`
	IsActive      bool   `json:"is_active"`
	IsFolded      bool   `json:"is_folded"`
	IsAllIn       bool   `json:"is_all_in"`
	IsReady       bool   `json:"is_ready"`
	IsDealer      bool   `json:"is_dealer"`
	IsCurrentTurn bool   `json:"is_current_turn"`
}

type TableStateResponse struct {
	Status         string         `json:"status"`
	MyHand         []CardResponse `json:"my_hand"`
	CommunityCards []CardResponse `json:"community_cards"`
	Pot            int            `json:"pot"`
	HighestBet     int            `json:"highest_bet"`
	MinRaise       int            `json:"min_raise"`
	ValidActions   []string       `json:"valid_actions"`
	IsMyTurn       bool           `json:"is_my_turn"`
	MyStack        int            `json:"my_stack"`
	CurrentTurnID  int            `json:"current_turn_id"`
	MyPlayerID     int            `json:"my_player_id"`
	DealerID       int            `json:"dealer_id"`
	SmallBlind     int            `json:"small_blind"`
	BigBlind       int            `json:"big_blind"`
}

type CardResponse struct {
	Suit    string `json:"suit"`
	Value   int    `json:"value"`
	Display string `json:"display"`
}

// AddPlayer adds a new player to the game
func (g *Game) AddPlayer(addr string) {
	g.lock.Lock()
	defer g.lock.Unlock()

	if _, exists := g.playerStates[addr]; exists {
		g.playerStates[addr].IsActive = true
		logrus.Infof("Player %s reconnected", addr)
		return
	}

	g.playerStates[addr] = &PlayerState{
		ListenAddr: addr,
		IsActive:   true,
		Stack:      1000,
	}

	logrus.Infof("Player %s added to game", addr)
}

// RemovePlayer removes a player from the game
func (g *Game) RemovePlayer(addr string) {
	g.lock.Lock()
	defer g.lock.Unlock()

	if state, ok := g.playerStates[addr]; ok {
		state.IsActive = false
		state.IsFolded = true
		logrus.Infof("Player %s removed from game", addr)

		// Check if we need to end the hand
		if g.currentStatus != GameStatusWaiting {
			g.checkRoundEnd()
		}
	}
}

// SetPlayerReady marks a player as ready
func (g *Game) SetPlayerReady(addr string) error {
	g.lock.Lock()
	defer g.lock.Unlock()

	state, ok := g.playerStates[addr]
	if !ok {
		return fmt.Errorf("player %s not found", addr)
	}

	if !state.IsReady {
		state.RotationID = g.nextRotationID
		g.rotationMap[state.RotationID] = addr
		g.nextRotationID++
		state.IsReady = true
		logrus.Infof("Player %s is ready (Rotation ID: %d)", addr, state.RotationID)
	}

	// Broadcast ready status
	g.sendToPlayers(protocol.TypePlayerReady, protocol.PlayerReadyPayload{
		PlayerID: addr,
	}, g.getOtherPlayers()...)

	// Check if we can start the game
	if len(g.getReadyPlayers()) >= 2 && g.currentStatus == GameStatusWaiting {
		g.StartNewHand()
	}

	return nil
}

// StartNewHand starts a new poker hand
func (g *Game) StartNewHand() {
	activeReadyPlayers := g.getReadyActivePlayers()
	if len(activeReadyPlayers) < 2 {
		g.setStatus(GameStatusWaiting)
		logrus.Warn("Not enough players to start a hand")
		return
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
