package pricing

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

type CoinGeckoProvider struct {
	apiKey  string
	baseURL string
	client  *http.Client
	logger  zerolog.Logger
}

// Coin symbol to CoinGecko ID mapping
var coinIDMap = map[string]string{
	"BTC":  "bitcoin",
	"ETH":  "ethereum",
	"BNB":  "binancecoin",
	"USDT": "tether",
	"USDC": "usd-coin",
	"XRP":  "ripple",
	"ADA":  "cardano",
	"DOGE": "dogecoin",
	"SOL":  "solana",
	"TRX":  "tron",
	"MATIC": "matic-network",
	"DOT":  "polkadot",
	"AVAX": "avalanche-2",
	"LINK": "chainlink",
}

type CoinGeckoPriceResponse map[string]map[string]float64

func NewCoinGeckoProvider(apiKey string, logger zerolog.Logger) Provider {
	return &CoinGeckoProvider{
		apiKey:  apiKey,
		baseURL: "https://api.coingecko.com/api/v3",
		client: &http.Client{
				Timeout: 10 * time.Second,
		},
		logger: logger,
	}
}

func (p *CoinGeckoProvider) GetPrice(symbol, currency string) (float64, error) {
	prices, err := p.GetPrices([]string{symbol}, currency)
	if err != nil {
		return 0, err
	}

	price, ok := prices[strings.ToUpper(symbol)]
	if !ok {
		return 0, fmt.Errorf("price not found for %s", symbol)
	}

	return price, nil
}

func (p *CoinGeckoProvider) GetPrices(symbols []string, currency string) (map[string]float64, error) {
	if len(symbols) == 0 {
		return nil, fmt.Errorf("no symbols provided")
	}

	// Convert symbols to CoinGecko IDs
	coinIDs := make([]string, 0, len(symbols))
	symbolToID := make(map[string]string)

	for _, symbol := range symbols {
		symbol = strings.ToUpper(strings.TrimSpace(symbol))
		if coinID, ok := coinIDMap[symbol]; ok {
				coinIDs = append(coinIDs, coinID)
				symbolToID[coinID] = symbol
		} else {
				p.logger.Warn().Str("symbol", symbol).Msg("Unknown cryptocurrency symbol")
		}
	}

	if len(coinIDs) == 0 {
		return nil, fmt.Errorf("no valid coin IDs found")
	}

	// Normalize currency
	currency = strings.ToLower(strings.TrimSpace(currency))
	if currency == "" {
		currency = "usd"
	}

	// Build URL
	ids := strings.Join(coinIDs, ",")
	params := url.Values{}
	params.Add("ids", ids)
	params.Add("vs_currencies", currency)

	fullURL := fmt.Sprintf("%s/simple/price?%s", p.baseURL, params.Encode())

	p.logger.Debug().
		Str("url", fullURL).
		Strs("coin_ids", coinIDs).
		Str("currency", currency).
		Msg("Fetching prices from CoinGecko")

	// Create request
	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add API key if available
	if p.apiKey != "" {
		req.Header.Set("x-cg-pro-api-key", p.apiKey)
	}

	// Make request
	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch prices: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Handle non-200 responses
	if resp.StatusCode != http.StatusOK {
		p.logger.Error().
			Int("status", resp.StatusCode).
			Str("body", string(body)).
			Msg("CoinGecko API error")
		return nil, fmt.Errorf("API error: status %d, body: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var cgResponse CoinGeckoPriceResponse
	if err := json.Unmarshal(body, &cgResponse); err != nil {
		p.logger.Error().
			Err(err).
			Str("body", string(body)).
			Msg("Failed to parse CoinGecko response")
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Convert to symbol-based map
	prices := make(map[string]float64)
	for coinID, priceData := range cgResponse {
		if symbol, ok := symbolToID[coinID]; ok {
			if price, ok := priceData[currency]; ok {
				prices[symbol] = price
			}
		}
	}

	p.logger.Info().
		Int("count", len(prices)).
		Interface("prices", prices).
		Msg("Successfully fetched prices")

  return prices, nil
}