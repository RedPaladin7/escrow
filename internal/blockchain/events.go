package blockchain

import (
	"context"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/sirupsen/logrus"
)

// EventListener listens for blockchain events
type EventListener struct {
	bc          *BlockchainClient
	subscribers map[string][]chan interface{}
	contractABI abi.ABI
}

// NewEventListener creates a new event listener
func NewEventListener(bc *BlockchainClient) *EventListener {
	// Parse contract ABI
	// In production, this would be loaded from the compiled contract
	abiJSON := getPokerTableABI()
	contractABI, err := abi.JSON(strings.NewReader(abiJSON))
	if err != nil {
		logrus.Errorf("Failed to parse contract ABI: %v", err)
	}

	return &EventListener{
		bc:          bc,
		subscribers: make(map[string][]chan interface{}),
		contractABI: contractABI,
	}
}

// GameCreatedEvent represents a GameCreated event
type GameCreatedEvent struct {
	GameID      [32]byte
	Creator     common.Address
	BuyIn       *big.Int
	MaxPlayers  *big.Int
	BlockNumber uint64
	TxHash      common.Hash
}

// PlayerJoinedEvent represents a PlayerJoined event
type PlayerJoinedEvent struct {
	GameID      [32]byte
	Player      common.Address
	Amount      *big.Int
	BlockNumber uint64
	TxHash      common.Hash
}

// GameStartedEvent represents a GameStarted event
type GameStartedEvent struct {
	GameID      [32]byte
	TotalPot    *big.Int
	BlockNumber uint64
	TxHash      common.Hash
}

// GameEndedEvent represents a GameEnded event
type GameEndedEvent struct {
	GameID      [32]byte
	Winners     []common.Address
	Payouts     []*big.Int
	BlockNumber uint64
	TxHash      common.Hash
}

// FundsLockedEvent represents a FundsLocked event
type FundsLockedEvent struct {
	GameID      [32]byte
	Player      common.Address
	Amount      *big.Int
	BlockNumber uint64
	TxHash      common.Hash
}

// Subscribe subscribes to a specific event type
func (el *EventListener) Subscribe(eventType string, ch chan interface{}) {
	if el.subscribers[eventType] == nil {
		el.subscribers[eventType] = make([]chan interface{}, 0)
	}
	el.subscribers[eventType] = append(el.subscribers[eventType], ch)
}

// Unsubscribe unsubscribes from an event type
func (el *EventListener) Unsubscribe(eventType string, ch chan interface{}) {
	if subs, ok := el.subscribers[eventType]; ok {
		for i, subscriber := range subs {
			if subscriber == ch {
				el.subscribers[eventType] = append(subs[:i], subs[i+1:]...)
				close(ch)
				break
			}
		}
	}
}

// publish publishes an event to all subscribers
func (el *EventListener) publish(eventType string, event interface{}) {
	if subs, ok := el.subscribers[eventType]; ok {
		for _, ch := range subs {
			select {
			case ch <- event:
			default:
				logrus.Warn("Subscriber channel full, skipping event")
			}
		}
	}
}

// ListenForEvents starts listening for blockchain events
func (el *EventListener) ListenForEvents(ctx context.Context) error {
	// Create filter query for PokerTable contract
	query := ethereum.FilterQuery{
		Addresses: []common.Address{el.bc.pokerTableAddress},
	}

	logs := make(chan types.Log)
	sub, err := el.bc.client.SubscribeFilterLogs(ctx, query, logs)
	if err != nil {
		return fmt.Errorf("failed to subscribe to logs: %w", err)
	}

	logrus.Info("Started listening for blockchain events")

	go func() {
		defer sub.Unsubscribe()

		for {
			select {
			case err := <-sub.Err():
				logrus.Errorf("Event subscription error: %v", err)
				return
			case vLog := <-logs:
				el.handleLog(vLog)
			case <-ctx.Done():
				logrus.Info("Stopped listening for blockchain events")
				return
			}
		}
	}()

	return nil
}

