package httpi

import (
	"encoding/json"
	"net/http"

	"github.com/kazshi01/payment-system/internal/domain/order"
	"github.com/kazshi01/payment-system/internal/usecase"
)

type OrderHandler struct{ UC *usecase.OrderUsecase }

func (h *OrderHandler) Create(w http.ResponseWriter, r *http.Request) {
	var body struct {
		AmountJPY int64 `json:"amount_jpy"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	o, err := h.UC.CreateOrder(r.Context(), body.AmountJPY)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	_ = json.NewEncoder(w).Encode(o)
}

func (h *OrderHandler) Pay(w http.ResponseWriter, r *http.Request) {
	id := order.ID(r.PathValue("id"))
	if err := h.UC.PayOrder(r.Context(), id); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
