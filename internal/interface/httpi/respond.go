package httpi

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/kazshi01/payment-system/internal/domain"
)

func WriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func WriteError(w http.ResponseWriter, err error) {
	code := http.StatusInternalServerError
	switch {
	case errors.Is(err, domain.ErrInvalidArgument):
		code = http.StatusBadRequest
	case errors.Is(err, domain.ErrNotFound):
		code = http.StatusNotFound
	case errors.Is(err, domain.ErrConflict):
		code = http.StatusConflict
	}
	http.Error(w, err.Error(), code)
}
