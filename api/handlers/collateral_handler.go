package handlers

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/thoraf20/loanee/internal/dtos"
	"github.com/thoraf20/loanee/internal/services"
	"github.com/thoraf20/loanee/internal/utils"
	"github.com/thoraf20/loanee/pkg/binding"
)

type CollateralHandler struct {
	CollateralService services.CollateralService
}

func NewCollateralHandler(service services.CollateralService) *CollateralHandler {
	return &CollateralHandler{CollateralService: service}
}

type PreviewRequest struct {
	LoanAmount   float64 `json:"loan_amount"`
	FiatCurrency string  `json:"fiat_currency"`
}

func (h *CollateralHandler) PreviewCollateral(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	loanAmountStr := query.Get("loan_amount")
	fiatCurrency := query.Get("fiat")

	if loanAmountStr == "" || fiatCurrency == "" {
		utils.Error(w, http.StatusBadRequest, "Missing required query params", "loan_amount and fiat are required")
		return
	}

	loanAmount, err := strconv.ParseFloat(loanAmountStr, 64)
	if err != nil || loanAmount <= 0 {
		utils.Error(w, http.StatusBadRequest, "Invalid loan amount", "loan_amount must be a positive number")
		return
	}

	result, err := h.CollateralService.PreviewCollateral(loanAmount, fiatCurrency)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to preview collateral", err.Error())
		return
	}

	utils.JSON(w, http.StatusOK, "Collateral preview successful", result)
}

func (h *CollateralHandler) LockCollateral(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	userID, err := utils.GetUserIDFromContext(ctx)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized", err.Error())
		return
	}

	var dto dtos.CollateralLockDTO
	_, verr := binding.StrictBindJSON[dtos.CollateralLockDTO](r)
	if verr != nil {
		utils.Error(w, http.StatusBadRequest, verr.Message, verr)
		return
	}

	collateral, err := h.CollateralService.LockCollateral(ctx, userID, dto)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to lock collateral", err.Error())
		return
	}

	utils.JSON(w, http.StatusOK, "Collateral locked successfully", collateral)
}