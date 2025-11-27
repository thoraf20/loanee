package collateral

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
	"github.com/thoraf20/loanee/config"
	"github.com/thoraf20/loanee/internal/blockchain"
	"github.com/thoraf20/loanee/internal/models"
)

func TestPreviewCollateral(t *testing.T) {
	service, _ := newTestService()

	resp, err := service.PreviewCollateral(context.Background(), 1000, "USD")
	require.NoError(t, err)
	require.Len(t, resp.Previews, len(SupportedAssets))
}

func TestCreateCollateralRequest(t *testing.T) {
	service, repo := newTestService()

	collateral, err := service.CreateCollateralRequest(context.Background(), CreateRequest{
		UserID:       uuid.New(),
		LoanAmount:   2000,
		FiatCurrency: "USD",
		AssetSymbol:  "BTC",
	})
	require.NoError(t, err)
	require.Equal(t, models.StatusPending, collateral.Status)
	require.Equal(t, 1, len(repo.created))
}

func TestLockAndReleaseFlow(t *testing.T) {
	service, _ := newTestService()
	userID := uuid.New()

	collateral, err := service.LockCollateral(context.Background(), userID, LockRequest{
		AssetSymbol:   "BTC",
		TxHash:        "0xabc",
		Amount:        0.5,
		WalletAddress: "addr",
		FiatCurrency:  "USD",
	})
	require.NoError(t, err)
	require.Equal(t, models.StatusActive, collateral.Status)

	updated, err := service.RequestRelease(context.Background(), userID, collateral.ID)
	require.NoError(t, err)
	require.Equal(t, models.StatusReleaseRequested, updated.Status)

	updated, err = service.ApproveRelease(context.Background(), collateral.ID)
	require.NoError(t, err)
	require.Equal(t, models.StatusReleased, updated.Status)
}

func newTestService() (*Service, *mockRepo) {
	repo := newMockRepo()
	pricingProvider := &fakePricing{
		prices: map[string]float64{
			"BTC":  20000,
			"ETH":  1000,
			"USDT": 1,
		},
	}
	verifier := &fakeVerifier{}
	cfg := &config.Config{
		Loan: config.LoanConfig{
			DefaultLTV: 0.5,
		},
	}

	service := NewService(repo, pricingProvider, verifier, nil, cfg, zerolog.Nop())
	return service, repo
}

type mockRepo struct {
	store   map[uuid.UUID]*models.Collateral
	created []*models.Collateral
}

func newMockRepo() *mockRepo {
	return &mockRepo{
		store: make(map[uuid.UUID]*models.Collateral),
	}
}

func (m *mockRepo) Create(ctx context.Context, collateral *models.Collateral) error {
	if collateral.ID == uuid.Nil {
		collateral.ID = uuid.New()
	}
	copy := *collateral
	m.store[collateral.ID] = &copy
	m.created = append(m.created, &copy)
	return nil
}

func (m *mockRepo) GetByID(ctx context.Context, id uuid.UUID) (*models.Collateral, error) {
	if col, ok := m.store[id]; ok {
		copy := *col
		return &copy, nil
	}
	return nil, nil
}

func (m *mockRepo) GetByUserID(ctx context.Context, userID uuid.UUID) ([]models.Collateral, error) {
	var result []models.Collateral
	for _, col := range m.store {
		if col.UserID == userID {
			result = append(result, *col)
		}
	}
	return result, nil
}

func (m *mockRepo) ListAll(ctx context.Context) ([]models.Collateral, error) {
	var result []models.Collateral
	for _, col := range m.store {
		result = append(result, *col)
	}
	return result, nil
}

func (m *mockRepo) Update(ctx context.Context, collateral *models.Collateral) error {
	copy := *collateral
	m.store[collateral.ID] = &copy
	return nil
}

func (m *mockRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status models.CollateralStatus) error {
	if col, ok := m.store[id]; ok {
		col.Status = status
		return nil
	}
	return nil
}

func (m *mockRepo) UpdateTxInfo(ctx context.Context, id uuid.UUID, txHash, walletAddress string, status models.CollateralStatus) error {
	if col, ok := m.store[id]; ok {
		col.TxHash = &txHash
		col.WalletAddress = &walletAddress
		col.Status = status
		return nil
	}
	return nil
}

type fakePricing struct {
	prices map[string]float64
}

func (f *fakePricing) GetPrice(symbol, currency string) (float64, error) {
	if price, ok := f.prices[symbol]; ok {
		return price, nil
	}
	return 0, fmt.Errorf("price not found")
}

func (f *fakePricing) GetPrices(symbols []string, currency string) (map[string]float64, error) {
	result := make(map[string]float64)
	for _, symbol := range symbols {
		if price, ok := f.prices[symbol]; ok {
			result[symbol] = price
		}
	}
	return result, nil
}

type fakeVerifier struct{}

func (f *fakeVerifier) VerifyTransaction(ctx context.Context, txHash, assetSymbol string, expectedAmount float64) (bool, *blockchain.TransactionData, error) {
	return true, &blockchain.TransactionData{
		Hash:   txHash,
		Amount: expectedAmount,
	}, nil
}
