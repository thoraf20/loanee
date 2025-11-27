package collateral

import "github.com/google/uuid"

// Supported fiat and asset codes. Keeping it small for now – extend as needed.
var SupportedAssets = []string{"BTC", "ETH", "USDT"}

type PreviewQuery struct {
	LoanAmount   float64 `form:"loan_amount" binding:"required,gt=0"`
	FiatCurrency string  `form:"fiat" binding:"required,oneof=USD NGN"`
}

type PreviewItem struct {
	AssetSymbol    string  `json:"asset"`
	FiatCurrency   string  `json:"fiat_currency"`
	LoanAmount     float64 `json:"loan_amount"`
	CollateralLTV  float64 `json:"ltv"`
	AssetPrice     float64 `json:"asset_price"`
	RequiredValue  float64 `json:"required_value"`
	RequiredAmount float64 `json:"required_amount"`
	Status         string  `json:"status"`
}

type PreviewResponse struct {
	FiatCurrency string        `json:"fiat_currency"`
	LoanAmount   float64       `json:"loan_amount"`
	Previews     []PreviewItem `json:"previews"`
}

type CreateRequest struct {
	LoanAmount   float64   `json:"loan_amount" validate:"required,gt=0"`
	FiatCurrency string    `json:"fiat_currency" validate:"required,oneof=USD NGN"`
	AssetSymbol  string    `json:"asset_symbol" validate:"required,oneof=BTC ETH USDT"`
	UserID       uuid.UUID `json:"-"`
}

type LockRequest struct {
	AssetSymbol   string  `json:"asset_symbol" validate:"required,oneof=BTC ETH USDT"`
	TxHash        string  `json:"tx_hash" validate:"required"`
	Amount        float64 `json:"amount" validate:"required,gt=0"`
	WalletAddress string  `json:"wallet_address"`
	FiatCurrency  string  `json:"fiat_currency" validate:"required,oneof=USD NGN"`
}

type VerifyRequest struct {
	CollateralID    uuid.UUID `json:"collateral_id" validate:"required"`
	TransactionHash string    `json:"transaction_hash" validate:"required"`
}
