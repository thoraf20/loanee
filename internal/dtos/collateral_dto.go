package dtos

type AddCollateralDTO struct {
	Asset    string `json:"asset" validate:"required,asset"`
	Amount     float64 `json:"amount" validate:"required,amount"`
	FiatValue float64 `json:"fiat_value" validate:"required,fiat_value"`
}


type CollateralPreviewItem struct {
	Asset          string  `json:"asset"`
	FiatCurrency   string  `json:"fiat_currency"`
	LoanAmount     float64 `json:"loan_amount"`
	LTV            float64 `json:"ltv"`
	AssetPrice     float64 `json:"asset_price"`
	RequiredValue  float64 `json:"required_value"`
	RequiredAmount float64 `json:"required_amount"`
	Status         string  `json:"status"`
}

type CollateralPreviewResponse struct {
	FiatCurrency string                       `json:"fiat_currency"`
	LoanAmount   float64                      `json:"loan_amount"`
	Previews     []CollateralPreviewItem  `json:"previews"`
}
