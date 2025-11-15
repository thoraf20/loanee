package services

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/thoraf20/loanee/config"
	"github.com/thoraf20/loanee/internal/dtos"
	"github.com/thoraf20/loanee/internal/models"
	repository "github.com/thoraf20/loanee/internal/repo"
)

type CollateralService interface {
	PreviewCollateral(ctx context.Context, loanAmount float64, fiatCurrency string) (*dtos.CollateralPreviewResponse, error)
	LockCollateral(ctx context.Context, userID uuid.UUID, dto dtos.CollateralLockDTO) (*models.Collateral, error)
	CreateCollateralRequest(ctx context.Context, dto dtos.CreateCollateralRequestDTO) (*models.Collateral, error)
	VerifyTransaction(ctx context.Context, req dtos.VerifyCollateralRequestDTO) (*models.Collateral, error)
}

type collateralService struct {
	repo         repository.CollateralRepository
	priceService CoinGeckoProvider
	verifier     TransactionVerifier
	config       *config.Config
	walletRepo 	 repository.WalletRepository
}

func NewCollateralService(repo repository.CollateralRepository, priceService CoinGeckoProvider, verifier TransactionVerifier, walletRepo repository.WalletRepository) CollateralService {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load config in collateral service")
	}
	return &collateralService{
		repo:         repo,
		priceService: priceService,
		verifier:     verifier,
		config:       cfg,
		walletRepo:   walletRepo,
	}
}

func (s *collateralService) PreviewCollateral(ctx context.Context, loanAmount float64, fiatCurrency string) (*dtos.CollateralPreviewResponse, error) {

	prices, err := s.priceService.GetPrices([]string{"BTC"}, fiatCurrency)
	if err != nil {
		return nil, fmt.Errorf("failed to get prices: %w", err)
	}

	if len(prices) == 0 {
		return nil, fmt.Errorf("no crypto prices available for %s", fiatCurrency)
	}

	defaultLTV, _ := strconv.ParseFloat(s.config.DefaultLTV, 64)
	requiredCollateral := loanAmount / defaultLTV

	previews := make([]dtos.CollateralPreviewItem, 0, len(prices))
	for asset, price := range prices {
		collateralAmount := requiredCollateral / price
		previews = append(previews, dtos.CollateralPreviewItem{
			AssetSymbol:   asset,
			CollateralLTV: defaultLTV,
			FiatCurrency:  fiatCurrency,
			LoanAmount:    loanAmount,
			AssetPrice:    price,
			RequiredAmount:   math.Round(collateralAmount*1e8) / 1e8,
			Status:        string(models.StatusPreview),
		})
	}

	return &dtos.CollateralPreviewResponse{
		FiatCurrency: fiatCurrency,
		LoanAmount:   loanAmount,
		Previews:     previews,
	}, nil
}

func (s *collateralService) LockCollateral(ctx context.Context, userID uuid.UUID, dto dtos.CollateralLockDTO) (*models.Collateral, error) {
	valid, _, err := s.verifier.VerifyTransaction(ctx, dto.TxHash, dto.AssetSymbol, dto.Amount)
	if err != nil {
		return nil, fmt.Errorf("transaction verification failed: %w", err)
	}
	if !valid {
		return nil, errors.New("transaction is invalid or unconfirmed")
	}

	price, err := s.priceService.GetPrice(dto.AssetSymbol, dto.FiatCurrency)
	if err != nil {
		return nil, fmt.Errorf("could not fetch asset price: %w", err)
	}	

	collateral := &models.Collateral{
		ID:            uuid.New(),
		UserID:        userID,
		AssetSymbol:   dto.AssetSymbol,
		AssetAmount:   dto.Amount,
		RequiredValue: price * dto.Amount,
		FiatCurrency:  dto.FiatCurrency,
		Status:        models.StatusActive,
		TxHash:        &dto.TxHash,
		WalletAddress: &dto.WalletAddress,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := s.repo.Create(ctx, collateral); err != nil {
		return nil, fmt.Errorf("failed to save collateral: %w", err)
	}

	return collateral, nil
}

func (s *collateralService) VerifyTransaction(ctx context.Context, req dtos.VerifyCollateralRequestDTO) (*models.Collateral, error) {
	collateral, err := s.repo.GetCollateralByID(ctx, req.CollateralID)
	if err != nil {
		return nil, fmt.Errorf("collateral not found: %w", err)
	}

	// Use the stored wallet address (if any) and the stored asset amount to verify the transaction.
	walletAddr := ""
	if collateral.WalletAddress != nil {
		walletAddr = *collateral.WalletAddress
	}

	isValid, txnData, err := s.verifier.VerifyTransaction(ctx, req.TransactionHash, collateral.AssetSymbol, collateral.RequiredValue)
	if err != nil {
		return nil, fmt.Errorf("verification failed: %w", err)
	}
	if !isValid || txnData.Amount != collateral.AssetValue {
		return nil, errors.New("transaction verification failed")
	}

	// Update collateral fields to reflect the confirmation
	collateral.Status = models.StatusActive
	collateral.TxHash = &req.TransactionHash
	if walletAddr != "" {
		collateral.WalletAddress = &walletAddr
	}

	// Update transaction info and status
	if err := s.repo.UpdateCollateralTxInfo(ctx, req.CollateralID, req.TransactionHash, walletAddr, "active"); err != nil {
		return nil, fmt.Errorf("failed to update collateral: %w", err)
	}

	return collateral, nil
}

func (s *collateralService) CreateCollateralRequest(ctx context.Context, dto dtos.CreateCollateralRequestDTO) (*models.Collateral, error) {
	price, err := s.priceService.GetPrice(dto.AssetSymbol, dto.FiatCurrency)
	if err != nil {
		return nil, fmt.Errorf("could not fetch asset price: %w", err)
	}

	defaultLoaToValue, _ := strconv.ParseFloat(s.config.DefaultLTV, 64)
	requiredCollateralValue := dto.LoanAmount / defaultLoaToValue
	requiredCryptoAmount := requiredCollateralValue / price

	collateral := &models.Collateral{
		ID:               uuid.New(),
		UserID:           dto.UserID,
		AssetSymbol:      dto.AssetSymbol,
		AssetAmount:      requiredCryptoAmount,
		RequiredValue:    price * requiredCryptoAmount,
		FiatCurrency:     dto.FiatCurrency,
		FiatAmount:       dto.LoanAmount,
		Status:           models.StatusPending,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	if err := s.repo.Create(ctx, collateral); err != nil {
		return nil, fmt.Errorf("failed to save collateral: %w", err)
	}

	return collateral, nil
}