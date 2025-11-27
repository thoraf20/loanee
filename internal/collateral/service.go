package collateral

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/thoraf20/loanee/config"
	"github.com/thoraf20/loanee/internal/blockchain"
	"github.com/thoraf20/loanee/internal/loan"
	"github.com/thoraf20/loanee/internal/models"
	"github.com/thoraf20/loanee/internal/pricing"
)

type Service struct {
	repo        Repository
	pricing     pricing.Provider
	verifier    blockchain.Verifier
	loanService *loan.Service
	cfg         *config.Config
	logger      zerolog.Logger
}

func NewService(repo Repository, pricing pricing.Provider, verifier blockchain.Verifier, loanService *loan.Service, cfg *config.Config, logger zerolog.Logger) *Service {
	return &Service{
		repo:        repo,
		pricing:     pricing,
		verifier:    verifier,
		loanService: loanService,
		cfg:         cfg,
		logger:      logger,
	}
}

func (s *Service) PreviewCollateral(ctx context.Context, loanAmount float64, fiatCurrency string) (*PreviewResponse, error) {
	fiat := normalizeFiat(fiatCurrency)
	if loanAmount <= 0 {
		return nil, fmt.Errorf("loan amount must be positive")
	}

	prices, err := s.pricing.GetPrices(SupportedAssets, fiat)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch prices: %w", err)
	}

	ltv := s.cfg.Loan.DefaultLTV
	if ltv <= 0 {
		return nil, fmt.Errorf("invalid default LTV configuration")
	}

	requiredValue := loanAmount / ltv
	previews := make([]PreviewItem, 0, len(prices))
	for symbol, price := range prices {
		requiredAmount := requiredValue / price
		previews = append(previews, PreviewItem{
			AssetSymbol:    symbol,
			FiatCurrency:   strings.ToUpper(fiat),
			LoanAmount:     loanAmount,
			CollateralLTV:  ltv,
			AssetPrice:     price,
			RequiredValue:  roundTo(requiredValue, 2),
			RequiredAmount: roundTo(requiredAmount, 8),
			Status:         string(models.StatusPreview),
		})
	}

	return &PreviewResponse{
		FiatCurrency: strings.ToUpper(fiat),
		LoanAmount:   loanAmount,
		Previews:     previews,
	}, nil
}

func (s *Service) CreateCollateralRequest(ctx context.Context, req CreateRequest) (*models.Collateral, error) {
	price, err := s.pricing.GetPrice(req.AssetSymbol, req.FiatCurrency)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch %s price: %w", req.AssetSymbol, err)
	}

	ltv := s.cfg.Loan.DefaultLTV
	if ltv <= 0 {
		return nil, fmt.Errorf("invalid default LTV configuration")
	}

	requiredValue := req.LoanAmount / ltv
	requiredAmount := requiredValue / price

	now := time.Now()
	collateral := &models.Collateral{
		ID:            uuid.New(),
		UserID:        req.UserID,
		AssetSymbol:   req.AssetSymbol,
		AssetAmount:   requiredAmount,
		AssetValue:    price * requiredAmount,
		RequiredValue: requiredValue,
		FiatCurrency:  req.FiatCurrency,
		FiatAmount:    req.LoanAmount,
		LTV:           ltv,
		Status:        models.StatusPending,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	if err := s.repo.Create(ctx, collateral); err != nil {
		return nil, err
	}

	if s.loanService != nil {
		if _, err := s.loanService.CreateFromCollateral(ctx, collateral); err != nil {
			s.logger.Warn().Err(err).Msg("failed to create loan from collateral")
		}
	}

	return collateral, nil
}

