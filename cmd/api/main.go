package main

import (
	"log"
	"net/http"

	httpi "github.com/kazshi01/payment-system/internal/interface/http"
	"github.com/kazshi01/payment-system/internal/usecase"
	// wire: repo, tx, pg をDI
)

func main() {
	uc := &usecase.OrderUsecase{ /* Repo: ..., Tx: ..., PG: ... */ }
	h := &httpi.OrderHandler{UC: uc}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /orders", h.Create)
	mux.HandleFunc("POST /orders/{id}/pay", h.Pay)
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, _ *http.Request) { w.Write([]byte("ok")) })

	log.Fatal(http.ListenAndServe(":8080", mux))
}
