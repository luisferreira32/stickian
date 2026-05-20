package utils

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

const (
	// MaxRead is the default maximum ammount of bytes a write request may have
	//
	// Rule-of-thumb: use it to all POST/PUT/PATCH endpoints, and adjust down/up
	// if necessary with a local variable.
	MaxRead = 1024 * 1024
)

func WithDefaultOKHeaders(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Pragma", "no-cache")
	w.WriteHeader(http.StatusOK)
}

func WithError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrUserError):
		http.Error(w, err.Error(), http.StatusBadRequest)
	case errors.Is(err, ErrUnauthorized):
		http.Error(w, err.Error(), http.StatusUnauthorized)
	case errors.Is(err, ErrForbidden):
		http.Error(w, err.Error(), http.StatusForbidden)
	case errors.Is(err, ErrNotFound):
		http.Error(w, err.Error(), http.StatusNotFound)
	case errors.Is(err, ErrUnauthorized):
		http.Error(w, err.Error(), http.StatusUnauthorized)
	default:
		// TODO: do a better translation of internal errors to codes
		// and do logging instead of returning to client - this way we
		// can avoid exposing internal issues but be able to diagnose it
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func LoadBase64QueryParam[T any](r *http.Request) (T, error) {
	var result T

	v := r.URL.Query().Get("data")
	if v == "" {
		return result, fmt.Errorf("missing query parameter data")
	}

	decoded, err := base64.StdEncoding.DecodeString(v)
	if err != nil {
		return result, fmt.Errorf("%w: invalid base64 data: %w", ErrUserError, err)
	}

	if err := json.Unmarshal(decoded, &result); err != nil {
		return result, fmt.Errorf("%w: invalid JSON data: %w", ErrUserError, err)
	}

	return result, nil
}
