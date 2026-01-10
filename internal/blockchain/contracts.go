package blockchain

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/sirupsen/logrus"
)

// Note: After deploying contracts, run these commands to generate bindings:
// abigen --abi=artifacts/contracts/PokerTable.sol/PokerTable.json --pkg=blockchain --type=PokerTableContract --out=internal/blockchain/poker_table_binding.go
// abigen --abi=artifacts/contracts/PotManager.sol/PotManager.json --pkg=blockchain --type=PotManagerContract --out=internal/blockchain/pot_manager_binding.go
// abigen --abi=artifacts/contracts/PlayerRegistry.sol/PlayerRegistry.json --pkg=blockchain --type=PlayerRegistryContract --out=internal/blockchain/player_registry_binding.go

// For now, we'll use interface-based approach until bindings are generated

type GameInfo struct {
	Creator     common.Address
	BuyIn       *big.Int
	SmallBlind  *big.Int
	BigBlind    *big.Int
	MaxPlayers  *big.Int
	TotalPot    *big.Int
	PlayerCount *big.Int
	Status      uint8
}

// CreateGame creates a new poker game on-chain
func (bc *BlockchainClient) CreateGame(buyIn, smallBlind, bigBlind *big.Int, maxPlayers uint8) ([32]byte, error) {
	var gameID [32]byte

	logrus.WithFields(logrus.Fields{
		"buy_in":      buyIn.String(),
		"small_blind": smallBlind.String(),
		"big_blind":   bigBlind.String(),
		"max_players": maxPlayers,
	}).Info("Creating game on blockchain")

	auth, err := bc.GetTransactor()
	if err != nil {
		return gameID, fmt.Errorf("failed to get transactor: %w", err)
	}

	// Call contract (will work once bindings are generated)
	// tx, err := bc.pokerTable.CreateGame(auth, buyIn, smallBlind, bigBlind, big.NewInt(int64(maxPlayers)))
	// if err != nil {
	//     return gameID, fmt.Errorf("failed to create game: %w", err)
	// }
	//
	// receipt, err := bind.WaitMined(context.Background(), bc.client, tx)
	// if err != nil {
	//     return gameID, fmt.Errorf("transaction failed: %w", err)
	// }
	//
	// // Parse GameCreated event
	// for _, log := range receipt.Logs {
	//     event, err := bc.pokerTable.ParseGameCreated(*log)
	//     if err == nil {
	//         logrus.WithField("game_id", fmt.Sprintf("0x%x", event.GameId)).Info("Game created successfully")
	//         return event.GameId, nil
	//     }
	// }

	logrus.Info("CreateGame called (bindings not generated yet)")
	// Return mock game ID for testing without blockchain
	gameID = GenerateGameID(bc.publicAddress, int64(1), buyIn)
	return gameID, nil
}

// JoinGame joins an existing game with buy-in
func (bc *BlockchainClient) JoinGame(gameID [32]byte, buyInAmount *big.Int) error {
	logrus.WithFields(logrus.Fields{
		"game_id": fmt.Sprintf("0x%x", gameID),
		"buy_in":  buyInAmount.String(),
	}).Info("Joining game on blockchain")

	auth, err := bc.GetTransactor()
	if err != nil {
		return fmt.Errorf("failed to get transactor: %w", err)
	}

	auth.Value = buyInAmount

	// Call contract (will work once bindings are generated)
	// tx, err := bc.pokerTable.JoinGame(auth, gameID)
	// if err != nil {
	//     return fmt.Errorf("failed to join game: %w", err)
	// }
	//
	// _, err = bind.WaitMined(context.Background(), bc.client, tx)
	// if err != nil {
	//     return fmt.Errorf("transaction failed: %w", err)
	// }
	//
	// logrus.Info("Joined game successfully")

	logrus.Info("JoinGame called (bindings not generated yet)")
	return nil
}

