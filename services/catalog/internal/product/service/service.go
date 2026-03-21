package service

import (
	"context"

	catrepo "shopkuber/catalog/internal/category/repository"
	prodrepo "shopkuber/catalog/internal/product/repository"
	apierrors "shopkuber/shared/errors"
	"shopkuber/shared/pagination"
)

// ProductDetail holds a product with its variants and images.
type ProductDetail struct {
	*prodrepo.Product
	Variants []prodrepo.Variant `json:"variants"`
	Images   []prodrepo.Image   `json:"images"`
}

// Service handles catalog business logic.
type Service struct {
	products   *prodrepo.Repository
	categories *catrepo.Repository
}

// New creates a new catalog service.
func New(products *prodrepo.Repository, categories *catrepo.Repository) *Service {
	return &Service{products: products, categories: categories}
}

// Search returns a paginated list of products matching the filter.
func (s *Service) Search(ctx context.Context, f prodrepo.SearchFilter, pg pagination.Request) (pagination.Page[prodrepo.Product], error) {
	// If category slug provided, resolve to ID
	if f.CategoryID == "" && f.Query == "" && f.Brand == "" {
		// no filter — all published products
	}

	items, total, err := s.products.Search(ctx, f, pg.Limit, pg.Offset)
	if err != nil {
		return pagination.Page[prodrepo.Product]{}, err
	}
	return pagination.NewPage(items, total, pg.Limit, pg.Offset), nil
}

// GetBySlug returns full product detail including variants and images.
func (s *Service) GetBySlug(ctx context.Context, slug string) (*ProductDetail, error) {
	p, err := s.products.FindBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}

	variants, err := s.products.Variants(ctx, p.ID)
	if err != nil {
		return nil, err
	}

	images, err := s.products.Images(ctx, p.ID)
	if err != nil {
		return nil, err
	}

	return &ProductDetail{Product: p, Variants: variants, Images: images}, nil
}

// GetCategoryProducts returns products in a category by slug.
func (s *Service) GetCategoryProducts(ctx context.Context, categorySlug string, f prodrepo.SearchFilter, pg pagination.Request) (pagination.Page[prodrepo.Product], error) {
	cat, err := s.categories.FindBySlug(ctx, categorySlug)
	if err != nil {
		return pagination.Page[prodrepo.Product]{}, err
	}
	f.CategoryID = cat.ID
	return s.Search(ctx, f, pg)
}

// ReserveStock reserves stock for a variant (called by orders service).
func (s *Service) ReserveStock(ctx context.Context, variantID string, qty int) error {
	if qty <= 0 {
		return apierrors.Wrap(apierrors.ErrBadRequest, "quantity must be positive")
	}
	return s.products.ReserveStock(ctx, variantID, qty)
}

// GetVariant returns a single variant with its product info for checkout snapshot.
func (s *Service) GetVariant(ctx context.Context, variantID string) (*prodrepo.Variant, *prodrepo.Product, error) {
	v, err := s.products.FindVariantByID(ctx, variantID)
	if err != nil {
		return nil, nil, err
	}
	p, err := s.products.FindByID(ctx, v.ProductID)
	if err != nil {
		return nil, nil, err
	}
	return v, p, nil
}

// UpdateRating updates a product's rating after a review is submitted.
func (s *Service) UpdateRating(ctx context.Context, productID string, rating float64, reviewCount int) error {
	return s.products.UpdateRating(ctx, productID, rating, reviewCount)
}
