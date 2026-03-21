package handler

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"

	apierrors "shopkuber/shared/errors"
	sharedmw "shopkuber/shared/middleware"
	paymentrepo "shopkuber/reviews/internal/payment/repository"
	paymentsvc "shopkuber/reviews/internal/payment/service"
)

// PaymentService is the interface the handler depends on.
type PaymentService interface {
	Initiate(ctx context.Context, orderID, userID string) (*paymentsvc.InitiateResponse, error)
	GetStatus(ctx context.Context, paymentID string) (*paymentrepo.Payment, error)
	HandleWebhook(ctx context.Context, rawBody []byte, signature string, event paymentsvc.WebhookEvent) error
}

// Handler handles payment HTTP requests.
type Handler struct {
	svc PaymentService
}

// New creates a new payment handler.
func New(svc PaymentService) *Handler {
	return &Handler{svc: svc}
}

// Initiate handles POST /orders/{id}/pay
func (h *Handler) Initiate(w http.ResponseWriter, r *http.Request) {
	claims, ok := sharedmw.ClaimsFromContext(r.Context())
	if !ok {
		apierrors.Respond(w, apierrors.ErrUnauthorized)
		return
	}
	resp, err := h.svc.Initiate(r.Context(), chi.URLParam(r, "id"), claims.UserID)
	if err != nil {
		apierrors.Respond(w, err)
		return
	}
	respond(w, http.StatusCreated, resp)
}

// GetStatus handles GET /payments/{id}
func (h *Handler) GetStatus(w http.ResponseWriter, r *http.Request) {
	payment, err := h.svc.GetStatus(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		apierrors.Respond(w, err)
		return
	}
	respond(w, http.StatusOK, payment)
}

// Webhook handles POST /payments/webhook
func (h *Handler) Webhook(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "cannot read body", http.StatusBadRequest)
		return
	}
	signature := r.Header.Get("X-Signature")

	var event paymentsvc.WebhookEvent
	if err := json.Unmarshal(body, &event); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	if err := h.svc.HandleWebhook(r.Context(), body, signature, event); err != nil {
		apierrors.Respond(w, err)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func respond(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