func (s *Service) LockCollateral(ctx context.Context, userID uuid.UUID, req LockRequest) (*models.Collateral, error) {
	if s.verifier != nil {
		valid, _, err := s.verifier.VerifyTransaction(ctx, req.TxHash, req.AssetSymbol, req.Amount)
		if err != nil {
			return nil, fmt.Errorf("transaction verification failed: %w", err)
		}
		if !valid {
			return nil, fmt.Errorf("transaction %s could not be verified", req.TxHash)
		}
	}

	price, err := s.pricing.GetPrice(req.AssetSymbol, req.FiatCurrency)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch %s price: %w", req.AssetSymbol, err)
	}

	assetValue := req.Amount * price
	ltv := s.cfg.Loan.DefaultLTV
	loanValue := assetValue * ltv

	now := time.Now()
	collateral := &models.Collateral{
		ID:            uuid.New(),
		UserID:        userID,
		AssetSymbol:   req.AssetSymbol,
		AssetAmount:   req.Amount,
		AssetValue:    assetValue,
		RequiredValue: assetValue,
		FiatCurrency:  req.FiatCurrency,
		FiatAmount:    loanValue,
		LTV:           ltv,
		Status:        models.StatusActive,
		TxHash:        &req.TxHash,
		WalletAddress: optionalString(req.WalletAddress),
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	if err := s.repo.Create(ctx, collateral); err != nil {
		return nil, err
	}

	return collateral, nil
}

func (s *Service) ListUserCollaterals(ctx context.Context, userID uuid.UUID) ([]models.Collateral, error) {
	return s.repo.GetByUserID(ctx, userID)
}

func (s *Service) ListAllCollaterals(ctx context.Context) ([]models.Collateral, error) {
	return s.repo.ListAll(ctx)
}

func (s *Service) RequestRelease(ctx context.Context, userID, collateralID uuid.UUID) (*models.Collateral, error) {
	collateral, err := s.repo.GetByID(ctx, collateralID)
	if err != nil {
		return nil, err
	}
	if collateral == nil {
		return nil, fmt.Errorf("collateral not found")
	}
	if collateral.UserID != userID {
		return nil, errors.New("not authorized to modify this collateral")
	}
	if collateral.Status != models.StatusActive {
		return nil, fmt.Errorf("collateral must be active to request release")
	}

	now := time.Now()
	collateral.Status = models.StatusReleaseRequested
	collateral.ReleaseRequestedAt = &now
	collateral.ReleaseResolvedAt = nil
	collateral.ReleaseNote = nil

	if err := s.repo.Update(ctx, collateral); err != nil {
		return nil, err
	}
	return collateral, nil
}

func (s *Service) ApproveRelease(ctx context.Context, collateralID uuid.UUID) (*models.Collateral, error) {
	collateral, err := s.repo.GetByID(ctx, collateralID)
	if err != nil {
		return nil, err
	}
	if collateral == nil {
		return nil, fmt.Errorf("collateral not found")
	}
	if collateral.Status != models.StatusReleaseRequested {
		return nil, fmt.Errorf("collateral not awaiting release")
	}

	now := time.Now()
	collateral.Status = models.StatusReleased
	collateral.ReleaseResolvedAt = &now
	collateral.ReleaseNote = nil

	if err := s.repo.Update(ctx, collateral); err != nil {
		return nil, err
	}
	return collateral, nil
}

func (s *Service) RejectRelease(ctx context.Context, collateralID uuid.UUID, reason string) (*models.Collateral, error) {
	collateral, err := s.repo.GetByID(ctx, collateralID)
	if err != nil {
		return nil, err
	}
	if collateral == nil {
		return nil, fmt.Errorf("collateral not found")
	}
	if collateral.Status != models.StatusReleaseRequested {
		return nil, fmt.Errorf("collateral not awaiting release")
	}

	now := time.Now()
	collateral.Status = models.StatusActive
	collateral.ReleaseResolvedAt = &now
	if reason != "" {
		collateral.ReleaseNote = &reason
	}

	if err := s.repo.Update(ctx, collateral); err != nil {
		return nil, err
	}
	return collateral, nil
}

func optionalString(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}

func normalizeFiat(fiat string) string {
	if fiat == "" {
		return "USD"
	}
	return strings.ToUpper(strings.TrimSpace(fiat))
}

func roundTo(value float64, precision int) float64 {
	pow := math.Pow10(precision)
	return math.Round(value*pow) / pow
}
