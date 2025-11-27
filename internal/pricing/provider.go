package pricing

type Provider interface {
	GetPrice(symbol, currency string) (float64, error)
	GetPrices(symbols []string, currency string) (map[string]float64, error)
}