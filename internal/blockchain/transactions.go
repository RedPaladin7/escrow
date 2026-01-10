package blockchain

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/sirupsen/logrus"
)

type TransactionStatus string

const (
	TxStatusPending   TransactionStatus = "pending"
	TxStatusConfirmed TransactionStatus = "confirmed"
	TxStatusFailed    TransactionStatus = "failed"
)

type Transaction struct {
	Hash        common.Hash
	From        common.Address
	To          common.Address
	Value       *big.Int
	GasPrice    *big.Int
	GasLimit    uint64
	Nonce       uint64
	Data        []byte
	Status      TransactionStatus
	BlockNumber *big.Int
	Timestamp   time.Time
}

// SendTransaction sends a raw transaction
func (bc *BlockchainClient) SendTransaction(to common.Address, value *big.Int, data []byte) (*types.Transaction, error) {
	nonce, err := bc.client.PendingNonceAt(context.Background(), bc.publicAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to get nonce: %w", err)
	}

	gasPrice, err := bc.client.SuggestGasPrice(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to get gas price: %w", err)
	}

	gasLimit := uint64(21000)
	if len(data) > 0 {
		gasLimit = uint64(3000000)
	}

	tx := types.NewTransaction(
		nonce,
		to,
		value,
		gasLimit,
		gasPrice,
		data,
	)

	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(bc.chainID), bc.privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign transaction: %w", err)
	}

	err = bc.client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		return nil, fmt.Errorf("failed to send transaction: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"tx_hash": signedTx.Hash().Hex(),
		"to":      to.Hex(),
		"value":   value.String(),
	}).Info("Transaction sent")

	return signedTx, nil
}

// WaitForTransaction waits for a transaction to be mined
func (bc *BlockchainClient) WaitForTransaction(txHash common.Hash, timeout time.Duration) (*types.Receipt, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("timeout waiting for transaction")
		case <-ticker.C:
			receipt, err := bc.client.TransactionReceipt(context.Background(), txHash)
			if err == nil {
				if receipt.Status == types.ReceiptStatusSuccessful {
					logrus.WithFields(logrus.Fields{
						"tx_hash":      txHash.Hex(),
						"block_number": receipt.BlockNumber.String(),
						"gas_used":     receipt.GasUsed,
					}).Info("Transaction confirmed")
					return receipt, nil
				}
				return receipt, fmt.Errorf("transaction failed")
			}
		}
	}
}

// GetTransactionStatus gets the status of a transaction
func (bc *BlockchainClient) GetTransactionStatus(txHash common.Hash) (TransactionStatus, error) {
	receipt, err := bc.client.TransactionReceipt(context.Background(), txHash)
	if err != nil {
		// Transaction not yet mined
		return TxStatusPending, nil
	}

	if receipt.Status == types.ReceiptStatusSuccessful {
		return TxStatusConfirmed, nil
	}

	return TxStatusFailed, nil
}

// EstimateGas estimates gas for a transaction
func (bc *BlockchainClient) EstimateGas(to common.Address, value *big.Int, data []byte) (uint64, error) {
	msg := types.NewMessage(
		bc.publicAddress,
		&to,
		0,
		value,
		3000000,
		big.NewInt(0),
		big.NewInt(0),
		big.NewInt(0),
		data,
		nil,
		false,
	)

	gas, err := bc.client.EstimateGas(context.Background(), msg)
	if err != nil {
		return 0, fmt.Errorf("failed to estimate gas: %w", err)
	}

	return gas, nil
}
