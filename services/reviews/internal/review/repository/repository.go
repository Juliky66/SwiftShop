package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"

	apierrors "shopkuber/shared/errors"
)

// Review is the DB model.
type Review struct {
	ID         string  `db:"id"          json:"id"`
	ProductID  string  `db:"product_id"  json:"product_id"`
	UserID     string  `db:"user_id"     json:"user_id"`
	OrderID    *string `db:"order_id"    json:"order_id"`
	Rating     int     `db:"rating"      json:"rating"`
	Title      *string `db:"title"       json:"title"`
	Body       *string `db:"body"        json:"body"`
	IsApproved bool    `db:"is_approved" json:"is_approved"`
}

// Repository handles review persistence.
type Repository struct {
	db *sqlx.DB
}

// New creates a new review repository.
func New(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

// Create inserts a new review.
func (r *Repository) Create(ctx context.Context, productID, userID string, orderID *string, rating int, title, body *string) (*Review, error) {
	rev := &Review{}
	err := r.db.QueryRowxContext(ctx,
		`INSERT INTO reviews (product_id, user_id, order_id, rating, title, body)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 RETURNING id, product_id, user_id, order_id, rating, title, body, is_approved`,
		productID, userID, orderID, rating, title, body,
	).StructScan(rev)
	if err != nil {
		if containsStr(err.Error(), "duplicate key") {
			return nil, apierrors.Wrap(apierrors.ErrConflict, "you have already reviewed this product")
		}
		return nil, err
	}
	return rev, nil
}

// FindByID finds a review by primary key.
func (r *Repository) FindByID(ctx context.Context, id string) (*Review, error) {
	rev := &Review{}
	err := r.db.QueryRowxContext(ctx,
		`SELECT id, product_id, user_id, order_id, rating, title, body, is_approved FROM reviews WHERE id = $1`, id,
	).StructScan(rev)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apierrors.ErrNotFound
		}
		return nil, err
	}
	return rev, nil
}

// ListByProduct returns paginated approved reviews for a product.
func (r *Repository) ListByProduct(ctx context.Context, productID string, ratingFilter int, limit, offset int) ([]Review, int, error) {
	cond := `WHERE product_id = $1 AND is_approved = true`
	args := []any{productID}
	if ratingFilter > 0 {
		cond += ` AND rating = $2`
		args = append(args, ratingFilter)
	}

	var total int
	countArgs := make([]any, len(args))
	copy(countArgs, args)
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM reviews `+cond, countArgs...).Scan(&total); err != nil {
		return nil, 0, err
	}

	args = append(args, limit, offset)
	limitIdx := len(args) - 1
	offsetIdx := len(args)
	_ = limitIdx
	_ = offsetIdx

	query := `SELECT id, product_id, user_id, order_id, rating, title, body, is_approved FROM reviews ` + cond
	if ratingFilter > 0 {
		query += ` ORDER BY created_at DESC LIMIT $3 OFFSET $4`
	} else {
		query += ` ORDER BY created_at DESC LIMIT $2 OFFSET $3`
	}

	var reviews []Review
	err := r.db.SelectContext(ctx, &reviews, query, args...)
	return reviews, total, err
}

// Update edits an existing review.
func (r *Repository) Update(ctx context.Context, id, userID string, rating int, title, body *string) (*Review, error) {
	rev := &Review{}
	err := r.db.QueryRowxContext(ctx,
		`UPDATE reviews SET rating = $1, title = $2, body = $3 WHERE id = $4 AND user_id = $5
		 RETURNING id, product_id, user_id, order_id, rating, title, body, is_approved`,
		rating, title, body, id, userID,
	).StructScan(rev)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apierrors.ErrNotFound
		}
		return nil, err
	}
	return rev, nil
}

// Delete removes a review owned by the user.
func (r *Repository) Delete(ctx context.Context, id, userID string) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM reviews WHERE id = $1 AND user_id = $2`, id, userID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return apierrors.ErrNotFound
	}
	return nil
}

// AverageRating computes the average rating and count for a product.
func (r *Repository) AverageRating(ctx context.Context, productID string) (float64, int, error) {
	var avg sql.NullFloat64
	var count int
	err := r.db.QueryRowContext(ctx,
		`SELECT AVG(rating), COUNT(*) FROM reviews WHERE product_id = $1 AND is_approved = true`, productID,
	).Scan(&avg, &count)
	if err != nil {
		return 0, 0, err
	}
	return avg.Float64, count, nil
}

func containsStr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
