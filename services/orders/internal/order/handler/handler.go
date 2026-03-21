package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"

	apierrors "shopkuber/shared/errors"
	sharedmw "shopkuber/shared/middleware"
	"shopkuber/shared/pagination"
	ordersvc "shopkuber/orders/internal/order/service"
)

var validate = validator.New()

// Handler handles order and cart HTTP endpoints.
type Handler struct {
	svc *ordersvc.Service
}

// New creates a new order handler.
func New(svc *ordersvc.Service) *Handler {
	return &Handler{svc: svc}
}

// GetCart handles GET /api/v1/cart
func (h *Handler) GetCart(w http.ResponseWriter, r *http.Request) {
	claims, ok := sharedmw.ClaimsFromContext(r.Context())
	if !ok {
		apierrors.Respond(w, apierrors.ErrUnauthorized)
		return
	}
	cart, items, err := h.svc.GetCart(r.Context(), claims.UserID)
	if err != nil {
		apierrors.Respond(w, err)
		return
	}
	respond(w, http.StatusOK, map[string]any{"cart": cart, "items": items})
}

// AddToCart handles POST /api/v1/cart/items
func (h *Handler) AddToCart(w http.ResponseWriter, r *http.Request) {
	claims, ok := sharedmw.ClaimsFromContext(r.Context())
	if !ok {
		apierrors.Respond(w, apierrors.ErrUnauthorized)
		return
	}
	var req struct {
		VariantID string `json:"variant_id" validate:"required,uuid"`
		Quantity  int    `json:"quantity"   validate:"required,min=1"`
	}
	if !bind(w, r, &req) {
		return
	}
	if err := h.svc.AddToCart(r.Context(), claims.UserID, req.VariantID, req.Quantity); err != nil {
		apierrors.Respond(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// UpdateCartItem handles PUT /api/v1/cart/items/{variant_id}
func (h *Handler) UpdateCartItem(w http.ResponseWriter, r *http.Request) {
	claims, ok := sharedmw.ClaimsFromContext(r.Context())
	if !ok {
		apierrors.Respond(w, apierrors.ErrUnauthorized)
		return
	}
	variantID := chi.URLParam(r, "variant_id")
	var req struct {
		Quantity int `json:"quantity" validate:"min=0"`
	}
	if !bind(w, r, &req) {
		return
	}
	if err := h.svc.UpdateCartItem(r.Context(), claims.UserID, variantID, req.Quantity); err != nil {
		apierrors.Respond(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// RemoveFromCart handles DELETE /api/v1/cart/items/{variant_id}
func (h *Handler) RemoveFromCart(w http.ResponseWriter, r *http.Request) {
	claims, ok := sharedmw.ClaimsFromContext(r.Context())
	if !ok {
		apierrors.Respond(w, apierrors.ErrUnauthorized)
		return
	}
	variantID := chi.URLParam(r, "variant_id")
	if err := h.svc.RemoveFromCart(r.Context(), claims.UserID, variantID); err != nil {
		apierrors.Respond(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ClearCart handles DELETE /api/v1/cart
func (h *Handler) ClearCart(w http.ResponseWriter, r *http.Request) {
	claims, ok := sharedmw.ClaimsFromContext(r.Context())
	if !ok {
		apierrors.Respond(w, apierrors.ErrUnauthorized)
		return
	}
	if err := h.svc.ClearCart(r.Context(), claims.UserID); err != nil {
		apierrors.Respond(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Checkout handles POST /api/v1/orders
func (h *Handler) Checkout(w http.ResponseWriter, r *http.Request) {
	claims, ok := sharedmw.ClaimsFromContext(r.Context())
	if !ok {
		apierrors.Respond(w, apierrors.ErrUnauthorized)
		return
	}
	var req struct {
		DeliveryAddress ordersvc.DeliveryAddress `json:"delivery_address" validate:"required"`
	}
	if !bind(w, r, &req) {
		return
	}
	order, err := h.svc.Checkout(r.Context(), claims.UserID, req.DeliveryAddress)
	if err != nil {
		apierrors.Respond(w, err)
		return
	}
	respond(w, http.StatusCreated, order)
}

// ListOrders handles GET /api/v1/orders
func (h *Handler) ListOrders(w http.ResponseWriter, r *http.Request) {
	claims, ok := sharedmw.ClaimsFromContext(r.Context())
	if !ok {
		apierrors.Respond(w, apierrors.ErrUnauthorized)
		return
	}
	pg := pagination.FromRequest(r)
	page, err := h.svc.ListOrders(r.Context(), claims.UserID, pg)
	if err != nil {
		apierrors.Respond(w, err)
		return
	}
	respond(w, http.StatusOK, page)
}

// GetOrder handles GET /api/v1/orders/{id}
func (h *Handler) GetOrder(w http.ResponseWriter, r *http.Request) {
	claims, ok := sharedmw.ClaimsFromContext(r.Context())
	if !ok {
		apierrors.Respond(w, apierrors.ErrUnauthorized)
		return
	}
	orderID := chi.URLParam(r, "id")
	order, err := h.svc.GetOrder(r.Context(), orderID, claims.UserID)
	if err != nil {
		apierrors.Respond(w, err)
		return
	}
	respond(w, http.StatusOK, order)
}

// CancelOrder handles POST /api/v1/orders/{id}/cancel
func (h *Handler) CancelOrder(w http.ResponseWriter, r *http.Request) {
	claims, ok := sharedmw.ClaimsFromContext(r.Context())
	if !ok {
		apierrors.Respond(w, apierrors.ErrUnauthorized)
		return
	}
	orderID := chi.URLParam(r, "id")
	if err := h.svc.CancelOrder(r.Context(), orderID, claims.UserID); err != nil {
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
