package payment

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
		logger:  logger.With().Str("component", "payment_handler").Logger(),
	}
}

func (h *Handler) RepayLoan(c *gin.Context) {
	userID, ok := utils.UserIDFromGin(c)
	if !ok {
		utils.Unauthorized(c, "authentication required")
		return
	}

	loanID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "invalid loan id", err.Error())
		return
	}

	var req RepaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "invalid payload", err.Error())
		return
	}

	result, err := h.service.RecordRepayment(c.Request.Context(), userID, loanID, req)
	if err != nil {
		h.logger.Error().Err(err).Any("loan_id", loanID).Msg("failed to record repayment")
		utils.InternalServerError(c, "failed to record repayment", err.Error())
		return
	}

	utils.Success(c, http.StatusOK, "repayment recorded", result)
}

func (h *Handler) ListRepayments(c *gin.Context) {
	userID, ok := utils.UserIDFromGin(c)
	if !ok {
		utils.Unauthorized(c, "authentication required")
		return
	}

	loanID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "invalid loan id", err.Error())
		return
	}

	// Loan ownership is validated at service layer when fetching payments by loan ID via loan service.
	payments, err := h.service.ListRepayments(c.Request.Context(), loanID, userID)
	if err != nil {
		h.logger.Error().Err(err).Any("loan_id", loanID).Msg("failed to list repayments")
		utils.InternalServerError(c, "failed to list repayments", err.Error())
		return
	}

	utils.Success(c, http.StatusOK, "repayments retrieved", payments)
}