// handleLog processes a single log entry
func (el *EventListener) handleLog(vLog types.Log) {
	if len(vLog.Topics) == 0 {
		return
	}

	eventSig := vLog.Topics[0]

	logrus.WithFields(logrus.Fields{
		"block":     vLog.BlockNumber,
		"tx_hash":   vLog.TxHash.Hex(),
		"event_sig": eventSig.Hex(),
	}).Debug("Processing blockchain event")

	// Event signatures
	gameCreatedSig := crypto.Keccak256Hash([]byte("GameCreated(bytes32,address,uint256,uint256)"))
	playerJoinedSig := crypto.Keccak256Hash([]byte("PlayerJoined(bytes32,address,uint256)"))
	gameStartedSig := crypto.Keccak256Hash([]byte("GameStarted(bytes32,uint256)"))
	gameEndedSig := crypto.Keccak256Hash([]byte("GameEnded(bytes32,address[],uint256[])"))
	fundsLockedSig := crypto.Keccak256Hash([]byte("FundsLocked(bytes32,address,uint256)"))

	switch eventSig {
	case gameCreatedSig:
		event := el.parseGameCreatedEvent(vLog)
		if event != nil {
			el.publish("GameCreated", event)
		}

	case playerJoinedSig:
		event := el.parsePlayerJoinedEvent(vLog)
		if event != nil {
			el.publish("PlayerJoined", event)
		}

	case gameStartedSig:
		event := el.parseGameStartedEvent(vLog)
		if event != nil {
			el.publish("GameStarted", event)
		}

	case gameEndedSig:
		event := el.parseGameEndedEvent(vLog)
		if event != nil {
			el.publish("GameEnded", event)
		}

	case fundsLockedSig:
		event := el.parseFundsLockedEvent(vLog)
		if event != nil {
			el.publish("FundsLocked", event)
		}

	default:
		logrus.Debugf("Unknown event signature: %s", eventSig.Hex())
	}
}

// parseGameCreatedEvent parses a GameCreated event
func (el *EventListener) parseGameCreatedEvent(vLog types.Log) *GameCreatedEvent {
	if len(vLog.Topics) < 2 {
		return nil
	}

	var gameID [32]byte
	copy(gameID[:], vLog.Topics[1].Bytes())

	var creator common.Address
	copy(creator[:], vLog.Topics[2].Bytes()[12:])

	// Parse data (buyIn, maxPlayers)
	if len(vLog.Data) < 64 {
		return nil
	}

	buyIn := new(big.Int).SetBytes(vLog.Data[0:32])
	maxPlayers := new(big.Int).SetBytes(vLog.Data[32:64])

	return &GameCreatedEvent{
		GameID:      gameID,
		Creator:     creator,
		BuyIn:       buyIn,
		MaxPlayers:  maxPlayers,
		BlockNumber: vLog.BlockNumber,
		TxHash:      vLog.TxHash,
	}
}

// parsePlayerJoinedEvent parses a PlayerJoined event
func (el *EventListener) parsePlayerJoinedEvent(vLog types.Log) *PlayerJoinedEvent {
	if len(vLog.Topics) < 2 {
		return nil
	}

	var gameID [32]byte
	copy(gameID[:], vLog.Topics[1].Bytes())

	var player common.Address
	copy(player[:], vLog.Topics[2].Bytes()[12:])

	if len(vLog.Data) < 32 {
		return nil
	}

	amount := new(big.Int).SetBytes(vLog.Data[0:32])

	return &PlayerJoinedEvent{
		GameID:      gameID,
		Player:      player,
		Amount:      amount,
		BlockNumber: vLog.BlockNumber,
		TxHash:      vLog.TxHash,
	}
}

// parseGameStartedEvent parses a GameStarted event
func (el *EventListener) parseGameStartedEvent(vLog types.Log) *GameStartedEvent {
	if len(vLog.Topics) < 2 {
		return nil
	}

	var gameID [32]byte
	copy(gameID[:], vLog.Topics[1].Bytes())

	if len(vLog.Data) < 32 {
		return nil
	}

	totalPot := new(big.Int).SetBytes(vLog.Data[0:32])

	return &GameStartedEvent{
		GameID:      gameID,
		TotalPot:    totalPot,
		BlockNumber: vLog.BlockNumber,
		TxHash:      vLog.TxHash,
	}
}

// parseGameEndedEvent parses a GameEnded event
func (el *EventListener) parseGameEndedEvent(vLog types.Log) *GameEndedEvent {
	if len(vLog.Topics) < 2 {
		return nil
	}

	var gameID [32]byte
	copy(gameID[:], vLog.Topics[1].Bytes())

	// Parse arrays from data (complex parsing would use ABI)
	// For simplicity, we'll return basic event
	return &GameEndedEvent{
		GameID:      gameID,
		Winners:     []common.Address{},
		Payouts:     []*big.Int{},
		BlockNumber: vLog.BlockNumber,
		TxHash:      vLog.TxHash,
	}
}

