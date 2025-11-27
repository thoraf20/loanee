package wallet

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/thoraf20/loanee/internal/utils"
)

type Handler struct {
	service *Service
	logger  zerolog.Logger
}

func NewHandler(service *Service, logger zerolog.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger.With().Str("component", "wallet_handler").Logger(),
	}
}

func (h *Handler) ListMine(c *gin.Context) {
	userID, ok := utils.UserIDFromGin(c)
	if !ok {
		utils.Unauthorized(c, "authentication required")
		return
	}

	wallets, err := h.service.ListUserWallets(c.Request.Context(), userID)
	if err != nil {
		h.logger.Error().Err(err).Any("user_id", userID).Msg("failed to list wallets")
		utils.InternalServerError(c, "failed to fetch wallets", err.Error())
		return
	}

	utils.Success(c, http.StatusOK, "wallets retrieved", wallets)
}
