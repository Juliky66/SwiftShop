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
	"shopkuber/reviews/internal/payment/handler"
	paymentrepo "shopkuber/reviews/internal/payment/repository"
	paymentsvc "shopkuber/reviews/internal/payment/service"
)

// ── mock ──────────────────────────────────────────────────────────────────────

type mockPaymentSvc struct {
	initiateFn func(ctx context.Context, orderID, userID string) (*paymentsvc.InitiateResponse, error)
	statusFn   func(ctx context.Context, paymentID string) (*paymentrepo.Payment, error)
	webhookFn  func(ctx context.Context, rawBody []byte, sig string, event paymentsvc.WebhookEvent) error
}

func (m *mockPaymentSvc) Initiate(ctx context.Context, orderID, userID string) (*paymentsvc.InitiateResponse, error) {
	return m.initiateFn(ctx, orderID, userID)
}
func (m *mockPaymentSvc) GetStatus(ctx context.Context, paymentID string) (*paymentrepo.Payment, error) {
	return m.statusFn(ctx, paymentID)
}
func (m *mockPaymentSvc) HandleWebhook(ctx context.Context, rawBody []byte, sig string, event paymentsvc.WebhookEvent) error {
	return m.webhookFn(ctx, rawBody, sig, event)
}

// ── helpers ───────────────────────────────────────────────────────────────────

func newRouter(h *handler.Handler) *chi.Mux {
	r := chi.NewRouter()
	r.Post("/orders/{id}/pay", h.Initiate)
	r.Get("/payments/{id}", h.GetStatus)
	r.Post("/payments/webhook", h.Webhook)
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

// ── Initiate ──────────────────────────────────────────────────────────────────

func TestInitiate_NoAuth(t *testing.T) {
	h := handler.New(&mockPaymentSvc{})
	r := newRouter(h)

	req := httptest.NewRequest(http.MethodPost, "/orders/order-1/pay", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestInitiate_Success(t *testing.T) {
	svc := &mockPaymentSvc{
		initiateFn: func(_ context.Context, _, _ string) (*paymentsvc.InitiateResponse, error) {
			return &paymentsvc.InitiateResponse{
				PaymentID:  "pay-1",
				ConfirmURL: "http://pay.example.com/confirm",
				Status:     "pending",
			}, nil
		},
	}
	h := handler.New(svc)
	r := newRouter(h)

	req := httptest.NewRequest(http.MethodPost, "/orders/order-1/pay", nil)
	req = withClaims(req, "user-1")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)

	var resp paymentsvc.InitiateResponse
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
	assert.Equal(t, "pay-1", resp.PaymentID)
	assert.Equal(t, "pending", resp.Status)
}

func TestInitiate_OrderNotFound(t *testing.T) {
	svc := &mockPaymentSvc{
		initiateFn: func(_ context.Context, _, _ string) (*paymentsvc.InitiateResponse, error) {
			return nil, apierrors.ErrNotFound
		},
	}
	h := handler.New(svc)
	r := newRouter(h)

	req := httptest.NewRequest(http.MethodPost, "/orders/missing/pay", nil)
	req = withClaims(req, "user-1")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

// ── GetStatus ─────────────────────────────────────────────────────────────────

func TestGetStatus_OK(t *testing.T) {
	pid := "prov-123"
	svc := &mockPaymentSvc{
		statusFn: func(_ context.Context, _ string) (*paymentrepo.Payment, error) {
			return &paymentrepo.Payment{
				ID:         "pay-1",
				OrderID:    "order-1",
				Provider:   "mock",
				ProviderID: &pid,
				Status:     "succeeded",
				Amount:     999.0,
				Currency:   "RUB",
			}, nil
		},
	}
	h := handler.New(svc)
	r := newRouter(h)

	req := httptest.NewRequest(http.MethodGet, "/payments/pay-1", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var p paymentrepo.Payment
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&p))
	assert.Equal(t, "succeeded", p.Status)
}

func TestGetStatus_NotFound(t *testing.T) {
	svc := &mockPaymentSvc{
		statusFn: func(_ context.Context, _ string) (*paymentrepo.Payment, error) {
			return nil, apierrors.ErrNotFound
		},
	}
	h := handler.New(svc)
	r := newRouter(h)

	req := httptest.NewRequest(http.MethodGet, "/payments/unknown", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

// ── Webhook ───────────────────────────────────────────────────────────────────

func TestWebhook_InvalidJSON(t *testing.T) {
	h := handler.New(&mockPaymentSvc{})
	r := newRouter(h)

	req := httptest.NewRequest(http.MethodPost, "/payments/webhook", bytes.NewBufferString("{invalid"))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestWebhook_Unauthorized(t *testing.T) {
	svc := &mockPaymentSvc{
		webhookFn: func(_ context.Context, _ []byte, _ string, _ paymentsvc.WebhookEvent) error {
			return apierrors.ErrUnauthorized
		},
	}
	h := handler.New(svc)
	r := newRouter(h)

	body := jsonBody(t, map[string]string{"ProviderPaymentID": "pay-1", "Status": "paid"})
	req := httptest.NewRequest(http.MethodPost, "/payments/webhook", body)
	req.Header.Set("X-Signature", "bad-signature")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestWebhook_Success(t *testing.T) {
	svc := &mockPaymentSvc{
		webhookFn: func(_ context.Context, _ []byte, _ string, _ paymentsvc.WebhookEvent) error {
			return nil
		},
	}
	h := handler.New(svc)
	r := newRouter(h)

	body := jsonBody(t, map[string]string{"ProviderPaymentID": "pay-1", "Status": "paid"})
	req := httptest.NewRequest(http.MethodPost, "/payments/webhook", body)
	req.Header.Set("X-Signature", "valid-signature")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}
