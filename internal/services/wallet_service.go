package services

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"log"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/google/uuid"
	"github.com/thoraf20/loanee/internal/models"
	repository "github.com/thoraf20/loanee/internal/repo"
)

type WalletService interface {
	GenerateUserWallets(ctx context.Context, userID uuid.UUID) error
	GetUserWallets(ctx context.Context, userID uuid.UUID) ([]models.Wallet, error)
}

type walletService struct {
	walletRepo repository.WalletRepository
	userRepo repository.UserRepository
}

func NewWalletService(walletRepo repository.WalletRepository, userRepo repository.UserRepository) *walletService {
	if walletRepo == nil {
    panic("nil WalletRepository passed to WalletService")
  }
	if userRepo == nil {
		panic("nil UserRepository passed to WalletService")
	}
	return &walletService{
		walletRepo:     walletRepo,
		userRepo: userRepo,
	}
}

func (s *walletService) GenerateUserWallets(ctx context.Context, userID uuid.UUID) error {
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return errors.New("user not found")
	}

	coins := []string{"BTC", "ETH", "USDT"}

	for _, coin := range coins {
		address, privKey, err := generateWallet(coin)
		if err != nil {
			log.Printf("Failed to generate %s wallet: %v", coin, err)
			return fmt.Errorf("failed to generate %s wallet: %w", coin, err)
		}

		wallet := &models.Wallet{
			UserID:     userID,
			AssetType:  coin,
			Address:    address,
			PrivateKey: privKey,
			Balance:    0,
			IsPrimary:  false,
		}

		if err := s.walletRepo.Create(ctx, wallet); err != nil {
			return fmt.Errorf("failed to save %s wallet: %w", coin, err)
		}
	}

	log.Printf("Wallets successfully generated for user %s", user.Email)

	return nil
}

func (s *walletService) GetUserWallets(ctx context.Context, userID uuid.UUID) ([]models.Wallet, error) {
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user: %w", err)
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	wallets, err := s.walletRepo.GetUserWallets(ctx, userID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to get user wallets: %w", err)
	}

	return wallets, nil
}

func generateWallet(coin string) (string, string, error) {
	switch coin {
	case "ETH", "USDT":
		return generateETHWallet()
	case "BTC":
		return generateBTCWallet()
	default:
		return "", "", fmt.Errorf("unsupported coin: %s", coin)
	}
}

func generateETHWallet() (string, string, error) {
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		return "", "", err
	}

	privateKeyBytes := crypto.FromECDSA(privateKey)
	address := crypto.PubkeyToAddress(privateKey.PublicKey).Hex()

	return address, hex.EncodeToString(privateKeyBytes), nil
}

// placeholder: BTC generation to be replaced with real logic later
func generateBTCWallet() (string, string, error) {
	return "btc_mock_address_" + uuid.New().String(), "btc_private_key_mock", nil
}