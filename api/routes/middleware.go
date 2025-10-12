package routes

import (
	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
	middleware "github.com/thoraf20/loanee/api/middlewares"
)

// ApplyAuthMiddleware applies the authentication middleware to a router
func ApplyAuthMiddleware(router *mux.Router, logger *zerolog.Logger) *mux.Router {
	protectedRouter := router.NewRoute().Subrouter()
	protectedRouter.Use(middleware.AuthMiddleware)
	return protectedRouter
}
