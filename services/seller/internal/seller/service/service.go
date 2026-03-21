package service

import (
	"context"

	apierrors "shopkuber/shared/errors"
	"shopkuber/shared/pagination"
	sellerrepo "shopkuber/seller/internal/seller/repository"
	prodrepo "shopkuber/seller/internal/product/repository"
)

// Service handles seller business logic.
type Service struct {
	sellers  *sellerrepo.Repository
	products *prodrepo.Repository
}

// New creates a new seller service.
func New(sellers *sellerrepo.Repository, products *prodrepo.Repository) *Service {
	return &Service{sellers: sellers, products: products}
}

// Register registers a new seller application for the authenticated user.
func (s *Service) Register(ctx context.Context, userID, brandName, inn string) (*sellerrepo.Seller, error) {
	return s.sellers.Create(ctx, userID, brandName, inn)
}

// Profile returns the seller profile for the authenticated user.
func (s *Service) Profile(ctx context.Context, userID string) (*sellerrepo.Seller, error) {
	return s.sellers.FindByUserID(ctx, userID)
}

// UpdateProfile updates the seller's brand name.
func (s *Service) UpdateProfile(ctx context.Context, userID, brandName string) (*sellerrepo.Seller, error) {
	seller, err := s.sellers.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if seller.Status != "approved" {
		return nil, apierrors.Wrap(apierrors.ErrForbidden, "seller account is not approved")
	}
	return s.sellers.UpdateProfile(ctx, seller.ID, brandName)
}

// ApproveOrBlock changes a seller's status (admin only).
func (s *Service) ApproveOrBlock(ctx context.Context, sellerID, status string) error {
	allowed := map[string]bool{"approved": true, "blocked": true, "pending": true}
	if !allowed[status] {
		return apierrors.Wrap(apierrors.ErrBadRequest, "invalid status")
	}
	return s.sellers.UpdateStatus(ctx, sellerID, status)
}

// ListSellers returns paginated sellers (admin only).
func (s *Service) ListSellers(ctx context.Context, status string, pg pagination.Request) (pagination.Page[sellerrepo.Seller], error) {
	sellers, total, err := s.sellers.List(ctx, status, pg.Limit, pg.Offset)
	if err != nil {
		return pagination.Page[sellerrepo.Seller]{}, err
	}
	return pagination.NewPage(sellers, total, pg.Limit, pg.Offset), nil
}

// CreateProduct creates a product draft for the seller.
func (s *Service) CreateProduct(ctx context.Context, userID, categoryID, name, slug string, description, brand *string) (*prodrepo.Product, error) {
	seller, err := s.sellers.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if seller.Status != "approved" {
		return nil, apierrors.Wrap(apierrors.ErrForbidden, "seller account is not approved")
	}
	return s.products.Create(ctx, seller.ID, categoryID, name, slug, description, brand)
}

// GetProduct returns a seller's product by ID.
func (s *Service) GetProduct(ctx context.Context, userID, productID string) (*prodrepo.Product, error) {
	seller, err := s.sellers.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return s.products.FindByID(ctx, productID, seller.ID)
}

// UpdateProduct updates a seller's product.
func (s *Service) UpdateProduct(ctx context.Context, userID, productID, categoryID, name, slug string, description, brand *string) (*prodrepo.Product, error) {
	seller, err := s.sellers.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return s.products.Update(ctx, productID, seller.ID, categoryID, name, slug, description, brand)
}

// PublishProduct submits a product for publishing.
func (s *Service) PublishProduct(ctx context.Context, userID, productID string) error {
	seller, err := s.sellers.FindByUserID(ctx, userID)
	if err != nil {
		return err
	}
	return s.products.Publish(ctx, productID, seller.ID)
}

// ArchiveProduct archives a product.
func (s *Service) ArchiveProduct(ctx context.Context, userID, productID string) error {
	seller, err := s.sellers.FindByUserID(ctx, userID)
	if err != nil {
		return err
	}
	return s.products.Archive(ctx, productID, seller.ID)
}

// ListProducts returns paginated products for the seller.
func (s *Service) ListProducts(ctx context.Context, userID string, pg pagination.Request) (pagination.Page[prodrepo.Product], error) {
	seller, err := s.sellers.FindByUserID(ctx, userID)
	if err != nil {
		return pagination.Page[prodrepo.Product]{}, err
	}
	products, total, err := s.products.List(ctx, seller.ID, pg.Limit, pg.Offset)
	if err != nil {
		return pagination.Page[prodrepo.Product]{}, err
	}
	return pagination.NewPage(products, total, pg.Limit, pg.Offset), nil
}

// AddVariant adds a variant to a seller's product.
func (s *Service) AddVariant(ctx context.Context, userID, productID, sku string, price float64, oldPrice *float64, stock int, attributesJSON string) (*prodrepo.Variant, error) {
	seller, err := s.sellers.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	// Verify product belongs to seller
	if _, err := s.products.FindByID(ctx, productID, seller.ID); err != nil {
		return nil, err
	}
	return s.products.AddVariant(ctx, productID, sku, price, oldPrice, stock, attributesJSON)
}

// UpdateVariant updates a variant's price and stock.
func (s *Service) UpdateVariant(ctx context.Context, userID, productID, variantID string, price float64, oldPrice *float64, stock int) error {
	seller, err := s.sellers.FindByUserID(ctx, userID)
	if err != nil {
		return err
	}
	if _, err := s.products.FindByID(ctx, productID, seller.ID); err != nil {
		return err
	}
	return s.products.UpdateVariant(ctx, variantID, productID, price, oldPrice, stock)
}

// DeleteVariant removes a variant.
func (s *Service) DeleteVariant(ctx context.Context, userID, productID, variantID string) error {
	seller, err := s.sellers.FindByUserID(ctx, userID)
	if err != nil {
		return err
	}
	if _, err := s.products.FindByID(ctx, productID, seller.ID); err != nil {
		return err
	}
	return s.products.DeleteVariant(ctx, variantID, productID)
}

// AddImage adds an image to a product.
func (s *Service) AddImage(ctx context.Context, userID, productID, url string, sortOrder int) (*prodrepo.Image, error) {
	seller, err := s.sellers.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if _, err := s.products.FindByID(ctx, productID, seller.ID); err != nil {
		return nil, err
	}
	return s.products.AddImage(ctx, productID, url, sortOrder)
}

// DeleteImage removes an image.
func (s *Service) DeleteImage(ctx context.Context, userID, productID, imageID string) error {
	seller, err := s.sellers.FindByUserID(ctx, userID)
	if err != nil {
		return err
	}
	if _, err := s.products.FindByID(ctx, productID, seller.ID); err != nil {
		return err
	}
	return s.products.DeleteImage(ctx, imageID, productID)
}
