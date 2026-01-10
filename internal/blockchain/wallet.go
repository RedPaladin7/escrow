package blockchain

import (
	"crypto/ecdsa"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

type Wallet struct {
	PrivateKey *ecdsa.PrivateKey
	PublicKey  *ecdsa.PublicKey
	Address    common.Address
}

// GenerateWallet creates a new wallet
func GenerateWallet() (*Wallet, error) {
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		return nil, fmt.Errorf("failed to generate private key: %w", err)
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("failed to cast public key")
	}

	address := crypto.PubkeyToAddress(*publicKeyECDSA)

	return &Wallet{
		PrivateKey: privateKey,
		PublicKey:  publicKeyECDSA,
		Address:    address,
	}, nil
}

// LoadWallet loads a wallet from a private key hex string
func LoadWallet(privateKeyHex string) (*Wallet, error) {
	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		return nil, fmt.Errorf("invalid private key: %w", err)
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("failed to cast public key")
	}

	address := crypto.PubkeyToAddress(*publicKeyECDSA)

	return &Wallet{
		PrivateKey: privateKey,
		PublicKey:  publicKeyECDSA,
		Address:    address,
	}, nil
}

// GetPrivateKeyHex returns the private key as hex string
func (w *Wallet) GetPrivateKeyHex() string {
	return fmt.Sprintf("%x", crypto.FromECDSA(w.PrivateKey))
}

// GetAddressHex returns the address as hex string
func (w *Wallet) GetAddressHex() string {
	return w.Address.Hex()
}

// SignMessage signs a message with the wallet's private key
func (w *Wallet) SignMessage(message []byte) ([]byte, error) {
	hash := crypto.Keccak256Hash(message)
	signature, err := crypto.Sign(hash.Bytes(), w.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign message: %w", err)
	}
	return signature, nil
}

// VerifySignature verifies a signature
func VerifySignature(message, signature []byte, address common.Address) bool {
	hash := crypto.Keccak256Hash(message)
	
	if len(signature) != 65 {
		return false
	}

	// Remove recovery ID
	sig := signature[:64]
	recoveryID := signature[64]

	// Recover public key
	pubKey, err := crypto.SigToPub(hash.Bytes(), signature)
	if err != nil {
		return false
	}

	recoveredAddr := crypto.PubkeyToAddress(*pubKey)
	
	// Verify signature
	return crypto.VerifySignature(
		crypto.FromECDSAPub(pubKey),
		hash.Bytes(),
		sig,
	) && recoveryID < 2 && recoveredAddr == address
}

// WeiToEth converts wei to eth
func WeiToEth(wei *big.Int) *big.Float {
	return new(big.Float).Quo(
		new(big.Float).SetInt(wei),
		big.NewFloat(1e18),
	)
}

// EthToWei converts eth to wei
func EthToWei(eth *big.Float) *big.Int {
	truncInt, _ := new(big.Float).Mul(
		eth,
		big.NewFloat(1e18),
	).Int(nil)
	return truncInt
}
