package httpi

import (
	"fmt"
	"net/http"
)

func Home(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	fmt.Fprintln(w, "Logged in! You can now call the API. (This is /)")
}
