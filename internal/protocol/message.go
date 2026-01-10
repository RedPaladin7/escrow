package protocol

import (
	"encoding/json"
	"time"
)

type MessageType string

const (
	TypeHandshake       MessageType = "handshake"
	TypePeerList        MessageType = "peer_list"
	TypePlayerAction    MessageType = "player_action"
	TypePlayerReady     MessageType = "player_ready"
	TypeEncDeck         MessageType = "enc_deck"
	TypeGameState       MessageType = "game_state"
	TypeShuffleStatus   MessageType = "shuffle_status"
	TypeGetRPC          MessageType = "get_rpc"
	TypeRPCResponse     MessageType = "rpc_response"
	TypeRevealKeys      MessageType = "reveal_keys"
	TypeShowdownResult  MessageType = "showdown_result"
	TypeError           MessageType = "error"
	TypePing            MessageType = "ping"
	TypePong            MessageType = "pong"
)

// Message is the base message structure for all communications
type Message struct {
	Type      MessageType     `json:"type"`
	From      string          `json:"from"`
	Payload   json.RawMessage `json:"payload"`
	Timestamp time.Time       `json:"timestamp"`
}

// NewMessage creates a new message with the given type and payload
func NewMessage(from string, msgType MessageType, payload interface{}) (*Message, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	return &Message{
		Type:      msgType,
		From:      from,
		Payload:   data,
		Timestamp: time.Now(),
	}, nil
}

// MarshalJSON custom marshaller to format timestamp
func (m *Message) MarshalJSON() ([]byte, error) {
	type Alias Message
	return json.Marshal(&struct {
		*Alias
		Timestamp string `json:"timestamp"`
	}{
		Alias:     (*Alias)(m),
		Timestamp: m.Timestamp.Format(time.RFC3339),
	})
}

// UnmarshalJSON custom unmarshaller to parse timestamp
func (m *Message) UnmarshalJSON(data []byte) error {
	type Alias Message
	aux := &struct {
		*Alias
		Timestamp string `json:"timestamp"`
	}{
		Alias: (*Alias)(m),
	}
	
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	
	if aux.Timestamp != "" {
		timestamp, err := time.Parse(time.RFC3339, aux.Timestamp)
		if err == nil {
			m.Timestamp = timestamp
		}
	}
	
	return nil
}

// HandshakePayload represents the handshake message
type HandshakePayload struct {
	Version     string `json:"version"`
	GameVariant string `json:"game_variant"`
	ListenAddr  string `json:"listen_addr"`
}

// PeerListPayload contains a list of connected peers
type PeerListPayload struct {
	Peers []string `json:"peers"`
}

// PlayerActionPayload represents a player's action
type PlayerActionPayload struct {
	Action            string `json:"action"`
	Value             int    `json:"value,omitempty"`
	CurrentGameStatus string `json:"current_game_status"`
}

// PlayerReadyPayload indicates a player is ready
type PlayerReadyPayload struct {
	PlayerID string `json:"player_id"`
}

// EncDeckPayload contains an encrypted deck
type EncDeckPayload struct {
	Deck [][]byte `json:"deck"`
}

// GameStatePayload represents the current game state
type GameStatePayload struct {
	Status         string   `json:"status"`
	CurrentPot     int      `json:"current_pot"`
	HighestBet     int      `json:"highest_bet"`
	CurrentTurn    int      `json:"current_turn"`
	DealerID       int      `json:"dealer_id"`
	CommunityCards []int    `json:"community_cards"`
	Players        []string `json:"players"`
}

// ShuffleStatusPayload contains a shuffled deck
type ShuffleStatusPayload struct {
	Deck [][]byte `json:"deck"`
}

// GetRPCPayload requests card decryption from other players
type GetRPCPayload struct {
	CardIndices   []int    `json:"card_indices"`
	EncryptedData [][]byte `json:"encrypted_data"`
	OriginalOwner string   `json:"original_owner"`
}

// RPCResponsePayload contains decrypted card data
type RPCResponsePayload struct {
	CardIndices   []int    `json:"card_indices"`
	DecryptedData [][]byte `json:"decrypted_data"`
}

// RevealKeysPayload contains encryption keys for verification
type RevealKeysPayload struct {
	EncryptionKey string `json:"encryption_key"`
	DecryptionKey string `json:"decryption_key"`
	Prime         string `json:"prime"`
}

// ShowdownResultPayload contains showdown results
type ShowdownResultPayload struct {
	PlayerAddr string   `json:"player_addr"`
	HandRank   int32    `json:"hand_rank"`
	HandName   string   `json:"hand_name"`
	Cards      []string `json:"cards,omitempty"`
}

// ErrorPayload represents an error message
type ErrorPayload struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// PingPayload for connection health check
type PingPayload struct {
	Timestamp int64 `json:"timestamp"`
}

// PongPayload response to ping
type PongPayload struct {
	Timestamp     int64 `json:"timestamp"`
	PingTimestamp int64 `json:"ping_timestamp"`
}
