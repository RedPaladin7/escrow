package blockchain

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/sirupsen/logrus"
)

// EventListener listens for blockchain events
type EventListener struct {
	bc          *BlockchainClient
	subscribers map[string][]chan interface{}
}

// NewEventListener creates a new event listener
func NewEventListener(bc *BlockchainClient) *EventListener {
	return &EventListener{
		bc:          bc,
		subscribers: make(map[string][]chan interface{}),
	}
}

// GameCreatedEvent represents a GameCreated event
type GameCreatedEvent struct {
	GameID     [32]byte
	Creator    common.Address
	BuyIn      *big.Int
	MaxPlayers *big.Int
	BlockNumber uint64
	TxHash     common.Hash
}

// PlayerJoinedEvent represents a PlayerJoined event
type PlayerJoinedEvent struct {
	GameID  [32]byte
	Player  common.Address
	Amount  *big.Int
	BlockNumber uint64
	TxHash  common.Hash
}

// GameStartedEvent represents a GameStarted event
type GameStartedEvent struct {
	GameID   [32]byte
	TotalPot *big.Int
	BlockNumber uint64
	TxHash   common.Hash
}

// GameEndedEvent represents a GameEnded event
type GameEndedEvent struct {
	GameID   [32]byte
	Winners  []common.Address
	Payouts  []*big.Int
	BlockNumber uint64
	TxHash   common.Hash
}

// FundsLockedEvent represents a FundsLocked event
type FundsLockedEvent struct {
	GameID  [32]byte
	Player  common.Address
	Amount  *big.Int
	BlockNumber uint64
	TxHash  common.Hash
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
	// Parse event based on topic
	// Topic[0] is the event signature hash
	
	// TODO: Implement actual event parsing using contract ABI
	// This would typically use the generated contract bindings
	
	logrus.WithFields(logrus.Fields{
		"block":   vLog.BlockNumber,
		"tx_hash": vLog.TxHash.Hex(),
		"topics":  len(vLog.Topics),
	}).Debug("Received blockchain event")

	// Example event parsing (would be replaced with actual ABI parsing):
	// if len(vLog.Topics) > 0 {
	//     eventSig := vLog.Topics[0].Hex()
	//     
	//     switch eventSig {
	//     case crypto.Keccak256Hash([]byte("GameCreated(bytes32,address,uint256,uint256)")).Hex():
	//         event := parseGameCreatedEvent(vLog)
	//         el.publish("GameCreated", event)
	//     case crypto.Keccak256Hash([]byte("PlayerJoined(bytes32,address,uint256)")).Hex():
	//         event := parsePlayerJoinedEvent(vLog)
	//         el.publish("PlayerJoined", event)
	//     // ... handle other events
	//     }
	// }
}

// WatchGameCreated watches for GameCreated events
func (el *EventListener) WatchGameCreated(ctx context.Context, gameIDChan chan [32]byte) error {
	// TODO: Implement using contract bindings
	// This would use the generated contract's WatchGameCreated method
	
	logrus.Info("Watching for GameCreated events")
	return nil
}

// WatchPlayerJoined watches for PlayerJoined events for a specific game
func (el *EventListener) WatchPlayerJoined(ctx context.Context, gameID [32]byte, playerChan chan common.Address) error {
	// TODO: Implement using contract bindings
	
	logrus.WithField("game_id", fmt.Sprintf("0x%x", gameID)).Info("Watching for PlayerJoined events")
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
	// TODO: Implement using contract bindings
	// This would parse logs and return typed events
	
	return []GameCreatedEvent{}, nil
}
