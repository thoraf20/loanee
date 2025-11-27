package user

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
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
		logger:    logger,
	}
}

// GetProfile godoc
// @Summary Get user profile
// @Description Get current user's profile
// @Tags users
// @Accept json
// @Produce json
// @Security Bearer
// @Success 200 {object} User
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /users/me [get]
func (h *Handler) GetProfile(c *gin.Context) {
	userID, ok := utils.UserIDFromGin(c)
	if !ok {
		utils.Unauthorized(c, "authentication required")
		return
	}

	user, err := h.service.GetByID(c.Request.Context(), userID)
	if err != nil {
		h.logger.Error().Err(err).Any("user_id", userID).Msg("Failed to get user profile")
		utils.InternalServerError(c, "Failed to get profile", err.Error())
		return
	}

	c.JSON(http.StatusOK, user)
}

// UpdateProfile godoc
// @Summary Update user profile
// @Description Update current user's profile
// @Tags users
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body UpdateProfileRequest true "Update profile request"
// @Success 200 {object} User
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /users/me [put]
// func (h *Handler) UpdateProfile(c *gin.Context,) {
// 	var req UpdateProfileRequestDTO
// 	if err := c.ShouldBindJSON(&req); err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
// 		return
// 	}

// 	if err := h.validator.Validate(req); err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
// 		return
// 	}

// 	userID := c.GetUint("userID")
// 	user, err := h.service.GetByID(c.Request.Context(), userID)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user"})
// 		return
// 	}

// 	// Update fields
// 	if req.FirstName != "" {
// 		user.FirstName = req.FirstName
// 	}
// 	if req.LastName != "" {
// 		user.LastName = req.LastName
// 	}

// 	if err := h.service.Update(c.Request.Context(), user); err != nil {
// 		h.logger.Error().Err(err).Uint("userID", userID).Msg("Failed to update user")
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile"})
// 		return
// 	}

// 	c.JSON(http.StatusOK, user)
// }

// ListUsers godoc
// @Summary List all users (Admin)
// @Description Get list of all users
// @Tags admin
// @Accept json
// @Produce json
// @Security Bearer
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {object} UserListResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /admin/users [get]
func (h *Handler) ListUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	offset := (page - 1) * limit

	users, total, err := h.service.List(c.Request.Context(), limit, offset)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to list users")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list users"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"users": users,
		"pagination": gin.H{
			"total": total,
			"page":  page,
			"limit": limit,
		},
	})
}
