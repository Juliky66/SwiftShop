package service

import (
	"context"

	apierrors "shopkuber/shared/errors"
	"shopkuber/shared/pagination"
	catalogclient "shopkuber/reviews/internal/catalog"
	ordersclient "shopkuber/reviews/internal/orders"
	reviewrepo "shopkuber/reviews/internal/review/repository"
)

// Service handles review business logic.
type Service struct {
	repo    *reviewrepo.Repository
	orders  *ordersclient.Client
	catalog *catalogclient.Client
}

// New creates a new review service.
func New(repo *reviewrepo.Repository, orders *ordersclient.Client, catalog *catalogclient.Client) *Service {
	return &Service{repo: repo, orders: orders, catalog: catalog}
}

// ListByProduct returns approved reviews for a product.
func (s *Service) ListByProduct(ctx context.Context, productID string, ratingFilter int, pg pagination.Request) (pagination.Page[reviewrepo.Review], error) {
	reviews, total, err := s.repo.ListByProduct(ctx, productID, ratingFilter, pg.Limit, pg.Offset)
	if err != nil {
		return pagination.Page[reviewrepo.Review]{}, err
	}
	return pagination.NewPage(reviews, total, pg.Limit, pg.Offset), nil
}

// Submit creates a new review, verifying the user has a delivered order.
func (s *Service) Submit(ctx context.Context, userID, productID string, rating int, title, body *string) (*reviewrepo.Review, error) {
	// Verify purchase
	orderID, err := s.orders.VerifiedPurchase(ctx, userID, productID)
	if err != nil {
		return nil, apierrors.Wrap(apierrors.ErrForbidden, "you can only review products from delivered orders")
	}

	rev, err := s.repo.Create(ctx, productID, userID, &orderID, rating, title, body)
	if err != nil {
		return nil, err
	}

	// Update aggregate rating in catalog (best-effort)
	go func() {
		avg, count, err := s.repo.AverageRating(context.Background(), productID)
		if err == nil {
			_ = s.catalog.UpdateRating(context.Background(), productID, avg, count)
		}
	}()

	return rev, nil
}

// Update edits a user's own review.
func (s *Service) Update(ctx context.Context, reviewID, userID string, rating int, title, body *string) (*reviewrepo.Review, error) {
	existing, err := s.repo.FindByID(ctx, reviewID)
	if err != nil {
		return nil, err
	}
	if existing.UserID != userID {
		return nil, apierrors.ErrForbidden
	}

	rev, err := s.repo.Update(ctx, reviewID, userID, rating, title, body)
	if err != nil {
		return nil, err
	}

	// Refresh aggregate
	go func() {
		avg, count, err := s.repo.AverageRating(context.Background(), existing.ProductID)
		if err == nil {
			_ = s.catalog.UpdateRating(context.Background(), existing.ProductID, avg, count)
		}
	}()

	return rev, nil
}

// Delete removes a user's own review.
func (s *Service) Delete(ctx context.Context, reviewID, userID string) error {
	existing, err := s.repo.FindByID(ctx, reviewID)
	if err != nil {
		return err
	}
	if existing.UserID != userID {
		return apierrors.ErrForbidden
	}

	if err := s.repo.Delete(ctx, reviewID, userID); err != nil {
		return err
	}

	go func() {
		avg, count, err := s.repo.AverageRating(context.Background(), existing.ProductID)
		if err == nil {
			_ = s.catalog.UpdateRating(context.Background(), existing.ProductID, avg, count)
		}
	}()

	return nil
}
