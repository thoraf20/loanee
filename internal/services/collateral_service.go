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
	return &collateralService{ priceService: priceService }
}

func (s *collateralService) PreviewCollateral(loanAmount float64, fiatCurrency string) (*dtos.CollateralPreviewResponse, error) {
	if loanAmount <= 0 {
		return nil, fmt.Errorf("loan amount must be greater than zero")
	}

	supportedAssets := []string{"ETH", "USDT", "BTC"}
	results := make([]dtos.CollateralPreviewItem, 0)

	requiredCollateralValue := loanAmount / DEFAULT_LTV

	for _, asset := range supportedAssets {
		price, err := s.priceService.GetPrice(asset, fiatCurrency)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch price for %s: %w", asset, err)
		}

		requiredAmount := requiredCollateralValue / price

		results = append(results, dtos.CollateralPreviewItem{
			Asset:           asset,
			FiatCurrency:    fiatCurrency,
			LoanAmount:      loanAmount,
			LTV:             DEFAULT_LTV,
			AssetPrice:      price,
			RequiredValue:   math.Round(requiredCollateralValue*1e2) / 1e2,
			RequiredAmount:  math.Round(requiredAmount*1e8) / 1e8,
			Status:          "preview",
		})
	}

	return &dtos.CollateralPreviewResponse{
		FiatCurrency: fiatCurrency,
		LoanAmount:   loanAmount,
		Previews:     results,
	}, nil
}