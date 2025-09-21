package httpi

import (
	"fmt"
	"net/http"
)

func Home(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")

	at, err := r.Cookie("at")
	if err != nil || at.Value == "" {
		http.Redirect(w, r, "/auth/logout", http.StatusSeeOther)
		return
	}

	fmt.Fprintf(w, "Logged in!")
}
