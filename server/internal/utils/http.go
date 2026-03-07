package utils

import (
	"errors"
	"net/http"
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
	case errors.Is(err, ErrNotFound):
		http.Error(w, err.Error(), http.StatusNotFound)
	default:
		// TODO: do a better translation of internal errors to codes
		// and do logging instead of returning to client - this way we
		// can avoid exposing internal issues but be able to diagnose it
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
