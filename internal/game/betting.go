package game

import (
	"github.com/sirupsen/logrus"
)

// advanceTurnAndCheckRoundEnd advances to the next player and checks if round is over
func (g *Game) advanceTurnAndCheckRoundEnd() {
	if g.checkRoundEnd() {
		g.advanceToNextRound()
		return
	}

	g.incNextPlayer()

	if g.checkRoundEnd() {
		g.advanceToNextRound()
	}
}

// incNextPlayer moves to the next active player
func (g *Game) incNextPlayer() {
	startID := g.currentPlayerTurn
	attempts := 0
	maxAttempts := g.nextRotationID + 1

	for attempts < maxAttempts {
		nextID := g.getNextPlayerID(startID)
		addr := g.rotationMap[nextID]
		state, ok := g.playerStates[addr]

		if ok && state.IsActive && !state.IsFolded && !state.IsAllIn {
			g.currentPlayerTurn = nextID
			return
		}

		startID = nextID
		attempts++
	}

	logrus.Warn("No active players found who can act")
}

// checkRoundEnd checks if the betting round is complete
func (g *Game) checkRoundEnd() bool {
	activeNonFoldedCount := 0
	canActCount := 0

	for _, state := range g.playerStates {
		if state.IsActive && !state.IsFolded {
			activeNonFoldedCount++
			if !state.IsAllIn {
				canActCount++
			}
		}
	}

	// Only one player left (everyone else folded)
	if activeNonFoldedCount <= 1 {
		return true
	}

	// All remaining players are all-in
	if canActCount == 0 {
		logrus.Info("All remaining players are all-in, advancing to showdown")
		return true
	}

	// Only one player can act and they've matched the bet
	if canActCount == 1 {
		allMatched := true
		for _, state := range g.playerStates {
			if state.IsActive && !state.IsFolded && !state.IsAllIn {
				if state.CurrentRoundBet < g.highestBet {
					allMatched = false
					break
				}
			}
		}
		if allMatched {
			return true
		}
	}

	// All players have matched the highest bet
	allMatchedOrOut := true
	for _, state := range g.playerStates {
		if state.IsActive && !state.IsFolded && !state.IsAllIn {
			if state.CurrentRoundBet < g.highestBet {
				allMatchedOrOut = false
				break
			}
		}
	}

	if allMatchedOrOut {
		// Check if we're back to the last raiser
		nextToAct := g.getNextActivePlayerID(g.lastRaiserID)
		if g.currentPlayerTurn == nextToAct {
			return true
		}
	}

	return false
}

// advanceToNextRound moves to the next betting round
func (g *Game) advanceToNextRound() {
	logrus.Infof("=== Advancing from %s ===", g.currentStatus.String())

	// Reset betting for new round
	for _, state := range g.playerStates {
		state.CurrentRoundBet = 0
	}
	g.highestBet = 0

	switch g.currentStatus {
	case GameStatusDealing:
		g.setStatus(GameStatusPreFlop)
		// Cards are dealt, ready for pre-flop betting

	case GameStatusPreFlop:
		g.setStatus(GameStatusFlop)
		g.dealFlop()

	case GameStatusFlop:
		g.setStatus(GameStatusTurn)
		g.dealTurn()

	case GameStatusTurn:
		g.setStatus(GameStatusRiver)
		g.dealRiver()

	case GameStatusRiver:
		g.setStatus(GameStatusShowdown)
		g.ResolveWinner()

	case GameStatusShowdown:
		g.resetHandState()
	}
}

// dealFlop deals the flop (3 community cards)
func (g *Game) dealFlop() {
	logrus.Info("Dealing flop (3 cards)...")
	g.dealCommunityCards(3)
	
	logrus.Infof("Flop: %v", g.communityCards)
	
	// Reset turn to first active player after dealer
	g.currentPlayerTurn = g.getNextActivePlayerID(g.currentDealerID)
	
	// Broadcast flop to all players
	g.broadcastCommunityCards("flop")
}

// dealTurn deals the turn (4th community card)
func (g *Game) dealTurn() {
	logrus.Info("Dealing turn (1 card)...")
	g.dealCommunityCards(1)
	
	logrus.Infof("Turn: %s", g.communityCards[3].String())
	
	// Reset turn to first active player after dealer
	g.currentPlayerTurn = g.getNextActivePlayerID(g.currentDealerID)
	
	// Broadcast turn to all players
	g.broadcastCommunityCards("turn")
}

// dealRiver deals the river (5th community card)
func (g *Game) dealRiver() {
	logrus.Info("Dealing river (1 card)...")
	g.dealCommunityCards(1)
	
	logrus.Infof("River: %s", g.communityCards[4].String())
	
	// Reset turn to first active player after dealer
	g.currentPlayerTurn = g.getNextActivePlayerID(g.currentDealerID)
	
	// Broadcast river to all players
	g.broadcastCommunityCards("river")
}

// broadcastCommunityCards broadcasts community cards to all players
func (g *Game) broadcastCommunityCards(stage string) {
	cards := make([]CardResponse, len(g.communityCards))
	for i, card := range g.communityCards {
		cards[i] = CardResponse{
			Suit:    card.Suit.String(),
			Value:   card.Value,
			Display: card.String(),
		}
	}

	// Send to all players via protocol
	// This will be picked up by the frontend
	logrus.WithFields(logrus.Fields{
		"stage": stage,
		"cards": len(cards),
	}).Info("Broadcasting community cards")
}