// VerifyBuyIn verifies that a player has locked funds for the game
func (bc *BlockchainClient) VerifyBuyIn(gameID [32]byte, playerAddr common.Address) (bool, error) {
	logrus.WithFields(logrus.Fields{
		"game_id": fmt.Sprintf("0x%x", gameID),
		"player":  playerAddr.Hex(),
	}).Debug("Verifying buy-in")

	callOpts := bc.GetCallOpts()

	// Call contract (will work once bindings are generated)
	// isInGame, err := bc.pokerTable.IsPlayerInGame(callOpts, gameID, playerAddr)
	// if err != nil {
	//     return false, fmt.Errorf("failed to verify buy-in: %w", err)
	// }
	// return isInGame, nil

	_ = callOpts // Suppress unused variable warning
	logrus.Debug("VerifyBuyIn called (bindings not generated yet)")
	return true, nil // Placeholder - allows game to proceed without blockchain
}

// StartGame starts the game on-chain
func (bc *BlockchainClient) StartGame(gameID [32]byte) error {
	logrus.WithFields(logrus.Fields{
		"game_id": fmt.Sprintf("0x%x", gameID),
	}).Info("Starting game on blockchain")

	auth, err := bc.GetTransactor()
	if err != nil {
		return fmt.Errorf("failed to get transactor: %w", err)
	}

	// Call contract (will work once bindings are generated)
	// tx, err := bc.pokerTable.StartGame(auth, gameID)
	// if err != nil {
	//     return fmt.Errorf("failed to start game: %w", err)
	// }
	//
	// _, err = bind.WaitMined(context.Background(), bc.client, tx)
	// if err != nil {
	//     return fmt.Errorf("transaction failed: %w", err)
	// }
	//
	// logrus.Info("Game started successfully")

	_ = auth // Suppress unused variable warning
	logrus.Info("StartGame called (bindings not generated yet)")
	return nil
}

// EndGame ends the game and distributes winnings
func (bc *BlockchainClient) EndGame(gameID [32]byte, winners []common.Address, amounts []*big.Int) error {
	logrus.WithFields(logrus.Fields{
		"game_id":      fmt.Sprintf("0x%x", gameID),
		"winners":      len(winners),
		"total_payout": sumAmounts(amounts).String(),
	}).Info("Ending game on blockchain")

	if len(winners) != len(amounts) {
		return fmt.Errorf("winners and amounts length mismatch")
	}

	auth, err := bc.GetTransactor()
	if err != nil {
		return fmt.Errorf("failed to get transactor: %w", err)
	}

	// Call contract (will work once bindings are generated)
	// tx, err := bc.pokerTable.EndGame(auth, gameID, winners, amounts)
	// if err != nil {
	//     return fmt.Errorf("failed to end game: %w", err)
	// }
	//
	// receipt, err := bind.WaitMined(context.Background(), bc.client, tx)
	// if err != nil {
	//     return fmt.Errorf("transaction failed: %w", err)
	// }
	//
	// logrus.WithField("tx_hash", receipt.TxHash.Hex()).Info("Game ended successfully")

	_ = auth // Suppress unused variable warning
	logrus.Info("EndGame called (bindings not generated yet)")
	return nil
}

