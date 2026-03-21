package pagination

import (
	"net/http"
	"strconv"
)

const (
	defaultLimit = 20
	maxLimit     = 100
)

// Request holds parsed pagination parameters.
type Request struct {
	Limit  int
	Offset int
}

// Page is a generic paginated response envelope.
type Page[T any] struct {
	Items  []T `json:"items"`
	Total  int `json:"total"`
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

// FromRequest parses limit and offset query params from an HTTP request.
func FromRequest(r *http.Request) Request {
	limit := defaultLimit
	offset := 0

	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			limit = n
		}
	}
	if limit > maxLimit {
		limit = maxLimit
	}

	if v := r.URL.Query().Get("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			offset = n
		}
	}

	return Request{Limit: limit, Offset: offset}
}

// NewPage constructs a Page response.
func NewPage[T any](items []T, total, limit, offset int) Page[T] {
	if items == nil {
		items = []T{}
	}
	return Page[T]{
		Items:  items,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	}
}
