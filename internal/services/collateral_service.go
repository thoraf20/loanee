package services

import (
	"fmt"
	"math"

	"github.com/thoraf20/loanee/internal/dtos"
)

const DEFAULT_LTV = 0.65

type CollateralService interface {
	PreviewCollateral(loanAmount float64, fiatCurrency string) (*dtos.CollateralPreviewResponse, error)
}

type collateralService struct {
	priceService PriceService
}

func NewCollateralService(priceService PriceService) CollateralService {
	return &collateralService{priceService: priceService}
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