// NEW: EndGameWithPenalty ends game with penalty applied to abandoned player
func (bc *BlockchainClient) EndGameWithPenalty(
	gameID string,
	abandonedPlayer common.Address,
	winners []common.Address,
	amounts []*big.Int,
) error {
	logrus.WithFields(logrus.Fields{
		"game_id":          gameID,
		"abandoned_player": abandonedPlayer.Hex(),
		"winners":          len(winners),
		"total_payout":     sumAmounts(amounts).String(),
	}).Info("💀 Ending game with penalty on blockchain")

	// Validate inputs
	if len(winners) != len(amounts) {
		return fmt.Errorf("winners and amounts length mismatch")
	}

	if len(winners) == 0 {
		return fmt.Errorf("no winners specified")
	}

	// Convert gameID string to [32]byte
	var gameIDBytes [32]byte
	gameIDHex := common.HexToHash(gameID)
	copy(gameIDBytes[:], gameIDHex[:])

	auth, err := bc.GetTransactor()
	if err != nil {
		return fmt.Errorf("failed to get transactor: %w", err)
	}

	// Log penalty details
	logrus.Info("📝 Penalty transaction details:")
	logrus.Infof("  - Game ID: %s", gameID)
	logrus.Infof("  - Abandoned Player: %s", abandonedPlayer.Hex())
	logrus.Infof("  - Number of Winners: %d", len(winners))

	totalPayout := big.NewInt(0)
	for i, winner := range winners {
		logrus.Infof("    Winner %d: %s -> %s wei", i+1, winner.Hex(), amounts[i].String())
		totalPayout.Add(totalPayout, amounts[i])
	}
	logrus.Infof("  - Total Payout: %s wei", totalPayout.String())

	// Call contract (will work once bindings are generated)
	// tx, err := bc.pokerTable.EndGameWithPenalty(auth, gameIDBytes, abandonedPlayer, winners, amounts)
	// if err != nil {
	//     return fmt.Errorf("failed to call endGameWithPenalty: %w", err)
	// }
	//
	// logrus.WithField("tx_hash", tx.Hash().Hex()).Info("Transaction submitted, waiting for confirmation...")
	//
	// ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	// defer cancel()
	//
	// receipt, err := bind.WaitMined(ctx, bc.client, tx)
	// if err != nil {
	//     return fmt.Errorf("transaction failed: %w", err)
	// }
	//
	// if receipt.Status != 1 {
	//     return fmt.Errorf("transaction reverted")
	// }
	//
	// logrus.WithFields(logrus.Fields{
	//     "tx_hash":  receipt.TxHash.Hex(),
	//     "gas_used": receipt.GasUsed,
	//     "block":    receipt.BlockNumber.Uint64(),
	// }).Info("✅ Penalty transaction confirmed")

	// Simulate blockchain delay for testing
	_ = auth // Suppress unused variable warning
	time.Sleep(1 * time.Second)
	logrus.Info("✅ EndGameWithPenalty called (bindings not generated yet)")

	return nil
}

// GetGameInfo retrieves game information from the blockchain
func (bc *BlockchainClient) GetGameInfo(gameID [32]byte) (*GameInfo, error) {
	callOpts := bc.GetCallOpts()

	// Call contract (will work once bindings are generated)
	// result, err := bc.pokerTable.GetGame(callOpts, gameID)
	// if err != nil {
	//     return nil, fmt.Errorf("failed to get game info: %w", err)
	// }
	//
	// return &GameInfo{
	//     Creator:     result.Creator,
	//     BuyIn:       result.BuyIn,
	//     SmallBlind:  result.SmallBlind,
	//     BigBlind:    result.BigBlind,
	//     MaxPlayers:  result.MaxPlayers,
	//     TotalPot:    result.TotalPot,
	//     PlayerCount: result.PlayerCount,
	//     Status:      result.Status,
	// }, nil

	_ = callOpts // Suppress unused variable warning
	logrus.Debug("GetGameInfo called (bindings not generated yet)")

	// Return mock data for testing
	return &GameInfo{
		Creator:     bc.publicAddress,
		BuyIn:       big.NewInt(1000000000000000000), // 1 ETH
		SmallBlind:  big.NewInt(10),
		BigBlind:    big.NewInt(20),
		MaxPlayers:  big.NewInt(6),
		TotalPot:    big.NewInt(0),
		PlayerCount: big.NewInt(0),
		Status:      0, // Waiting
	}, nil
}

// Helper function to sum amounts
func sumAmounts(amounts []*big.Int) *big.Int {
	total := big.NewInt(0)
	for _, amount := range amounts {
		total.Add(total, amount)
	}
	return total
}
