package errors

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

// Sentinel errors used across services.
var (
	ErrNotFound     = errors.New("not found")
	ErrUnauthorized = errors.New("unauthorized")
	ErrForbidden    = errors.New("forbidden")
	ErrConflict     = errors.New("conflict")
	ErrBadRequest   = errors.New("bad request")
	ErrInternal     = errors.New("internal server error")
)

// APIError is the standard JSON error response body.
type APIError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// Respond writes a JSON error response.
func Respond(w http.ResponseWriter, err error) {
	code, msg := mapError(err)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(APIError{Code: code, Message: msg})
}

func mapError(err error) (int, string) {
	switch {
	case errors.Is(err, ErrNotFound):
		return http.StatusNotFound, err.Error()
	case errors.Is(err, ErrUnauthorized):
		return http.StatusUnauthorized, err.Error()
	case errors.Is(err, ErrForbidden):
		return http.StatusForbidden, err.Error()
	case errors.Is(err, ErrConflict):
		return http.StatusConflict, err.Error()
	case errors.Is(err, ErrBadRequest):
		return http.StatusBadRequest, err.Error()
	default:
		return http.StatusInternalServerError, "internal server error"
	}
}

// Wrap wraps a sentinel error with an additional message.
func Wrap(sentinel error, msg string) error {
	return fmt.Errorf("%w: %s", sentinel, msg)
}
