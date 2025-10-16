package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
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
	vars := mux.Vars(r)
	userID, err := uuid.Parse(vars["user_id"])
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "invalid user id", err.Error())
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	wallets, err := h.walletService.GetUserWallets(ctx, userID)
	if err != nil {
		utils.Error(w, http.StatusNotFound, "failed to fetch wallets", err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	utils.JSON(w, http.StatusOK, "Wallets fetched successfully", wallets)
}