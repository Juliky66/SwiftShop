package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"

	apierrors "shopkuber/shared/errors"
	sharedmw "shopkuber/shared/middleware"
	"shopkuber/shared/pagination"
	reviewrepo "shopkuber/reviews/internal/review/repository"
)

var validate = validator.New()

// ReviewService is the interface the handler depends on.
type ReviewService interface {
	ListByProduct(ctx context.Context, productID string, ratingFilter int, pg pagination.Request) (pagination.Page[reviewrepo.Review], error)
	Submit(ctx context.Context, userID, productID string, rating int, title, body *string) (*reviewrepo.Review, error)
	Update(ctx context.Context, reviewID, userID string, rating int, title, body *string) (*reviewrepo.Review, error)
	Delete(ctx context.Context, reviewID, userID string) error
}

// Handler handles review HTTP requests.
type Handler struct {
	svc ReviewService
}

// New creates a new review handler.
func New(svc ReviewService) *Handler {
	return &Handler{svc: svc}
}

// List handles GET /products/{id}/reviews
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	productID := chi.URLParam(r, "id")
	pg := pagination.FromRequest(r)
	ratingFilter := 0
	if v := r.URL.Query().Get("rating"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			ratingFilter = n
		}
	}
	page, err := h.svc.ListByProduct(r.Context(), productID, ratingFilter, pg)
	if err != nil {
		apierrors.Respond(w, err)
		return
	}
	respond(w, http.StatusOK, page)
}

// Submit handles POST /products/{id}/reviews
func (h *Handler) Submit(w http.ResponseWriter, r *http.Request) {
	claims, ok := sharedmw.ClaimsFromContext(r.Context())
	if !ok {
		apierrors.Respond(w, apierrors.ErrUnauthorized)
		return
	}
	var req struct {
		Rating int     `json:"rating" validate:"required,min=1,max=5"`
		Title  *string `json:"title"`
		Body   *string `json:"body"`
	}
	if !bind(w, r, &req) {
		return
	}
	rev, err := h.svc.Submit(r.Context(), claims.UserID, chi.URLParam(r, "id"), req.Rating, req.Title, req.Body)
	if err != nil {
		apierrors.Respond(w, err)
		return
	}
	respond(w, http.StatusCreated, rev)
}

// Update handles PUT /reviews/{id}
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	claims, ok := sharedmw.ClaimsFromContext(r.Context())
	if !ok {
		apierrors.Respond(w, apierrors.ErrUnauthorized)
		return
	}
	var req struct {
		Rating int     `json:"rating" validate:"required,min=1,max=5"`
		Title  *string `json:"title"`
		Body   *string `json:"body"`
	}
	if !bind(w, r, &req) {
		return
	}
	rev, err := h.svc.Update(r.Context(), chi.URLParam(r, "id"), claims.UserID, req.Rating, req.Title, req.Body)
	if err != nil {
		apierrors.Respond(w, err)
		return
	}
	respond(w, http.StatusOK, rev)
}

// Delete handles DELETE /reviews/{id}
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	claims, ok := sharedmw.ClaimsFromContext(r.Context())
	if !ok {
		apierrors.Respond(w, apierrors.ErrUnauthorized)
		return
	}
	if err := h.svc.Delete(r.Context(), chi.URLParam(r, "id"), claims.UserID); err != nil {
		apierrors.Respond(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ── helpers ───────────────────────────────────────────────────────────────────

func bind(w http.ResponseWriter, r *http.Request, dst any) bool {
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		apierrors.Respond(w, apierrors.Wrap(apierrors.ErrBadRequest, "invalid JSON"))
		return false
	}
	if err := validate.Struct(dst); err != nil {
		apierrors.Respond(w, apierrors.Wrap(apierrors.ErrBadRequest, err.Error()))
		return false
	}
	return true
}

func respond(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
