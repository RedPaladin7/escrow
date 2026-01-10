package blockchain

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/sirupsen/logrus"
)

// VerifyBuyIns verifies that all players have locked their buy-ins on-chain
func (bc *BlockchainClient) VerifyBuyIns(gameID [32]byte, players []string) (bool, error) {
	logrus.WithFields(logrus.Fields{
		"game_id":      fmt.Sprintf("0x%x", gameID),
		"player_count": len(players),
	}).Info("Verifying buy-ins for all players")

	for _, playerAddr := range players {
		addr := common.HexToAddress(playerAddr)
		
		verified, err := bc.VerifyBuyIn(gameID, addr)
		if err != nil {
			return false, fmt.Errorf("failed to verify buy-in for %s: %w", playerAddr, err)
		}

		if !verified {
			logrus.Warnf("Player %s has not locked buy-in", playerAddr)
			return false, nil
		}
	}

	logrus.Info("All players have verified buy-ins")
	return true, nil
}

// VerifyPlayerBalance verifies a player has sufficient balance
func (bc *BlockchainClient) VerifyPlayerBalance(playerAddr common.Address, requiredAmount *big.Int) (bool, error) {
	balance, err := bc.GetBalance(playerAddr)
	if err != nil {
		return false, fmt.Errorf("failed to get player balance: %w", err)
	}

	hasBalance := balance.Cmp(requiredAmount) >= 0

	logrus.WithFields(logrus.Fields{
		"player":   playerAddr.Hex(),
		"balance":  balance.String(),
		"required": requiredAmount.String(),
		"verified": hasBalance,
	}).Debug("Verified player balance")

	return hasBalance, nil
}

// VerifyGameState verifies the on-chain game state matches expected state
func (bc *BlockchainClient) VerifyGameState(gameID [32]byte, expectedPlayers int, expectedPot *big.Int) (bool, error) {
	gameInfo, err := bc.GetGameInfo(gameID)
	if err != nil {
		return false, fmt.Errorf("failed to get game info: %w", err)
	}

	playersMatch := gameInfo.PlayerCount.Cmp(big.NewInt(int64(expectedPlayers))) == 0
	potMatch := gameInfo.TotalPot.Cmp(expectedPot) == 0

	verified := playersMatch && potMatch

	logrus.WithFields(logrus.Fields{
		"game_id":         fmt.Sprintf("0x%x", gameID),
		"expected_players": expectedPlayers,
		"actual_players":   gameInfo.PlayerCount.String(),
		"expected_pot":     expectedPot.String(),
		"actual_pot":       gameInfo.TotalPot.String(),
		"verified":         verified,
	}).Info("Verified game state")

	return verified, nil
}

// VerifyTransaction verifies a transaction was successfully mined
func (bc *BlockchainClient) VerifyTransaction(txHash common.Hash) (bool, error) {
	receipt, err := bc.client.TransactionReceipt(context.Background(), txHash)
	if err != nil {
		return false, fmt.Errorf("transaction not found: %w", err)
	}

	success := receipt.Status == 1

	logrus.WithFields(logrus.Fields{
		"tx_hash":      txHash.Hex(),
		"block_number": receipt.BlockNumber.String(),
		"gas_used":     receipt.GasUsed,
		"success":      success,
	}).Info("Verified transaction")

	return success, nil
}

// VerifyContractDeployment verifies a contract is deployed at an address
func (bc *BlockchainClient) VerifyContractDeployment(contractAddr common.Address) (bool, error) {
	code, err := bc.client.CodeAt(context.Background(), contractAddr, nil)
	if err != nil {
		return false, fmt.Errorf("failed to get contract code: %w", err)
	}

	isDeployed := len(code) > 0

	logrus.WithFields(logrus.Fields{
		"address":     contractAddr.Hex(),
		"code_size":   len(code),
		"is_deployed": isDeployed,
	}).Debug("Verified contract deployment")

	return isDeployed, nil
}

// VerifyPlayerInGame verifies a player is registered in a game
func (bc *BlockchainClient) VerifyPlayerInGame(gameID [32]byte, playerAddr common.Address) (bool, error) {
	return bc.VerifyBuyIn(gameID, playerAddr)
}

// VerifyWinnings verifies that winnings were distributed correctly
func (bc *BlockchainClient) VerifyWinnings(gameID [32]byte, winner common.Address, expectedAmount *big.Int) (bool, error) {
	// This would check the GameEnded event and verify the payout
	// TODO: Implement using event logs

	logrus.WithFields(logrus.Fields{
		"game_id": fmt.Sprintf("0x%x", gameID),
		"winner":  winner.Hex(),
		"amount":  expectedAmount.String(),
	}).Info("Verifying winnings distribution")

	return true, nil // Placeholder
}

// CheckGameExists verifies a game exists on-chain
func (bc *BlockchainClient) CheckGameExists(gameID [32]byte) (bool, error) {
	gameInfo, err := bc.GetGameInfo(gameID)
	if err != nil {
		return false, err
	}

	exists := gameInfo.Creator != common.HexToAddress("0x0")

	logrus.WithFields(logrus.Fields{
		"game_id": fmt.Sprintf("0x%x", gameID),
		"exists":  exists,
	}).Debug("Checked game existence")

	return exists, nil
}
