package blockchain

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

// GenerateGameID generates a unique game ID from parameters
func GenerateGameID(creator common.Address, timestamp int64, buyIn *big.Int) [32]byte {
	data := append(creator.Bytes(), big.NewInt(timestamp).Bytes()...)
	data = append(data, buyIn.Bytes()...)
	return crypto.Keccak256Hash(data)
}

// BytesToGameID converts a byte slice to a game ID
func BytesToGameID(b []byte) ([32]byte, error) {
	var gameID [32]byte
	if len(b) != 32 {
		return gameID, fmt.Errorf("invalid game ID length: expected 32, got %d", len(b))
	}
	copy(gameID[:], b)
	return gameID, nil
}

// HexToGameID converts a hex string to a game ID
func HexToGameID(hexStr string) ([32]byte, error) {
	hexStr = strings.TrimPrefix(hexStr, "0x")
	b, err := hex.DecodeString(hexStr)
	if err != nil {
		return [32]byte{}, fmt.Errorf("invalid hex string: %w", err)
	}
	return BytesToGameID(b)
}

// GameIDToHex converts a game ID to a hex string
func GameIDToHex(gameID [32]byte) string {
	return "0x" + hex.EncodeToString(gameID[:])
}

// IsValidAddress checks if a string is a valid Ethereum address
func IsValidAddress(address string) bool {
	return common.IsHexAddress(address)
}

// FormatAddress formats an address with checksum
func FormatAddress(address common.Address) string {
	return address.Hex()
}

// ParseAddress parses a string to an Ethereum address
func ParseAddress(addressStr string) (common.Address, error) {
	if !IsValidAddress(addressStr) {
		return common.Address{}, fmt.Errorf("invalid address: %s", addressStr)
	}
	return common.HexToAddress(addressStr), nil
}

// ConvertToWei converts a decimal amount to wei
func ConvertToWei(amount float64) *big.Int {
	return EthToWei(big.NewFloat(amount))
}

// ConvertFromWei converts wei to a decimal amount
func ConvertFromWei(wei *big.Int) float64 {
	eth := WeiToEth(wei)
	result, _ := eth.Float64()
	return result
}

// FormatWei formats wei as a human-readable string
func FormatWei(wei *big.Int) string {
	eth := WeiToEth(wei)
	return fmt.Sprintf("%.6f ETH", eth)
}

// CalculatePlatformFee calculates the platform fee from a pot
func CalculatePlatformFee(pot *big.Int, feePercent int) *big.Int {
	fee := new(big.Int).Mul(pot, big.NewInt(int64(feePercent)))
	return new(big.Int).Div(fee, big.NewInt(100))
}

// CalculateNetPot calculates the net pot after platform fee
func CalculateNetPot(pot *big.Int, feePercent int) *big.Int {
	fee := CalculatePlatformFee(pot, feePercent)
	return new(big.Int).Sub(pot, fee)
}

// SplitPot splits a pot among multiple winners
func SplitPot(pot *big.Int, numWinners int) []*big.Int {
	if numWinners <= 0 {
		return []*big.Int{}
	}

	share := new(big.Int).Div(pot, big.NewInt(int64(numWinners)))
	remainder := new(big.Int).Mod(pot, big.NewInt(int64(numWinners)))

	shares := make([]*big.Int, numWinners)
	for i := 0; i < numWinners; i++ {
		shares[i] = new(big.Int).Set(share)
		if i == 0 {
			// Give remainder to first winner
			shares[i].Add(shares[i], remainder)
		}
	}

	return shares
}

// ValidateBuyIn validates a buy-in amount
func ValidateBuyIn(buyIn *big.Int, minBuyIn, maxBuyIn *big.Int) error {
	if buyIn.Cmp(minBuyIn) < 0 {
		return fmt.Errorf("buy-in %s is less than minimum %s", FormatWei(buyIn), FormatWei(minBuyIn))
	}
	if buyIn.Cmp(maxBuyIn) > 0 {
		return fmt.Errorf("buy-in %s exceeds maximum %s", FormatWei(buyIn), FormatWei(maxBuyIn))
	}
	return nil
}

// CalculateGasCost calculates the total gas cost
func CalculateGasCost(gasUsed uint64, gasPrice *big.Int) *big.Int {
	return new(big.Int).Mul(big.NewInt(int64(gasUsed)), gasPrice)
}

// FormatGasCost formats gas cost as a human-readable string
func FormatGasCost(gasUsed uint64, gasPrice *big.Int) string {
	cost := CalculateGasCost(gasUsed, gasPrice)
	return FormatWei(cost)
}

// HashMessage hashes a message using Keccak256
func HashMessage(message []byte) common.Hash {
	return crypto.Keccak256Hash(message)
}

// AddressToString converts an address slice to string slice
func AddressToString(addresses []common.Address) []string {
	result := make([]string, len(addresses))
	for i, addr := range addresses {
		result[i] = addr.Hex()
	}
	return result
}

// StringToAddress converts a string slice to address slice
func StringToAddress(addressStrings []string) ([]common.Address, error) {
	result := make([]common.Address, len(addressStrings))
	for i, addrStr := range addressStrings {
		if !IsValidAddress(addrStr) {
			return nil, fmt.Errorf("invalid address at index %d: %s", i, addrStr)
		}
		result[i] = common.HexToAddress(addrStr)
	}
	return result, nil
}
