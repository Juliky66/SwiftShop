package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	apierrors "shopkuber/shared/errors"
	sharedmw "shopkuber/shared/middleware"
	"shopkuber/shared/pagination"
	reviewrepo "shopkuber/reviews/internal/review/repository"
	"shopkuber/reviews/internal/review/handler"
)

// ── mock ──────────────────────────────────────────────────────────────────────

type mockReviewSvc struct {
	listFn   func(ctx context.Context, productID string, ratingFilter int, pg pagination.Request) (pagination.Page[reviewrepo.Review], error)
	submitFn func(ctx context.Context, userID, productID string, rating int, title, body *string) (*reviewrepo.Review, error)
	updateFn func(ctx context.Context, reviewID, userID string, rating int, title, body *string) (*reviewrepo.Review, error)
	deleteFn func(ctx context.Context, reviewID, userID string) error
}

func (m *mockReviewSvc) ListByProduct(ctx context.Context, productID string, ratingFilter int, pg pagination.Request) (pagination.Page[reviewrepo.Review], error) {
	return m.listFn(ctx, productID, ratingFilter, pg)
}
func (m *mockReviewSvc) Submit(ctx context.Context, userID, productID string, rating int, title, body *string) (*reviewrepo.Review, error) {
	return m.submitFn(ctx, userID, productID, rating, title, body)
}
func (m *mockReviewSvc) Update(ctx context.Context, reviewID, userID string, rating int, title, body *string) (*reviewrepo.Review, error) {
	return m.updateFn(ctx, reviewID, userID, rating, title, body)
}
func (m *mockReviewSvc) Delete(ctx context.Context, reviewID, userID string) error {
	return m.deleteFn(ctx, reviewID, userID)
}

// ── helpers ───────────────────────────────────────────────────────────────────

func newRouter(h *handler.Handler) *chi.Mux {
	r := chi.NewRouter()
	r.Get("/products/{id}/reviews", h.List)
	r.Post("/products/{id}/reviews", h.Submit)
	r.Put("/reviews/{id}", h.Update)
	r.Delete("/reviews/{id}", h.Delete)
	return r
}

func withClaims(r *http.Request, userID string) *http.Request {
	ctx := context.WithValue(r.Context(), sharedmw.ClaimsKey, &sharedmw.Claims{UserID: userID, Role: "buyer"})
	return r.WithContext(ctx)
}

func jsonBody(t *testing.T, v any) *bytes.Buffer {
	t.Helper()
	b, err := json.Marshal(v)
	require.NoError(t, err)
	return bytes.NewBuffer(b)
}

// ── List ──────────────────────────────────────────────────────────────────────

func TestList_OK(t *testing.T) {
	rev := reviewrepo.Review{ID: "rev-1", ProductID: "prod-1", UserID: "user-1", Rating: 5}
	svc := &mockReviewSvc{
		listFn: func(_ context.Context, _ string, _ int, _ pagination.Request) (pagination.Page[reviewrepo.Review], error) {
			return pagination.NewPage([]reviewrepo.Review{rev}, 1, 10, 0), nil
		},
	}
	h := handler.New(svc)
	r := newRouter(h)

	req := httptest.NewRequest(http.MethodGet, "/products/prod-1/reviews", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Header().Get("Content-Type"), "application/json")
}

