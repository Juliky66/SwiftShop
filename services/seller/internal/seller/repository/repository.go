package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"

	apierrors "shopkuber/shared/errors"
)

// Seller is the DB model.
type Seller struct {
	ID        string  `db:"id"         json:"id"`
	UserID    string  `db:"user_id"    json:"user_id"`
	BrandName string  `db:"brand_name" json:"brand_name"`
	INN       string  `db:"inn"        json:"inn"`
	Status    string  `db:"status"     json:"status"`
	Rating    float64 `db:"rating"     json:"rating"`
}

// Repository handles seller persistence.
type Repository struct {
	db *sqlx.DB
}

// New creates a new seller repository.
func New(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

// Create registers a new seller application.
func (r *Repository) Create(ctx context.Context, userID, brandName, inn string) (*Seller, error) {
	s := &Seller{}
	err := r.db.QueryRowxContext(ctx,
		`INSERT INTO sellers (user_id, brand_name, inn)
		 VALUES ($1, $2, $3)
		 RETURNING id, user_id, brand_name, inn, status, rating`,
		userID, brandName, inn,
	).StructScan(s)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, apierrors.Wrap(apierrors.ErrConflict, "seller already registered or INN already in use")
		}
		return nil, err
	}
	return s, nil
}

// FindByUserID finds a seller by the owning user's ID.
func (r *Repository) FindByUserID(ctx context.Context, userID string) (*Seller, error) {
	s := &Seller{}
	err := r.db.QueryRowxContext(ctx,
		`SELECT id, user_id, brand_name, inn, status, rating FROM sellers WHERE user_id = $1`, userID,
	).StructScan(s)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apierrors.ErrNotFound
		}
		return nil, err
	}
	return s, nil
}

// FindByID finds a seller by primary key.
func (r *Repository) FindByID(ctx context.Context, id string) (*Seller, error) {
	s := &Seller{}
	err := r.db.QueryRowxContext(ctx,
		`SELECT id, user_id, brand_name, inn, status, rating FROM sellers WHERE id = $1`, id,
	).StructScan(s)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apierrors.ErrNotFound
		}
		return nil, err
	}
	return s, nil
}

// UpdateProfile updates seller brand name.
func (r *Repository) UpdateProfile(ctx context.Context, id, brandName string) (*Seller, error) {
	s := &Seller{}
	err := r.db.QueryRowxContext(ctx,
		`UPDATE sellers SET brand_name = $1 WHERE id = $2
		 RETURNING id, user_id, brand_name, inn, status, rating`,
		brandName, id,
	).StructScan(s)
	return s, err
}

// UpdateStatus changes seller approval status.
func (r *Repository) UpdateStatus(ctx context.Context, id, status string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE sellers SET status = $1 WHERE id = $2`, status, id)
	return err
}

// List returns sellers with optional status filter.
func (r *Repository) List(ctx context.Context, status string, limit, offset int) ([]Seller, int, error) {
	var total int
	query := `SELECT COUNT(*) FROM sellers`
	args := []any{}
	if status != "" {
		query += ` WHERE status = $1`
		args = append(args, status)
	}
	if err := r.db.QueryRowContext(ctx, query, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	dataQ := `SELECT id, user_id, brand_name, inn, status, rating FROM sellers`
	if status != "" {
		dataQ += ` WHERE status = $1 LIMIT $2 OFFSET $3`
		args = append(args, limit, offset)
	} else {
		dataQ += ` LIMIT $1 OFFSET $2`
		args = append(args, limit, offset)
	}

	var sellers []Seller
	err := r.db.SelectContext(ctx, &sellers, dataQ, args...)
	return sellers, total, err
}

func isUniqueViolation(err error) bool {
	return err != nil && containsStr(err.Error(), "duplicate key")
}

func containsStr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