// parseFundsLockedEvent parses a FundsLocked event
func (el *EventListener) parseFundsLockedEvent(vLog types.Log) *FundsLockedEvent {
	if len(vLog.Topics) < 2 {
		return nil
	}

	var gameID [32]byte
	copy(gameID[:], vLog.Topics[1].Bytes())

	var player common.Address
	copy(player[:], vLog.Topics[2].Bytes()[12:])

	if len(vLog.Data) < 32 {
		return nil
	}

	amount := new(big.Int).SetBytes(vLog.Data[0:32])

	return &FundsLockedEvent{
		GameID:      gameID,
		Player:      player,
		Amount:      amount,
		BlockNumber: vLog.BlockNumber,
		TxHash:      vLog.TxHash,
	}
}

// WatchGameCreated watches for GameCreated events
func (el *EventListener) WatchGameCreated(ctx context.Context, gameIDChan chan [32]byte) error {
	ch := make(chan interface{}, 10)
	el.Subscribe("GameCreated", ch)

	go func() {
		for {
			select {
			case event := <-ch:
				if gameCreated, ok := event.(*GameCreatedEvent); ok {
					gameIDChan <- gameCreated.GameID
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	return nil
}

// WatchPlayerJoined watches for PlayerJoined events for a specific game
func (el *EventListener) WatchPlayerJoined(ctx context.Context, gameID [32]byte, playerChan chan common.Address) error {
	ch := make(chan interface{}, 10)
	el.Subscribe("PlayerJoined", ch)

	go func() {
		for {
			select {
			case event := <-ch:
				if playerJoined, ok := event.(*PlayerJoinedEvent); ok {
					if playerJoined.GameID == gameID {
						playerChan <- playerJoined.Player
					}
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	return nil
}

// GetPastEvents retrieves past events from a block range
func (el *EventListener) GetPastEvents(fromBlock, toBlock *big.Int) ([]types.Log, error) {
	query := ethereum.FilterQuery{
		FromBlock: fromBlock,
		ToBlock:   toBlock,
		Addresses: []common.Address{el.bc.pokerTableAddress},
	}

	logs, err := el.bc.client.FilterLogs(context.Background(), query)
	if err != nil {
		return nil, fmt.Errorf("failed to filter logs: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"from_block": fromBlock.String(),
		"to_block":   toBlock.String(),
		"events":     len(logs),
	}).Info("Retrieved past events")

	return logs, nil
}

// GetGameCreatedEvents retrieves all GameCreated events
func (el *EventListener) GetGameCreatedEvents(fromBlock *big.Int) ([]GameCreatedEvent, error) {
	logs, err := el.GetPastEvents(fromBlock, nil)
	if err != nil {
		return nil, err
	}

	events := []GameCreatedEvent{}
	gameCreatedSig := crypto.Keccak256Hash([]byte("GameCreated(bytes32,address,uint256,uint256)"))

	for _, vLog := range logs {
		if len(vLog.Topics) > 0 && vLog.Topics[0] == gameCreatedSig {
			if event := el.parseGameCreatedEvent(vLog); event != nil {
				events = append(events, *event)
			}
		}
	}

	return events, nil
}

// getPokerTableABI returns a simplified ABI for the PokerTable contract
func getPokerTableABI() string {
	return `[
		{
			"anonymous": false,
			"inputs": [
				{"indexed": true, "name": "gameId", "type": "bytes32"},
				{"indexed": true, "name": "creator", "type": "address"},
				{"indexed": false, "name": "buyIn", "type": "uint256"},
				{"indexed": false, "name": "maxPlayers", "type": "uint256"}
			],
			"name": "GameCreated",
			"type": "event"
		},
		{
			"anonymous": false,
			"inputs": [
				{"indexed": true, "name": "gameId", "type": "bytes32"},
				{"indexed": true, "name": "player", "type": "address"},
				{"indexed": false, "name": "amount", "type": "uint256"}
			],
			"name": "PlayerJoined",
			"type": "event"
		},
		{
			"anonymous": false,
			"inputs": [
				{"indexed": true, "name": "gameId", "type": "bytes32"},
				{"indexed": false, "name": "totalPot", "type": "uint256"}
			],
			"name": "GameStarted",
			"type": "event"
		},
		{
			"anonymous": false,
			"inputs": [
				{"indexed": true, "name": "gameId", "type": "bytes32"},
				{"indexed": false, "name": "winners", "type": "address[]"},
				{"indexed": false, "name": "payouts", "type": "uint256[]"}
			],
			"name": "GameEnded",
			"type": "event"
		},
		{
			"anonymous": false,
			"inputs": [
				{"indexed": true, "name": "gameId", "type": "bytes32"},
				{"indexed": true, "name": "player", "type": "address"},
				{"indexed": false, "name": "amount", "type": "uint256"}
			],
			"name": "FundsLocked",
			"type": "event"
		}
	]`
}
