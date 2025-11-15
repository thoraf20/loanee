package services_test

import "go.uber.org/zap"

func TestCoinGeckoAPI() {
	logger, _ := zap.NewDevelopment()
	provider := NewCoinGeckoProvider(logger, "")

	// Test single price
	price, err := provider.GetPrice("BTC", "usd")
	if err != nil {
		logger.Fatal("Failed to get BTC price", zap.Error(err))
	}
	logger.Info("BTC Price", zap.Float64("usd", price))

	// Test multiple prices
	prices, err := provider.GetPrices([]string{"BTC", "ETH", "USDT"}, "usd")
	if err != nil {
		logger.Fatal("Failed to get prices", zap.Error(err))
	}
	logger.Info("Prices", zap.Any("prices", prices))
}