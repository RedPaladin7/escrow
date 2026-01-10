package protocol

import (
	"fmt"
)

// ValidateMessage validates a message structure
func ValidateMessage(msg *Message) error {
	if msg == nil {
		return fmt.Errorf("message is nil")
	}

	if msg.Type == "" {
		return fmt.Errorf("message type is empty")
	}

	if msg.From == "" {
		return fmt.Errorf("message sender is empty")
	}

	if len(msg.Payload) == 0 {
		return fmt.Errorf("message payload is empty")
	}

	return nil
}

// ValidatePlayerAction validates a player action
func ValidatePlayerAction(action string) error {
	switch action {
	case ActionFold, ActionCheck, ActionCall, ActionBet, ActionRaise, ActionAllIn:
		return nil
	default:
		return fmt.Errorf("invalid action: %s", action)
	}
}

// ValidateGameVariant validates a game variant
func ValidateGameVariant(variant string) error {
	switch variant {
	case GameVariantTexasHoldem, GameVariantOmaha, GameVariantSevenCard:
		return nil
	default:
		return fmt.Errorf("invalid game variant: %s", variant)
	}
}

// ValidateBetAmount validates a bet/raise amount
func ValidateBetAmount(amount, minBet, maxBet int) error {
	if amount < minBet {
		return fmt.Errorf("bet amount %d is less than minimum %d", amount, minBet)
	}

	if amount > maxBet {
		return fmt.Errorf("bet amount %d exceeds maximum %d", amount, maxBet)
	}

	return nil
}

// ValidatePlayerID validates a player ID
func ValidatePlayerID(playerID string) error {
	if playerID == "" {
		return fmt.Errorf("player ID is empty")
	}

	if len(playerID) > 255 {
		return fmt.Errorf("player ID too long")
	}

	return nil
}
