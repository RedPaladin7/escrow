package blockchain

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/sirupsen/logrus"
)

type BlockchainClient struct {
	client              *ethclient.Client
	chainID             *big.Int
	privateKey          *ecdsa.PrivateKey
	publicAddress       common.Address
	pokerTableAddress   common.Address
	potManagerAddress   common.Address
	playerRegistryAddress common.Address
	disputeResolverAddress common.Address
	
	pokerTable      *PokerTable
	potManager      *PotManager
	playerRegistry  *PlayerRegistry
	disputeResolver *DisputeResolver
}

type Config struct {
	RPCURL                  string
	PrivateKey              string
	PokerTableAddress       string
	PotManagerAddress       string
	PlayerRegistryAddress   string
	DisputeResolverAddress  string
}

func NewBlockchainClient(cfg *Config) (*BlockchainClient, error) {
	client, err := ethclient.Dial(cfg.RPCURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to blockchain: %w", err)
	}

	chainID, err := client.ChainID(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to get chain ID: %w", err)
	}

	privateKey, err := crypto.HexToECDSA(cfg.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("invalid private key: %w", err)
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("failed to cast public key to ECDSA")
	}

	publicAddress := crypto.PubkeyToAddress(*publicKeyECDSA)

	bc := &BlockchainClient{
		client:                 client,
		chainID:                chainID,
		privateKey:             privateKey,
		publicAddress:          publicAddress,
		pokerTableAddress:      common.HexToAddress(cfg.PokerTableAddress),
		potManagerAddress:      common.HexToAddress(cfg.PotManagerAddress),
		playerRegistryAddress:  common.HexToAddress(cfg.PlayerRegistryAddress),
		disputeResolverAddress: common.HexToAddress(cfg.DisputeResolverAddress),
	}

	// Initialize contract instances (these will be generated from ABIs)
	// pokerTable, err := NewPokerTable(bc.pokerTableAddress, client)
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to instantiate poker table contract: %w", err)
	// }
	// bc.pokerTable = pokerTable

	logrus.WithFields(logrus.Fields{
		"address":  publicAddress.Hex(),
		"chain_id": chainID.String(),
	}).Info("Blockchain client initialized")

	return bc, nil
}

func (bc *BlockchainClient) GetTransactor() (*bind.TransactOpts, error) {
	nonce, err := bc.client.PendingNonceAt(context.Background(), bc.publicAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to get nonce: %w", err)
	}

	gasPrice, err := bc.client.SuggestGasPrice(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to get gas price: %w", err)
	}

	auth, err := bind.NewKeyedTransactorWithChainID(bc.privateKey, bc.chainID)
	if err != nil {
		return nil, fmt.Errorf("failed to create transactor: %w", err)
	}

	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(0)
	auth.GasLimit = uint64(3000000)
	auth.GasPrice = gasPrice

	return auth, nil
}

func (bc *BlockchainClient) GetCallOpts() *bind.CallOpts {
	return &bind.CallOpts{
		From: bc.publicAddress,
	}
}

func (bc *BlockchainClient) Close() {
	if bc.client != nil {
		bc.client.Close()
	}
}

func (bc *BlockchainClient) GetBalance(address common.Address) (*big.Int, error) {
	balance, err := bc.client.BalanceAt(context.Background(), address, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get balance: %w", err)
	}
	return balance, nil
}

func (bc *BlockchainClient) GetMyBalance() (*big.Int, error) {
	return bc.GetBalance(bc.publicAddress)
}
