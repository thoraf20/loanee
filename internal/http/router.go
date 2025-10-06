package http

import (
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

func NewRouter() *mux.Router {
	r := mux.NewRouter()

	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		JSON(w, http.StatusOK, "Service is healthy", map[string]string{
			"uptime": "ok",
			"timestamptz": time.Now().Local().Format(time.RFC3339),
		})
	}).Methods("GET")

	return r
}