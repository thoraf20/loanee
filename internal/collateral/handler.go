package collateral

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/thoraf20/loanee/internal/utils"
	"github.com/thoraf20/loanee/pkg/validator"
)

type Handler struct {
	service   *Service
	validator *validator.Validator
	logger    zerolog.Logger
}

func NewHandler(service *Service, validator *validator.Validator, logger zerolog.Logger) *Handler {
	return &Handler{
		service:   service,
		validator: validator,
		// Use a component-specific logger to make filtering easier.
		logger: logger.With().Str("component", "collateral_handler").Logger(),
	}
}

func (h *Handler) Preview(c *gin.Context) {
	var query PreviewQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		utils.BadRequest(c, "invalid query parameters", err.Error())
		return
	}

	result, err := h.service.PreviewCollateral(c.Request.Context(), query.LoanAmount, query.FiatCurrency)
	if err != nil {
		h.logger.Error().Err(err).Msg("failed to preview collateral")
		utils.InternalServerError(c, "failed to preview collateral", err.Error())
		return
	}

	utils.OK(c, "collateral preview generated", result)
}

func (h *Handler) CreateRequest(c *gin.Context) {
	userID, ok := utils.UserIDFromGin(c)
	if !ok {
		utils.Unauthorized(c, "authentication required")
		return
	}

	var payload CreateRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		utils.BadRequest(c, "invalid payload", err.Error())
		return
	}

	payload.UserID = userID
	if err := h.validator.Validate(&payload); err != nil {
		utils.BadRequest(c, "validation failed", err.Error())
		return
	}

	result, err := h.service.CreateCollateralRequest(c.Request.Context(), payload)
	if err != nil {
		h.logger.Error().Err(err).Msg("failed to create collateral request")
		utils.InternalServerError(c, "failed to create collateral request", err.Error())
		return
	}

	utils.Created(c, "collateral request created", result)
}

func (h *Handler) Lock(c *gin.Context) {
	userID, ok := utils.UserIDFromGin(c)
	if !ok {
		utils.Unauthorized(c, "authentication required")
		return
	}

	var payload LockRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		utils.BadRequest(c, "invalid payload", err.Error())
		return
	}

	if err := h.validator.Validate(&payload); err != nil {
		utils.BadRequest(c, "validation failed", err.Error())
		return
	}

	collateral, err := h.service.LockCollateral(c.Request.Context(), userID, payload)
	if err != nil {
		h.logger.Error().Err(err).Msg("failed to lock collateral")
		utils.InternalServerError(c, "failed to lock collateral", err.Error())
		return
	}

	utils.OK(c, "collateral locked", collateral)
}

func (h *Handler) ListMine(c *gin.Context) {
	userID, ok := utils.UserIDFromGin(c)
	if !ok {
		utils.Unauthorized(c, "authentication required")
		return
	}

	collaterals, err := h.service.ListUserCollaterals(c.Request.Context(), userID)
	if err != nil {
		h.logger.Error().Err(err).Any("user_id", userID).Msg("failed to list collaterals")
		utils.InternalServerError(c, "failed to fetch collaterals", err.Error())
		return
	}

	utils.Success(c, http.StatusOK, "collaterals retrieved", collaterals)
}

func (h *Handler) RequestRelease(c *gin.Context) {
	userID, ok := utils.UserIDFromGin(c)
	if !ok {
		utils.Unauthorized(c, "authentication required")
		return
	}

	collateralID, ok := parseUUIDParam(c, "id")
	if !ok {
		utils.BadRequest(c, "invalid collateral id", nil)
		return
	}

	collateral, err := h.service.RequestRelease(c.Request.Context(), userID, collateralID)
	if err != nil {
		h.logger.Error().Err(err).Any("user_id", userID).Any("collateral_id", collateralID).Msg("failed to request release")
		utils.InternalServerError(c, "failed to request release", err.Error())
		return
	}

	utils.OK(c, "collateral release requested", collateral)
}

func (h *Handler) AdminList(c *gin.Context) {
	collaterals, err := h.service.ListAllCollaterals(c.Request.Context())
	if err != nil {
		h.logger.Error().Err(err).Msg("failed to list collaterals")
		utils.InternalServerError(c, "failed to fetch collaterals", err.Error())
		return
	}

	utils.Success(c, http.StatusOK, "collaterals retrieved", collaterals)
}

func (h *Handler) AdminApproveRelease(c *gin.Context) {
	collateralID, ok := parseUUIDParam(c, "id")
	if !ok {
		utils.BadRequest(c, "invalid collateral id", nil)
		return
	}

	collateral, err := h.service.ApproveRelease(c.Request.Context(), collateralID)
	if err != nil {
		h.logger.Error().Err(err).Any("collateral_id", collateralID).Msg("failed to approve release")
		utils.InternalServerError(c, "failed to approve release", err.Error())
		return
	}

	utils.OK(c, "collateral release approved", collateral)
}

type adminDecisionDTO struct {
	Reason string `json:"reason"`
}

func (h *Handler) AdminRejectRelease(c *gin.Context) {
	collateralID, ok := parseUUIDParam(c, "id")
	if !ok {
		utils.BadRequest(c, "invalid collateral id", nil)
		return
	}

	var dto adminDecisionDTO
	if c.Request.ContentLength > 0 {
		if err := c.ShouldBindJSON(&dto); err != nil {
			utils.BadRequest(c, "invalid payload", err.Error())
			return
		}
	}

	collateral, err := h.service.RejectRelease(c.Request.Context(), collateralID, dto.Reason)
	if err != nil {
		h.logger.Error().Err(err).Any("collateral_id", collateralID).Msg("failed to reject release")
		utils.InternalServerError(c, "failed to reject release", err.Error())
		return
	}

	utils.OK(c, "collateral release rejected", collateral)
}

func parseUUIDParam(c *gin.Context, param string) (uuid.UUID, bool) {
	value := c.Param(param)
	id, err := uuid.Parse(value)
	if err != nil {
		return uuid.Nil, false
	}
	return id, true
}
