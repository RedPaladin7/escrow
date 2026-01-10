package game

import (
	"fmt"

	"github.com/RedPaladin7/peerpoker/internal/protocol"
	"github.com/sirupsen/logrus"
)

type PlayerAction int

const (
	PlayerActionFold PlayerAction = iota
	PlayerActionCheck
	PlayerActionCall
	PlayerActionBet
	PlayerActionRaise
)

func (pa PlayerAction) String() string {
	switch pa {
	case PlayerActionFold:
		return "fold"
	case PlayerActionCheck:
		return "check"
	case PlayerActionCall:
		return "call"
	case PlayerActionBet:
		return "bet"
	case PlayerActionRaise:
		return "raise"
	default:
		return "unknown"
	}
}

func ParsePlayerAction(action string) (PlayerAction, error) {
	switch action {
	case "fold":
		return PlayerActionFold, nil
	case "check":
		return PlayerActionCheck, nil
	case "call":
		return PlayerActionCall, nil
	case "bet":
		return PlayerActionBet, nil
	case "raise":
		return PlayerActionRaise, nil
	default:
		return 0, fmt.Errorf("invalid action: %s", action)
	}
}

// HandlePlayerAction processes a player action
func (g *Game) HandlePlayerAction(clientID, actionStr string, value int) error {
	g.lock.Lock()
	defer g.lock.Unlock()

	action, err := ParsePlayerAction(actionStr)
	if err != nil {
		return err
	}

	myState, ok := g.playerStates[clientID]
	if !ok {
		return fmt.Errorf("player %s not found", clientID)
	}

	// Check if it's this player's turn
	if myState.RotationID != g.currentPlayerTurn {
		return fmt.Errorf("it is not your turn")
	}

	// Validate action
	validActions := g.getValidActions(clientID)
	isValid := false
	for _, validAction := range validActions {
		if validAction == action {
			isValid = true
			break
		}
	}

	if !isValid {
		return fmt.Errorf("invalid action: %s", actionStr)
	}

	// Validate bet/raise amounts
	switch action {
	case PlayerActionBet:
		if value < BigBlind {
			return fmt.Errorf("bet must be at least the big blind (%d)", BigBlind)
		}
		if value > myState.Stack {
			return fmt.Errorf("bet (%d) exceeds your stack (%d)", value, myState.Stack)
		}
		g.lastRaiseAmount = value

	case PlayerActionRaise:
		minRaise := g.highestBet + g.lastRaiseAmount
		if value < minRaise {
			return fmt.Errorf("raise must be at least %d", minRaise)
		}
		if value > myState.Stack {
			return fmt.Errorf("raise (%d) exceeds your stack (%d)", value, myState.Stack)
		}
		g.lastRaiseAmount = value - g.highestBet

	case PlayerActionCall:
		amountNeeded := g.highestBet - myState.CurrentRoundBet
		if amountNeeded > myState.Stack {
			logrus.Infof("Call will be all-in for %d", myState.Stack)
		}
	}

	// Handle fold - reveal keys to other players
	if action == PlayerActionFold {
		g.sendToPlayers(protocol.TypeRevealKeys, protocol.RevealKeysPayload{
			EncryptionKey: g.deckKeys.EncKey.String(),
			DecryptionKey: g.deckKeys.DecKey.String(),
			Prime:         g.deckKeys.Prime.String(),
		}, g.getOtherPlayers()...)
		myState.IsFolded = true
	}

	// Update state
	g.updatePlayerState(clientID, action, value)

	// Broadcast action to other players
	g.sendToPlayers(protocol.TypePlayerAction, protocol.PlayerActionPayload{
		Action:            actionStr,
		Value:             value,
		CurrentGameStatus: g.currentStatus.String(),
	}, g.getOtherPlayers()...)

	// Advance turn
	g.advanceTurnAndCheckRoundEnd()

	return nil
}

// Get valid actions for a player
func (g *Game) getValidActions(clientID string) []PlayerAction {
	state, ok := g.playerStates[clientID]
	if !ok {
		return []PlayerAction{}
	}

	actions := []PlayerAction{PlayerActionFold}

	// Check
	if g.highestBet == 0 || state.CurrentRoundBet == g.highestBet {
		actions = append(actions, PlayerActionCheck)
	}

	// Call
	if g.highestBet > state.CurrentRoundBet && state.Stack > 0 {
		actions = append(actions, PlayerActionCall)
	}

	// Bet or Raise
	minRaise := g.highestBet + g.lastRaiseAmount
	if g.highestBet == 0 {
		minRaise = BigBlind
	}

	if state.Stack > (minRaise - state.CurrentRoundBet) {
		if g.highestBet == 0 {
			actions = append(actions, PlayerActionBet)
		} else {
			actions = append(actions, PlayerActionRaise)
		}
	}

	return actions
}
