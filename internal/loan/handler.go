package loan

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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
		logger:  logger.With().Str("component", "loan_handler").Logger(),
	}
}

func (h *Handler) ListMine(c *gin.Context) {
	userID, ok := utils.UserIDFromGin(c)
	if !ok {
		utils.Unauthorized(c, "authentication required")
		return
	}

	loans, err := h.service.ListUserLoans(c.Request.Context(), userID)
	if err != nil {
		h.logger.Error().Err(err).Any("user_id", userID).Msg("failed to list loans")
		utils.InternalServerError(c, "failed to fetch loans", err.Error())
		return
	}

	utils.Success(c, http.StatusOK, "loans retrieved", loans)
}

func (h *Handler) AdminList(c *gin.Context) {
	loans, err := h.service.ListAllLoans(c.Request.Context())
	if err != nil {
		h.logger.Error().Err(err).Msg("failed to list loans")
		utils.InternalServerError(c, "failed to fetch loans", err.Error())
		return
	}

	utils.Success(c, http.StatusOK, "loans retrieved", loans)
}

type approveLoanDTO struct {
	Amount float64 `json:"amount"`
}

func (h *Handler) AdminApprove(c *gin.Context) {
	loanID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "invalid loan id", err.Error())
		return
	}

	var dto approveLoanDTO
	if c.Request.ContentLength > 0 {
		if err := c.ShouldBindJSON(&dto); err != nil {
			utils.BadRequest(c, "invalid payload", err.Error())
			return
		}
	}

	loan, err := h.service.ApproveLoan(c.Request.Context(), loanID, dto.Amount)
	if err != nil {
		h.logger.Error().Err(err).Any("loan_id", loanID).Msg("failed to approve loan")
		utils.InternalServerError(c, "failed to approve loan", err.Error())
		return
	}

	utils.Success(c, http.StatusOK, "loan approved", loan)
}

func (h *Handler) AdminDisburse(c *gin.Context) {
	loanID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "invalid loan id", err.Error())
		return
	}

	loan, err := h.service.DisburseLoan(c.Request.Context(), loanID)
	if err != nil {
		h.logger.Error().Err(err).Any("loan_id", loanID).Msg("failed to disburse loan")
		utils.InternalServerError(c, "failed to disburse loan", err.Error())
		return
	}

	utils.Success(c, http.StatusOK, "loan disbursed", loan)
}
