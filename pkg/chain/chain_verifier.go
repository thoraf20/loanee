package chain

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/ethclient"
)

type ChainVerifier interface {
  VerifyTransaction(ctx context.Context, txHash, assetSymbol, expectedAddress string, expectedAmount float64) (bool, error)
}

type EthereumVerifier struct {
	client           *ethclient.Client
	minConfirmations int64
}

func NewEthereumVerifier(rpcURL string, minConfirmations int64) (*EthereumVerifier, error) {
	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Ethereum node: %w", err)
	}

	return &EthereumVerifier{
		client:           client,
		minConfirmations: minConfirmations,
	}, nil
}

// var _ ChainVerifier = (*EthereumVerifi er)(nil)
