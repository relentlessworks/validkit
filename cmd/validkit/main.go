package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/relentlessworks/validkit/internal/api"
	"github.com/relentlessworks/validkit/internal/config"
)

func main() {
	cfg := config.Load()

	handler := api.NewHandler(cfg.NoAuth)
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	log.Printf("validkit: listening on %s (auth=%t)", cfg.Addr, !cfg.NoAuth)
	if err := http.ListenAndServe(cfg.Addr, mux); err != nil {
		fmt.Fprintf(log.Writer(), "validkit: server error: %s\n", err)
	}
}
