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

// resetHandState resets the game state for a new hand
func (g *Game) resetHandState() {
	logrus.Info("=== Resetting for new hand ===")

	g.currentPot = 0
	g.highestBet = 0
	g.lastRaiseAmount = BigBlind
	g.myHand = make([]deck.Card, 0, 2)
	g.communityCards = make([]deck.Card, 0, 5)
	g.currentDeck = nil
	g.sidePots = []SidePot{}
	g.revealedKeys = make(map[string]*crypto.CardKeys)
	g.foldedPlayerKeys = make(map[string]*crypto.CardKeys)

	// Remove players with no chips
	for addr, state := range g.playerStates {
		if state.Stack <= 0 {
			state.IsActive = false
			logrus.Infof("Player %s eliminated (no chips)", addr)
		}
	}

	// Check if we have enough players
	if len(g.getReadyActivePlayers()) >= 2 {
		g.setStatus(GameStatusWaiting)
		// Auto-start next hand after a delay if all players are still ready
		// For now, we'll wait for ready signals again
	} else {
		g.setStatus(GameStatusWaiting)
		logrus.Info("Not enough players, waiting for more")
	}
}

// dealFlop deals the flop (3 community cards)
func (g *Game) dealFlop() {
	logrus.Info("Dealing flop...")
	// TODO: Implement actual card dealing from encrypted deck
	g.currentPlayerTurn = g.getNextActivePlayerID(g.currentDealerID)
}

// dealTurn deals the turn (4th community card)
func (g *Game) dealTurn() {
	logrus.Info("Dealing turn...")
	// TODO: Implement actual card dealing from encrypted deck
	g.currentPlayerTurn = g.getNextActivePlayerID(g.currentDealerID)
}

// dealRiver deals the river (5th community card)
func (g *Game) dealRiver() {
	logrus.Info("Dealing river...")
	// TODO: Implement actual card dealing from encrypted deck
	g.currentPlayerTurn = g.getNextActivePlayerID(g.currentDealerID)
}
