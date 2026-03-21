package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"

	apierrors "shopkuber/shared/errors"
	sharedmw "shopkuber/shared/middleware"
	"shopkuber/shared/pagination"
	sellersvc "shopkuber/seller/internal/seller/service"
)

var validate = validator.New()

// Handler handles seller HTTP endpoints.
type Handler struct {
	svc *sellersvc.Service
}

// New creates a new seller handler.
func New(svc *sellersvc.Service) *Handler {
	return &Handler{svc: svc}
}

// Register handles POST /api/v1/seller/register
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	claims, ok := sharedmw.ClaimsFromContext(r.Context())
	if !ok {
		apierrors.Respond(w, apierrors.ErrUnauthorized)
		return
	}
	var req struct {
		BrandName string `json:"brand_name" validate:"required"`
		INN       string `json:"inn"        validate:"required"`
	}
	if !bind(w, r, &req) {
		return
	}
	seller, err := h.svc.Register(r.Context(), claims.UserID, req.BrandName, req.INN)
	if err != nil {
		apierrors.Respond(w, err)
		return
	}
	respond(w, http.StatusCreated, seller)
}

// GetProfile handles GET /api/v1/seller/profile
func (h *Handler) GetProfile(w http.ResponseWriter, r *http.Request) {
	claims, ok := sharedmw.ClaimsFromContext(r.Context())
	if !ok {
		apierrors.Respond(w, apierrors.ErrUnauthorized)
		return
	}
	seller, err := h.svc.Profile(r.Context(), claims.UserID)
	if err != nil {
		apierrors.Respond(w, err)
		return
	}
	respond(w, http.StatusOK, seller)
}

// UpdateProfile handles PUT /api/v1/seller/profile
func (h *Handler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	claims, ok := sharedmw.ClaimsFromContext(r.Context())
	if !ok {
		apierrors.Respond(w, apierrors.ErrUnauthorized)
		return
	}
	var req struct {
		BrandName string `json:"brand_name" validate:"required"`
	}
	if !bind(w, r, &req) {
		return
	}
	seller, err := h.svc.UpdateProfile(r.Context(), claims.UserID, req.BrandName)
	if err != nil {
		apierrors.Respond(w, err)
		return
	}
	respond(w, http.StatusOK, seller)
}

// ListProducts handles GET /api/v1/seller/products
func (h *Handler) ListProducts(w http.ResponseWriter, r *http.Request) {
	claims, ok := sharedmw.ClaimsFromContext(r.Context())
	if !ok {
		apierrors.Respond(w, apierrors.ErrUnauthorized)
		return
	}
	pg := pagination.FromRequest(r)
	page, err := h.svc.ListProducts(r.Context(), claims.UserID, pg)
	if err != nil {
		apierrors.Respond(w, err)
		return
	}
	respond(w, http.StatusOK, page)
}

// CreateProduct handles POST /api/v1/seller/products
func (h *Handler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	claims, ok := sharedmw.ClaimsFromContext(r.Context())
	if !ok {
		apierrors.Respond(w, apierrors.ErrUnauthorized)
		return
	}
	var req struct {
		CategoryID  string  `json:"category_id"  validate:"required,uuid"`
		Name        string  `json:"name"         validate:"required"`
		Slug        string  `json:"slug"         validate:"required"`
		Description *string `json:"description"`
		Brand       *string `json:"brand"`
	}
	if !bind(w, r, &req) {
		return
	}
	product, err := h.svc.CreateProduct(r.Context(), claims.UserID, req.CategoryID, req.Name, req.Slug, req.Description, req.Brand)
	if err != nil {
		apierrors.Respond(w, err)
		return
	}
	respond(w, http.StatusCreated, product)
}

// GetProduct handles GET /api/v1/seller/products/{id}
func (h *Handler) GetProduct(w http.ResponseWriter, r *http.Request) {
	claims, ok := sharedmw.ClaimsFromContext(r.Context())
	if !ok {
		apierrors.Respond(w, apierrors.ErrUnauthorized)
		return
	}
	product, err := h.svc.GetProduct(r.Context(), claims.UserID, chi.URLParam(r, "id"))
	if err != nil {
		apierrors.Respond(w, err)
		return
	}
	respond(w, http.StatusOK, product)
}

// UpdateProduct handles PUT /api/v1/seller/products/{id}
func (h *Handler) UpdateProduct(w http.ResponseWriter, r *http.Request) {
	claims, ok := sharedmw.ClaimsFromContext(r.Context())
	if !ok {
		apierrors.Respond(w, apierrors.ErrUnauthorized)
		return
	}
	var req struct {
		CategoryID  string  `json:"category_id"  validate:"required,uuid"`
		Name        string  `json:"name"         validate:"required"`
		Slug        string  `json:"slug"         validate:"required"`
		Description *string `json:"description"`
		Brand       *string `json:"brand"`
	}
	if !bind(w, r, &req) {
		return
	}
	product, err := h.svc.UpdateProduct(r.Context(), claims.UserID, chi.URLParam(r, "id"), req.CategoryID, req.Name, req.Slug, req.Description, req.Brand)
	if err != nil {
		apierrors.Respond(w, err)
		return
	}
	respond(w, http.StatusOK, product)
}

