package dtos

import "github.com/google/uuid"

type AddCollateralDTO struct {
	Asset     string  `json:"asset" validate:"required,asset"`
	Amount    float64 `json:"amount" validate:"required,amount"`
	FiatValue float64 `json:"fiat_value" validate:"required,fiat_value"`
}

type CollateralPreviewDTO struct {
	LoanAmount   float64 `json:"loan_amount"`
	FiatCurrency string  `json:"fiat_currency"`
}

type CollateralLockDTO struct {
	// LoanRequestID string  `json:"loan_request_id" validate:"required,uuid4"`
	AssetSymbol   string  `json:"asset_symbol" validate:"required,oneof=BTC ETH USDT"`
	TxHash        string  `json:"tx_hash" validate:"required"`
	Amount        float64 `json:"amount" validate:"required,gt=0"`
	WalletAddress string  `json:"wallet_address"`
	FiatCurrency  string  `json:"fiat_currency" validate:"required,oneof=USD NGN"`
}

type CollateralPreviewItem struct {
	AssetSymbol    string  `json:"asset"`
	FiatCurrency   string  `json:"fiat_currency"`
	LoanAmount     float64 `json:"loan_amount"`
	CollateralLTV  float64 `json:"ltv"`
	AssetPrice     float64 `json:"asset_price"`
	RequiredValue  float64 `json:"required_value"`
	RequiredAmount float64 `json:"required_amount"`
	Status         string  `json:"status"`
}

type CollateralPreviewResponse struct {
	FiatCurrency string                  `json:"fiat_currency"`
	LoanAmount   float64                 `json:"loan_amount"`
	Previews     []CollateralPreviewItem `json:"previews"`
}

type CreateCollateralRequestDTO struct {
	UserID        uuid.UUID  `json:"user_id" validate:"uuid"`
	LoanRequestID string  `json:"loan_request_id" validate:"uuid"`
	LoanAmount    float64 `json:"loan_amount" validate:"required,gt=0"`
	FiatCurrency  string  `json:"fiat_currency" validate:"required,oneof=USD NGN"`
	AssetSymbol   string  `json:"asset_symbol" validate:"required,oneof=BTC ETH USDT"`
}

type VerifyCollateralRequestDTO struct {
	CollateralID    uuid.UUID `json:"collateralId" binding:"required"`
	TransactionHash string `json:"transactionHash" binding:"required"`
}