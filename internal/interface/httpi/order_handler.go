package httpi

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/kazshi01/payment-system/internal/domain/order"
	"github.com/kazshi01/payment-system/internal/usecase"
)

type orderJSON struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	AmountJPY int64     `json:"amount_jpy"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type OrderHandler struct {
	UC *usecase.OrderUsecase
}

// POST /orders
func (h *OrderHandler) Create(w http.ResponseWriter, r *http.Request) {
	// 事故防止のボディ上限
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // 1MB
	defer r.Body.Close()

	var body struct {
		AmountJPY int64 `json:"amount_jpy"`
	}

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields() // 未知のフィールドを禁止
	if err := dec.Decode(&body); err != nil {
		if errors.Is(err, io.EOF) {
			http.Error(w, "empty body", http.StatusBadRequest)
			return
		}
		http.Error(w, "invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}
	if dec.More() {
		http.Error(w, "unexpected extra JSON", http.StatusBadRequest)
		return
	}

	o, err := h.UC.CreateOrder(r.Context(), body.AmountJPY)

	if err != nil {
		WriteError(w, err)
		return
	}

	// JSON形式で返却するための処理
	resp := orderJSON{
		ID:        string(o.ID),
		UserID:    o.UserID,
		AmountJPY: o.AmountJPY,
		Status:    string(o.Status),
		CreatedAt: o.CreatedAt,
		UpdatedAt: o.UpdatedAt,
	}

	b, err := json.Marshal(resp)
	if err != nil {
		log.Printf("CreateOrder success (marshal error): %+v", resp)
	} else {
		log.Printf("CreateOrder success: %s", b)
	}

	w.Header().Set("Location", "/orders/"+string(o.ID))
	WriteJSON(w, http.StatusCreated, resp)
}

// POST /orders/{id}/pay
func (h *OrderHandler) Pay(w http.ResponseWriter, r *http.Request) {
	id := order.ID(r.PathValue("id"))
	if id == "" {
		http.Error(w, "missing id", http.StatusBadRequest)
		return
	}

	if err := h.UC.PayOrder(r.Context(), id); err != nil {
		WriteError(w, err)
		return
	}

	log.Printf("PayOrder success: order_id=%s", id)

	w.WriteHeader(http.StatusNoContent)
}
