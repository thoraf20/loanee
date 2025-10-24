package services

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/thoraf20/loanee/internal/dtos"
	"github.com/thoraf20/loanee/internal/models"
	repository "github.com/thoraf20/loanee/internal/repo"
)

const DEFAULT_LTV = 0.65

type CollateralService interface {
	PreviewCollateral(loanAmount float64, fiatCurrency string) (*dtos.CollateralPreviewResponse, error)
	LockCollateral(ctx context.Context, userID uuid.UUID, dto dtos.CollateralLockDTO) (*models.Collateral, error)
}

type collateralService struct {
	repo         repository.CollateralRepository
	priceService PriceService
}

func NewCollateralService(repo repository.CollateralRepository, priceService PriceService) CollateralService {
	return &collateralService{ repo: repo, priceService: priceService }
}

func (s *collateralService) PreviewCollateral(loanAmount float64, fiatCurrency string) (*dtos.CollateralPreviewResponse, error) {
	if loanAmount <= 0 {
		return nil, fmt.Errorf("loan amount must be greater than zero")
	}

	prices, err := s.priceService.GetPrices(fiatCurrency)
	if err != nil {
		return nil, fmt.Errorf("failed to get prices: %w", err)
	}

	if len(prices) == 0 {
		return nil, fmt.Errorf("no crypto prices available for %s", fiatCurrency)
	}

	requiredCollateral := loanAmount / DEFAULT_LTV

	previews := make([]dtos.CollateralPreviewItem, 0, len(prices))
	for asset, price := range prices {
		collateralAmount := requiredCollateral / price
		previews = append(previews, dtos.CollateralPreviewItem{
			AssetSymbol:   asset,
			CollateralLTV: DEFAULT_LTV,
			FiatCurrency:  fiatCurrency,
			LoanAmount:    loanAmount,
			AssetPrice:    price,
			RequiredAmount:   math.Round(collateralAmount*1e8) / 1e8,
			Status:        "preview",
		})
	}

	return &dtos.CollateralPreviewResponse{
		FiatCurrency: fiatCurrency,
		LoanAmount:   loanAmount,
		Previews:     previews,
	}, nil
}

func (s *collateralService) LockCollateral(ctx context.Context, userID uuid.UUID, dto dtos.CollateralLockDTO) (*models.Collateral, error) {
	verified := true // Replace with actual blockchain verification logic
	if !verified {
		return nil, errors.New("transaction could not be verified")
	}

	price, err := s.priceService.GetPrice(dto.AssetSymbol, "USD")
	if err != nil {
		return nil, fmt.Errorf("could not fetch asset price: %w", err)
	}

	collateral := &models.Collateral{
		ID:            uuid.New(),
		UserID:        userID,
		LoanRequestID: uuid.MustParse(dto.LoanRequestID),
		AssetSymbol:   dto.AssetSymbol,
		AssetAmount:   dto.Amount,
		UsdValue:      price * dto.Amount,
		FiatCurrency:  "USD",
		Status:        models.StatusActive,
		TxHash:        &dto.TxHash,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := s.repo.Create(ctx, collateral); err != nil {
		return nil, fmt.Errorf("failed to save collateral: %w", err)
	}

	return collateral, nil
}