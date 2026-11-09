// Package httpjson writes JSON responses and maps request-body errors.
package httpjson

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/y0f/gophling/v2/internal/platform/problem"
)

// Write marshals rather than streams so the bytes are exactly what an
// idempotent replay stores and serves back: json.Encoder would append a
// newline the stored copy does not have.
func Write(w http.ResponseWriter, status int, v any) {
	body, err := json.Marshal(v)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write(body)
}

// BodyError maps a body-read error to 413 when the configured size cap was
// exceeded, and to 400 otherwise.
func BodyError(w http.ResponseWriter, r *http.Request, err error, badRequestMsg string) {
	var tooLarge *http.MaxBytesError
	if errors.As(err, &tooLarge) {
		problem.Status(w, r, http.StatusRequestEntityTooLarge, "request body exceeds the size limit")
		return
	}
	problem.Status(w, r, http.StatusBadRequest, badRequestMsg)
}
