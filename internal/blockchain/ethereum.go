package blockchain

import (
	"context"
	"errors"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/rs/zerolog/log"
)

// TransactionData captures basic on-chain transaction metadata.
type TransactionData struct {
	Hash          string
	From          string
	To            string
	Amount        float64
	Confirmations int64
}

// Verifier defines the behaviour required to verify a blockchain transaction.
type Verifier interface {
	VerifyTransaction(ctx context.Context, txHash, assetSymbol string, expectedAmount float64) (bool, *TransactionData, error)
}

// EthereumVerifier implements on-chain verification against an Ethereum RPC endpoint.
type EthereumVerifier struct {
	client           *ethclient.Client
	minConfirmations int64
}

// NewEthereumVerifier dials an RPC endpoint and returns a verifier instance.
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

// VerifyTransaction checks that a transaction exists on-chain, has the required confirmations,
// and roughly matches the expected amount (in ETH units).
func (v *EthereumVerifier) VerifyTransaction(ctx context.Context, txHash string, assetSymbol string, expectedAmount float64) (bool, *TransactionData, error) {
	hash := common.HexToHash(txHash)
	tx, isPending, err := v.client.TransactionByHash(ctx, hash)
	if err != nil {
		return false, nil, fmt.Errorf("failed to fetch transaction: %w", err)
	}
	if isPending {
		return false, nil, errors.New("transaction is still pending")
	}

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
		log.Warn().
			Int64("confirmations", confirmations).
			Int64("required", v.minConfirmations).
			Msg("transaction does not have enough confirmations")
		return false, nil, fmt.Errorf("transaction has only %d confirmations", confirmations)
	}

	to := ""
	if tx.To() != nil {
		to = tx.To().Hex()
	}
	from, err := v.getSenderAddress(ctx, tx)
	if err != nil {
		log.Warn().Err(err).Msg("could not determine sender address")
	}

	valueEth := new(big.Float).Quo(new(big.Float).SetInt(tx.Value()), big.NewFloat(1e18))
	expected := big.NewFloat(expectedAmount)
	diff := new(big.Float).Sub(valueEth, expected)

	if diff.Cmp(big.NewFloat(0.000001)) > 0 {
		actual, _ := valueEth.Float64()
		expectedFloat, _ := expected.Float64()
		return false, nil, fmt.Errorf("transaction value %.6f ETH does not match expected %.6f ETH", actual, expectedFloat)
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

func (v *EthereumVerifier) getSenderAddress(ctx context.Context, tx *types.Transaction) (string, error) {
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

// NoopVerifier accepts every transaction. Useful for local development.
type NoopVerifier struct{}

func NewNoopVerifier() Verifier {
	return &NoopVerifier{}
}

func (n *NoopVerifier) VerifyTransaction(ctx context.Context, txHash, assetSymbol string, expectedAmount float64) (bool, *TransactionData, error) {
	return true, &TransactionData{
		Hash:   txHash,
		Amount: expectedAmount,
	}, nil
}
