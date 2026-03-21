package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	apierrors "shopkuber/shared/errors"
	"shopkuber/shared/pagination"
	prodrepo "shopkuber/catalog/internal/product/repository"
	prodsvc "shopkuber/catalog/internal/product/service"
	catsvc "shopkuber/catalog/internal/category/service"
)

// Handler handles product HTTP endpoints.
type Handler struct {
	productSvc  *prodsvc.Service
	categorySvc *catsvc.Service
}

// New creates a new product handler.
func New(productSvc *prodsvc.Service, categorySvc *catsvc.Service) *Handler {
	return &Handler{productSvc: productSvc, categorySvc: categorySvc}
}

// List handles GET /api/v1/products
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	pg := pagination.FromRequest(r)
	f := parseFilter(r)

	page, err := h.productSvc.Search(r.Context(), f, pg)
	if err != nil {
		apierrors.Respond(w, err)
		return
	}
	respond(w, http.StatusOK, page)
}

// Search handles GET /api/v1/products/search
func (h *Handler) Search(w http.ResponseWriter, r *http.Request) {
	pg := pagination.FromRequest(r)
	f := parseFilter(r)
	f.Query = r.URL.Query().Get("q")

	page, err := h.productSvc.Search(r.Context(), f, pg)
	if err != nil {
		apierrors.Respond(w, err)
		return
	}
	respond(w, http.StatusOK, page)
}

// Get handles GET /api/v1/products/{slug}
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	detail, err := h.productSvc.GetBySlug(r.Context(), slug)
	if err != nil {
		apierrors.Respond(w, err)
		return
	}
	respond(w, http.StatusOK, detail)
}

// CategoryProducts handles GET /api/v1/categories/{slug}/products
func (h *Handler) CategoryProducts(w http.ResponseWriter, r *http.Request) {
	categorySlug := chi.URLParam(r, "slug")
	pg := pagination.FromRequest(r)
	f := parseFilter(r)

	page, err := h.productSvc.GetCategoryProducts(r.Context(), categorySlug, f, pg)
	if err != nil {
		apierrors.Respond(w, err)
		return
	}
	respond(w, http.StatusOK, page)
}

// parseFilter reads common filter query params.
func parseFilter(r *http.Request) prodrepo.SearchFilter {
	q := r.URL.Query()
	f := prodrepo.SearchFilter{
		CategoryID: q.Get("category_id"),
		SellerID:   q.Get("seller_id"),
		Brand:      q.Get("brand"),
		Sort:       q.Get("sort"),
	}

	if v := q.Get("price_min"); v != "" {
		if n, err := strconv.ParseFloat(v, 64); err == nil {
			f.PriceMin = &n
		}
	}
	if v := q.Get("price_max"); v != "" {
		if n, err := strconv.ParseFloat(v, 64); err == nil {
			f.PriceMax = &n
		}
	}
	if v := q.Get("in_stock"); v == "true" {
		b := true
		f.InStock = &b
	}

	return f
}

func respond(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