// PublishProduct handles POST /api/v1/seller/products/{id}/publish
func (h *Handler) PublishProduct(w http.ResponseWriter, r *http.Request) {
	claims, ok := sharedmw.ClaimsFromContext(r.Context())
	if !ok {
		apierrors.Respond(w, apierrors.ErrUnauthorized)
		return
	}
	if err := h.svc.PublishProduct(r.Context(), claims.UserID, chi.URLParam(r, "id")); err != nil {
		apierrors.Respond(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ArchiveProduct handles DELETE /api/v1/seller/products/{id}
func (h *Handler) ArchiveProduct(w http.ResponseWriter, r *http.Request) {
	claims, ok := sharedmw.ClaimsFromContext(r.Context())
	if !ok {
		apierrors.Respond(w, apierrors.ErrUnauthorized)
		return
	}
	if err := h.svc.ArchiveProduct(r.Context(), claims.UserID, chi.URLParam(r, "id")); err != nil {
		apierrors.Respond(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// AddVariant handles POST /api/v1/seller/products/{id}/variants
func (h *Handler) AddVariant(w http.ResponseWriter, r *http.Request) {
	claims, ok := sharedmw.ClaimsFromContext(r.Context())
	if !ok {
		apierrors.Respond(w, apierrors.ErrUnauthorized)
		return
	}
	var req struct {
		SKU        string   `json:"sku"        validate:"required"`
		Price      float64  `json:"price"      validate:"required,gt=0"`
		OldPrice   *float64 `json:"old_price"`
		Stock      int      `json:"stock"      validate:"min=0"`
		Attributes string   `json:"attributes"` // raw JSON object
	}
	if !bind(w, r, &req) {
		return
	}
	if req.Attributes == "" {
		req.Attributes = "{}"
	}
	variant, err := h.svc.AddVariant(r.Context(), claims.UserID, chi.URLParam(r, "id"), req.SKU, req.Price, req.OldPrice, req.Stock, req.Attributes)
	if err != nil {
		apierrors.Respond(w, err)
		return
	}
	respond(w, http.StatusCreated, variant)
}

// UpdateVariant handles PUT /api/v1/seller/products/{id}/variants/{vid}
func (h *Handler) UpdateVariant(w http.ResponseWriter, r *http.Request) {
	claims, ok := sharedmw.ClaimsFromContext(r.Context())
	if !ok {
		apierrors.Respond(w, apierrors.ErrUnauthorized)
		return
	}
	var req struct {
		Price    float64  `json:"price"    validate:"required,gt=0"`
		OldPrice *float64 `json:"old_price"`
		Stock    int      `json:"stock"    validate:"min=0"`
	}
	if !bind(w, r, &req) {
		return
	}
	if err := h.svc.UpdateVariant(r.Context(), claims.UserID, chi.URLParam(r, "id"), chi.URLParam(r, "vid"), req.Price, req.OldPrice, req.Stock); err != nil {
		apierrors.Respond(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// DeleteVariant handles DELETE /api/v1/seller/products/{id}/variants/{vid}
func (h *Handler) DeleteVariant(w http.ResponseWriter, r *http.Request) {
	claims, ok := sharedmw.ClaimsFromContext(r.Context())
	if !ok {
		apierrors.Respond(w, apierrors.ErrUnauthorized)
		return
	}
	if err := h.svc.DeleteVariant(r.Context(), claims.UserID, chi.URLParam(r, "id"), chi.URLParam(r, "vid")); err != nil {
		apierrors.Respond(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// AddImage handles POST /api/v1/seller/products/{id}/images
func (h *Handler) AddImage(w http.ResponseWriter, r *http.Request) {
	claims, ok := sharedmw.ClaimsFromContext(r.Context())
	if !ok {
		apierrors.Respond(w, apierrors.ErrUnauthorized)
		return
	}
	var req struct {
		URL       string `json:"url"        validate:"required,url"`
		SortOrder int    `json:"sort_order"`
	}
	if !bind(w, r, &req) {
		return
	}
	img, err := h.svc.AddImage(r.Context(), claims.UserID, chi.URLParam(r, "id"), req.URL, req.SortOrder)
	if err != nil {
		apierrors.Respond(w, err)
		return
	}
	respond(w, http.StatusCreated, img)
}

// DeleteImage handles DELETE /api/v1/seller/products/{id}/images/{iid}
func (h *Handler) DeleteImage(w http.ResponseWriter, r *http.Request) {
	claims, ok := sharedmw.ClaimsFromContext(r.Context())
	if !ok {
		apierrors.Respond(w, apierrors.ErrUnauthorized)
		return
	}
	if err := h.svc.DeleteImage(r.Context(), claims.UserID, chi.URLParam(r, "id"), chi.URLParam(r, "iid")); err != nil {
		apierrors.Respond(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// AdminListSellers handles GET /api/v1/admin/sellers
func (h *Handler) AdminListSellers(w http.ResponseWriter, r *http.Request) {
	pg := pagination.FromRequest(r)
	status := r.URL.Query().Get("status")
	page, err := h.svc.ListSellers(r.Context(), status, pg)
	if err != nil {
		apierrors.Respond(w, err)
		return
	}
	respond(w, http.StatusOK, page)
}

// AdminUpdateSellerStatus handles PUT /api/v1/admin/sellers/{id}/approve and /block
func (h *Handler) AdminUpdateSellerStatus(status string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sellerID := chi.URLParam(r, "id")
		if err := h.svc.ApproveOrBlock(r.Context(), sellerID, status); err != nil {
			apierrors.Respond(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
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
