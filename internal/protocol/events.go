package protocol

import (
	"encoding/json"
	"time"
)

// EventType represents different types of real-time events
type EventType string

const (
	EventGameStateUpdate EventType = "game_state_update"
	EventPlayerJoined    EventType = "player_joined"
	EventPlayerLeft      EventType = "player_left"
	EventPlayerAction    EventType = "player_action"
	EventNewHand         EventType = "new_hand"
	EventCommunityCard   EventType = "community_card"
	EventShowdown        EventType = "showdown"
	EventWinner          EventType = "winner"
	EventError           EventType = "error"
	EventTurnChange      EventType = "turn_change"
	EventBlindsPosted    EventType = "blinds_posted"
)

// Event represents a real-time event sent to clients
type Event struct {
	Type      EventType       `json:"type"`
	Data      json.RawMessage `json:"data"`
	Timestamp time.Time       `json:"timestamp"`
}

// NewEvent creates a new event with the given type and data
func NewEvent(eventType EventType, data interface{}) (*Event, error) {
	payload, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	return &Event{
		Type:      eventType,
		Data:      payload,
		Timestamp: time.Now(),
	}, nil
}

// MarshalJSON custom marshaller for Event
func (e *Event) MarshalJSON() ([]byte, error) {
	type Alias Event
	return json.Marshal(&struct {
		*Alias
		Timestamp string `json:"timestamp"`
	}{
		Alias:     (*Alias)(e),
		Timestamp: e.Timestamp.Format(time.RFC3339),
	})
}

// GameStateUpdateEvent contains full game state
type GameStateUpdateEvent struct {
	Status         string       `json:"status"`
	Pot            int          `json:"pot"`
	HighestBet     int          `json:"highest_bet"`
	CurrentTurn    string       `json:"current_turn"`
	CommunityCards []CardData   `json:"community_cards"`
	Players        []PlayerData `json:"players"`
}

// PlayerJoinedEvent notifies when a player joins
type PlayerJoinedEvent struct {
	PlayerID   string `json:"player_id"`
	PlayerName string `json:"player_name,omitempty"`
	Stack      int    `json:"stack"`
}

// PlayerLeftEvent notifies when a player leaves
type PlayerLeftEvent struct {
	PlayerID string `json:"player_id"`
	Reason   string `json:"reason,omitempty"`
}

// PlayerActionEvent notifies of a player action
type PlayerActionEvent struct {
	PlayerID string `json:"player_id"`
	Action   string `json:"action"`
	Amount   int    `json:"amount,omitempty"`
	NewPot   int    `json:"new_pot"`
	NewStack int    `json:"new_stack"`
}

// NewHandEvent notifies when a new hand starts
type NewHandEvent struct {
	DealerID    int      `json:"dealer_id"`
	SmallBlind  int      `json:"small_blind"`
	BigBlind    int      `json:"big_blind"`
	PlayerCount int      `json:"player_count"`
	Players     []string `json:"players"`
}

// CommunityCardEvent notifies when community cards are dealt
type CommunityCardEvent struct {
	Stage string     `json:"stage"` // "flop", "turn", "river"
	Cards []CardData `json:"cards"`
}

// ShowdownEvent notifies when showdown occurs
type ShowdownEvent struct {
	Results []ShowdownPlayerResult `json:"results"`
}

// WinnerEvent notifies of hand winner(s)
type WinnerEvent struct {
	Winners []WinnerData `json:"winners"`
	Pot     int          `json:"pot"`
}

// TurnChangeEvent notifies when the turn changes
type TurnChangeEvent struct {
	PlayerID      string   `json:"player_id"`
	RotationID    int      `json:"rotation_id"`
	ValidActions  []string `json:"valid_actions"`
	TimeRemaining int      `json:"time_remaining,omitempty"`
}

// BlindsPostedEvent notifies when blinds are posted
type BlindsPostedEvent struct {
	SmallBlindPlayer string `json:"small_blind_player"`
	BigBlindPlayer   string `json:"big_blind_player"`
	SmallBlindAmount int    `json:"small_blind_amount"`
	BigBlindAmount   int    `json:"big_blind_amount"`
}

// CardData represents a card in events
type CardData struct {
	Suit    string `json:"suit"`
	Value   int    `json:"value"`
	Display string `json:"display"`
}

// PlayerData represents player state in events
type PlayerData struct {
	PlayerID      string `json:"player_id"`
	Stack         int    `json:"stack"`
	CurrentBet    int    `json:"current_bet"`
	IsActive      bool   `json:"is_active"`
	IsFolded      bool   `json:"is_folded"`
	IsAllIn       bool   `json:"is_all_in"`
	IsDealer      bool   `json:"is_dealer"`
	IsCurrentTurn bool   `json:"is_current_turn"`
}

// ShowdownPlayerResult represents a player's result at showdown
type ShowdownPlayerResult struct {
	PlayerID string     `json:"player_id"`
	Hand     []CardData `json:"hand"`
	HandRank string     `json:"hand_rank"`
	Rank     int32      `json:"rank"`
}

// WinnerData represents a winner's information
type WinnerData struct {
	PlayerID  string `json:"player_id"`
	Amount    int    `json:"amount"`
	HandName  string `json:"hand_name,omitempty"`
	NewStack  int    `json:"new_stack"`
}

// ErrorEvent represents an error event
type ErrorEvent struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}
