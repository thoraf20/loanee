package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/thoraf20/loanee/internal/services"
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
		http.Error(w, "invalid user id", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	wallets, err := h.walletService.GetUserWallets(ctx, userID)
	if err != nil {
		http.Error(w, "failed to fetch wallets", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(wallets)
}