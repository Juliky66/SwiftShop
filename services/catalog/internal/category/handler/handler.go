package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	apierrors "shopkuber/shared/errors"
	catsvc "shopkuber/catalog/internal/category/service"
)

// Handler handles category HTTP endpoints.
type Handler struct {
	svc *catsvc.Service
}

// New creates a new category handler.
func New(svc *catsvc.Service) *Handler {
	return &Handler{svc: svc}
}

// List handles GET /api/v1/categories
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	tree, err := h.svc.Tree(r.Context())
	if err != nil {
		apierrors.Respond(w, err)
		return
	}
	respond(w, http.StatusOK, tree)
}

// Get handles GET /api/v1/categories/{slug}
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	cat, err := h.svc.GetBySlug(r.Context(), slug)
	if err != nil {
		apierrors.Respond(w, err)
		return
	}
	respond(w, http.StatusOK, cat)
}

func respond(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
