package handlers

import (
	"net/http"


	"github.com/thoraf20/loanee/internal/services"
	"github.com/thoraf20/loanee/internal/utils"
)

type UserHandler struct {
	UserService services.UserService
}

func NewUserHandler(userService services.UserService) *UserHandler {
	return &UserHandler{UserService: userService}
}

func (h *UserHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user, err := h.UserService.GetLoggedInUser(ctx)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized", err.Error())
		return
	}

	utils.JSON(w, http.StatusOK, "User profile fetched successfully", user)
}