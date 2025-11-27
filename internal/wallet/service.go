package wallet

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/thoraf20/loanee/internal/models"
)

type Service struct {
	repo   Repository
	logger zerolog.Logger
}

func NewService(repo Repository, logger zerolog.Logger) *Service {
	return &Service{
		repo:   repo,
		logger: logger.With().Str("component", "wallet_service").Logger(),
	}
}

func (s *Service) ListUserWallets(ctx context.Context, userID uuid.UUID) ([]models.Wallet, error) {
	return s.repo.ListByUser(ctx, userID)
}

func (s *Service) GetOrCreatePrimary(ctx context.Context, userID uuid.UUID, asset string) (*models.Wallet, error) {
	if wallet, err := s.repo.GetPrimaryByAsset(ctx, userID, asset); err != nil {
		return nil, err
	} else if wallet != nil {
		return wallet, nil
	}

	newWallet := &models.Wallet{
		ID:         uuid.New(),
		UserID:     userID,
		AssetType:  asset,
		Address:    fmt.Sprintf("auto-generated-%s-%s", asset, uuid.New().String()),
		Balance:    0,
		IsPrimary:  true,
		PrivateKey: "", // never expose; placeholder until real generation
	}

	if err := s.repo.Create(ctx, newWallet); err != nil {
		return nil, err
	}
	return newWallet, nil
}
