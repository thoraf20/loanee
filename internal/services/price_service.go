package services

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type PriceService interface {
	GetPrice(assetSymbol, fiatCurrency string) (float64, error)
	GetPrices(fiatCurrency string) (map[string]float64, error)
}

type priceService struct {
	client *http.Client
}

func NewPriceService() PriceService {
	return &priceService{
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

// Map our asset symbols to CoinGecko IDs
var assetMap = map[string]string{
	"BTC":  "bitcoin",
	"ETH":  "ethereum",
	"USDT": "tether",
}

// We need the current price of an asset in the specified fiat
func (p *priceService) GetPrice(assetSymbol, fiatCurrency string) (float64, error) {
	prices, err := p.GetPrices(fiatCurrency)
	if err != nil {
		return 0, err
	}
	price, ok := prices[assetSymbol]
	if !ok {
		return 0, fmt.Errorf("price not found for %s", assetSymbol)
	}
	return price, nil
}

func (p *priceService) GetPrices(fiatCurrency string) (map[string]float64, error) {
	assetIDs := make([]string, 0, len(assetMap))
	for _, id := range assetMap {
		assetIDs = append(assetIDs, id)
	}

	ids := strings.Join(assetIDs, ",")
	url := fmt.Sprintf(
		"https://api.coingecko.com/api/v3/simple/price?ids=%s&vs_currencies=%s",
		ids,
		fiatCurrency,
	)

	resp, err := p.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch prices: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch prices, status: %s", resp.Status)
	}

	var data map[string]map[string]float64
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("invalid response format: %w", err)
	}

	result := make(map[string]float64)
	for symbol, id := range assetMap {
		// CoinGecko requires lowercase fiat codes
		if val, ok := data[id][strings.ToLower(fiatCurrency)]; ok {
			result[symbol] = val
		}
	}

	if len(result) == 0 {
		return nil, fmt.Errorf("no prices found for fiat currency: %s", fiatCurrency)
	}

	return result, nil
}