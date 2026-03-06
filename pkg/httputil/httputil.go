// Package httputil provides lightweight helpers for JSON HTTP handlers.
package httputil

import (
	"encoding/json"
	"net/http"
)

// Decode reads a JSON body into dst. Returns a non-nil error on bad input.
func Decode(r *http.Request, dst any) error {
	return json.NewDecoder(r.Body).Decode(dst)
}

// JSON writes v as a JSON response with the given status code.
func JSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

// Error writes a plain-text error response.
func Error(w http.ResponseWriter, status int, msg string) {
	http.Error(w, msg, status)
}
