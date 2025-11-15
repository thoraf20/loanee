package services

import (
	"context"
	"errors"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/rs/zerolog/log"
	// "github.com/ethereum/go-ethereum/crypto"
)

type EthereumVerifier struct {
	client           *ethclient.Client
	minConfirmations int64
}

type TransactionData struct {
	Hash          string
	From          string
	To            string
	Amount        float64
	Confirmations int64
}

type TransactionVerifier interface {
	VerifyTransaction(ctx context.Context, txHash, assetSymbol string , expectedAmount float64) (bool, *TransactionData, error)
}

func NewEthereumVerifier(rpcURL string, minConfirmations int) (*EthereumVerifier, error) {
	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Ethereum RPC: %w", err)
	}
	return &EthereumVerifier{
		client:           client,
		minConfirmations: int64(minConfirmations),
	}, nil
}

func (v *EthereumVerifier) VerifyTransaction(ctx context.Context, txHash string, assetSymbol string, expectedAmount float64) (bool, *TransactionData, error) {
	hash := common.HexToHash(txHash)
	tx, isPending, err := v.client.TransactionByHash(ctx, hash)
	if err != nil {
		return false, nil, fmt.Errorf("failed to fetch transaction: %w", err)
	}
	if isPending {
		return false, nil, errors.New("transaction is still pending")
	}


	// Get transaction receipt for confirmation info
	receipt, err := v.client.TransactionReceipt(ctx, hash)
	if err != nil {
		return false, nil, fmt.Errorf("could not fetch transaction receipt: %w", err)
	}

	blockHeader, err := v.client.HeaderByNumber(ctx, nil)
	if err != nil {
		return false, nil, fmt.Errorf("could not fetch latest block header: %w", err)
	}

	currentBlock := blockHeader.Number.Int64()
	txBlock := receipt.BlockNumber.Int64()
	confirmations := currentBlock - txBlock

	if confirmations < v.minConfirmations {
		log.Warn().Msgf("transaction has only %d confirmations (need %d)", confirmations, v.minConfirmations)
		return false, nil, fmt.Errorf("transaction has only %d confirmations", confirmations)
	}

	to := ""
	if tx.To() != nil {
		to = tx.To().Hex()
	}
	from, err := v.getSenderAddress(ctx, tx, receipt)
	if err != nil {
		log.Warn().Err(err).Msg("could not determine sender address")
	}

	// Check amount (convert wei → ETH)
	valueEth := new(big.Float).Quo(new(big.Float).SetInt(tx.Value()), big.NewFloat(1e18))
	expected := big.NewFloat(expectedAmount)
	diff := new(big.Float).Sub(valueEth, expected)

	if diff.Cmp(big.NewFloat(0.000001)) > 0 {
		return false, nil, fmt.Errorf("transaction value %.6f ETH does not match expected %.6f ETH", valueEth, expected)
	}

	amount, _ := valueEth.Float64()
	txData := &TransactionData{
		Hash:          txHash,
		From:          from,
		To:            to,
		Amount:        amount,
		Confirmations: confirmations,
	}

	return true, txData, nil
}

func (v *EthereumVerifier) getSenderAddress(ctx context.Context, tx *types.Transaction, receipt *types.Receipt) (string, error) {
	chainID, err := v.client.NetworkID(ctx)
	if err != nil {
		return "", err
	}
	signer := types.LatestSignerForChainID(chainID)
	from, err := types.Sender(signer, tx)
	if err != nil {
		return "", err
	}
	return from.Hex(), nil
}