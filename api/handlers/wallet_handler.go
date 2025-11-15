package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/thoraf20/loanee/internal/services"
	"github.com/thoraf20/loanee/internal/utils"
)

type WalletHandler struct {
	walletService services.WalletService
}

func NewWalletHandler(walletService services.WalletService) *WalletHandler {
	return &WalletHandler{walletService: walletService}
}

func (h *WalletHandler) GetUserWallets(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	userID, err := utils.GetUserIDFromContext(ctx)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "invalid user id", err.Error())
		return
	}

	wallets, err := h.walletService.GetUserWallets(ctx, userID)
	if err != nil {
		utils.Error(w, http.StatusNotFound, "failed to fetch wallets", err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	utils.JSON(w, http.StatusOK, "Wallets fetched successfully", wallets)
}