package main

import (
	"fmt"
	"log"
	"net/http"

	router "github.com/thoraf20/loanee/internal/http"
)

func main() {
	router := router.NewRouter()

	port := ":8080"
	fmt.Printf("Server running on http://localhost%s\n", port)

	if err := http.ListenAndServe(port, router); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