func TestList_ServiceError(t *testing.T) {
	svc := &mockReviewSvc{
		listFn: func(_ context.Context, _ string, _ int, _ pagination.Request) (pagination.Page[reviewrepo.Review], error) {
			return pagination.Page[reviewrepo.Review]{}, apierrors.ErrNotFound
		},
	}
	h := handler.New(svc)
	r := newRouter(h)

	req := httptest.NewRequest(http.MethodGet, "/products/missing/reviews", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

// ── Submit ────────────────────────────────────────────────────────────────────

func TestSubmit_NoAuth(t *testing.T) {
	h := handler.New(&mockReviewSvc{})
	r := newRouter(h)

	body := jsonBody(t, map[string]any{"rating": 5})
	req := httptest.NewRequest(http.MethodPost, "/products/prod-1/reviews", body)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestSubmit_InvalidRating(t *testing.T) {
	h := handler.New(&mockReviewSvc{})
	r := newRouter(h)

	body := jsonBody(t, map[string]any{"rating": 0}) // rating must be 1–5
	req := httptest.NewRequest(http.MethodPost, "/products/prod-1/reviews", body)
	req = withClaims(req, "user-1")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestSubmit_Success(t *testing.T) {
	rev := &reviewrepo.Review{ID: "rev-1", ProductID: "prod-1", UserID: "user-1", Rating: 4}
	svc := &mockReviewSvc{
		submitFn: func(_ context.Context, _, _ string, _ int, _, _ *string) (*reviewrepo.Review, error) {
			return rev, nil
		},
	}
	h := handler.New(svc)
	r := newRouter(h)

	body := jsonBody(t, map[string]any{"rating": 4})
	req := httptest.NewRequest(http.MethodPost, "/products/prod-1/reviews", body)
	req = withClaims(req, "user-1")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)
}

func TestSubmit_Forbidden(t *testing.T) {
	svc := &mockReviewSvc{
		submitFn: func(_ context.Context, _, _ string, _ int, _, _ *string) (*reviewrepo.Review, error) {
			return nil, apierrors.ErrForbidden
		},
	}
	h := handler.New(svc)
	r := newRouter(h)

	body := jsonBody(t, map[string]any{"rating": 3})
	req := httptest.NewRequest(http.MethodPost, "/products/prod-1/reviews", body)
	req = withClaims(req, "user-1")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusForbidden, rec.Code)
}

// ── Update ────────────────────────────────────────────────────────────────────

func TestUpdate_NoAuth(t *testing.T) {
	h := handler.New(&mockReviewSvc{})
	r := newRouter(h)

	body := jsonBody(t, map[string]any{"rating": 3})
	req := httptest.NewRequest(http.MethodPut, "/reviews/rev-1", body)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestUpdate_Forbidden(t *testing.T) {
	svc := &mockReviewSvc{
		updateFn: func(_ context.Context, _, _ string, _ int, _, _ *string) (*reviewrepo.Review, error) {
			return nil, apierrors.ErrForbidden
		},
	}
	h := handler.New(svc)
	r := newRouter(h)

	body := jsonBody(t, map[string]any{"rating": 3})
	req := httptest.NewRequest(http.MethodPut, "/reviews/rev-1", body)
	req = withClaims(req, "other-user")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusForbidden, rec.Code)
}

func TestUpdate_Success(t *testing.T) {
	rev := &reviewrepo.Review{ID: "rev-1", UserID: "user-1", Rating: 3}
	svc := &mockReviewSvc{
		updateFn: func(_ context.Context, _, _ string, _ int, _, _ *string) (*reviewrepo.Review, error) {
			return rev, nil
		},
	}
	h := handler.New(svc)
	r := newRouter(h)

	body := jsonBody(t, map[string]any{"rating": 3})
	req := httptest.NewRequest(http.MethodPut, "/reviews/rev-1", body)
	req = withClaims(req, "user-1")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

// ── Delete ────────────────────────────────────────────────────────────────────

func TestDelete_NoAuth(t *testing.T) {
	h := handler.New(&mockReviewSvc{})
	r := newRouter(h)

	req := httptest.NewRequest(http.MethodDelete, "/reviews/rev-1", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestDelete_Success(t *testing.T) {
	svc := &mockReviewSvc{
		deleteFn: func(_ context.Context, _, _ string) error { return nil },
	}
	h := handler.New(svc)
	r := newRouter(h)

	req := httptest.NewRequest(http.MethodDelete, "/reviews/rev-1", nil)
	req = withClaims(req, "user-1")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNoContent, rec.Code)
}

func TestDelete_NotFound(t *testing.T) {
	svc := &mockReviewSvc{
		deleteFn: func(_ context.Context, _, _ string) error { return apierrors.ErrNotFound },
	}
	h := handler.New(svc)
	r := newRouter(h)

	req := httptest.NewRequest(http.MethodDelete, "/reviews/missing", nil)
	req = withClaims(req, "user-1")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}
