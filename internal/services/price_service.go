package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"go.uber.org/zap"
	"golang.org/x/time/rate"
)

type CoinGeckoProvider struct {
	client *http.Client
	logger *zap.Logger
	apiKey string // Optional: for higher rate limits
	limiter *rate.Limiter // 10 requests per minute
}

// CoinGecko ID mapping (symbol -> coingecko_id)
var coinIDMap = map[string]string{
	"BTC":  "bitcoin",
	"ETH":  "ethereum",
	"USDT": "tether",
	"USDC": "usd-coin",
	"BNB":  "binancecoin",
	"XRP":  "ripple",
	"ADA":  "cardano",
	"DOGE": "dogecoin",
	"SOL":  "solana",
	"TRX":  "tron",
	"MATIC": "matic-network",
	"DOT":  "polkadot",
	"SHIB": "shiba-inu",
	"AVAX": "avalanche-2",
	"LINK": "chainlink",
	// Add more as needed
}

type CoinGeckoPriceResponse map[string]map[string]float64

// NewCoinGeckoProvider creates a new CoinGecko price provider
func NewCoinGeckoProvider(logger *zap.Logger, apiKey string) *CoinGeckoProvider {
	return &CoinGeckoProvider{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		logger: logger,
		apiKey: apiKey,
		limiter: rate.NewLimiter(rate.Every(6*time.Second), 1),
	}
}

func (p *CoinGeckoProvider) GetPrices(symbols []string, fiatCurrency string) (map[string]float64, error) {
	// Wait for rate limiter
	if err := p.limiter.Wait(context.Background()); err != nil {
		return nil, fmt.Errorf("rate limiter error: %w", err)
	}

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
			p.logger.Warn("Unknown cryptocurrency symbol",
					zap.String("symbol", symbol),
			)
		}
	}

	if len(coinIDs) == 0 {
		return nil, fmt.Errorf("no valid coin IDs found for symbols: %v", symbols)
	}

	// Normalize currency to lowercase
	fiatCurrency = strings.ToLower(strings.TrimSpace(fiatCurrency))
	if fiatCurrency == "" {
		fiatCurrency = "usd"
	}

	// Build URL with proper encoding
	ids := strings.Join(coinIDs, ",")
	baseURL := "https://api.coingecko.com/api/v3/simple/price"
	
	params := url.Values{}
	params.Add("ids", ids)
	params.Add("vs_currencies", fiatCurrency)
	
	fullURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())

	p.logger.Debug("Fetching prices from CoinGecko",
		zap.String("url", fullURL),
		zap.Strings("coin_ids", coinIDs),
		zap.String("currency", fiatCurrency),
	)

	// Create request
	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add API key if available (for pro plan)
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
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Handle non-200 responses
	if resp.StatusCode != http.StatusOK {
		p.logger.Error("CoinGecko API error",
			zap.Int("status_code", resp.StatusCode),
			zap.String("status", resp.Status),
			zap.String("response_body", string(body)),
			zap.String("url", fullURL),
		)
		
		return nil, fmt.Errorf("failed to fetch prices, status: %s, body: %s", 
			resp.Status, string(body))
	}

	// Parse response
	var cgResponse CoinGeckoPriceResponse
	if err := json.Unmarshal(body, &cgResponse); err != nil {
		p.logger.Error("Failed to parse CoinGecko response",
			zap.Error(err),
			zap.String("body", string(body)),
		)
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Convert back to symbol-based map
	prices := make(map[string]float64)
	for coinID, priceData := range cgResponse {
		if symbol, ok := symbolToID[coinID]; ok {
			if price, ok := priceData[fiatCurrency]; ok {
					prices[symbol] = price
			}
		}
	}

	p.logger.Info("Successfully fetched prices",
		zap.Int("count", len(prices)),
		zap.Any("prices", prices),
	)

  return prices, nil
}

// GetPrice fetches price for a single asset
func (p *CoinGeckoProvider) GetPrice(symbol, fiatCurrency string) (float64, error) {
	prices, err := p.GetPrices([]string{symbol}, fiatCurrency)
	if err != nil {
		return 0, err
	}

	price, ok := prices[strings.ToUpper(symbol)]
	if !ok {
		return 0, fmt.Errorf("price not found for %s", symbol)
	}

	return price, nil
}

// GetSupportedCoins fetches list of all coins from CoinGecko
func (p *CoinGeckoProvider) GetSupportedCoins() (map[string]string, error) {
  url := "https://api.coingecko.com/api/v3/coins/list"
    
	resp, err := p.client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var coins []struct {
		ID     string `json:"id"`
		Symbol string `json:"symbol"`
		Name   string `json:"name"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&coins); err != nil {
		return nil, err
	}

	// Build symbol -> id map
	coinMap := make(map[string]string)
	for _, coin := range coins {
		symbol := strings.ToUpper(coin.Symbol)
		coinMap[symbol] = coin.ID
	}

  return coinMap, nil
}