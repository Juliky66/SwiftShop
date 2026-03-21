package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"

	apierrors "shopkuber/shared/errors"
)

// Payment is the DB model.
type Payment struct {
	ID         string  `db:"id"          json:"id"`
	OrderID    string  `db:"order_id"    json:"order_id"`
	Provider   string  `db:"provider"    json:"provider"`
	ProviderID *string `db:"provider_id" json:"provider_id"`
	Status     string  `db:"status"      json:"status"`
	Amount     float64 `db:"amount"      json:"amount"`
	Currency   string  `db:"currency"    json:"currency"`
}

// Repository handles payment persistence.
type Repository struct {
	db *sqlx.DB
}

// New creates a new payment repository.
func New(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

// Create records a new payment.
func (r *Repository) Create(ctx context.Context, orderID, provider, providerID string, amount float64, currency string) (*Payment, error) {
	p := &Payment{}
	err := r.db.QueryRowxContext(ctx,
		`INSERT INTO payments (order_id, provider, provider_id, amount, currency)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, order_id, provider, provider_id, status, amount, currency`,
		orderID, provider, providerID, amount, currency,
	).StructScan(p)
	return p, err
}

// FindByID returns a payment by ID.
func (r *Repository) FindByID(ctx context.Context, id string) (*Payment, error) {
	p := &Payment{}
	err := r.db.QueryRowxContext(ctx,
		`SELECT id, order_id, provider, provider_id, status, amount, currency FROM payments WHERE id = $1`, id,
	).StructScan(p)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apierrors.ErrNotFound
		}
		return nil, err
	}
	return p, nil
}

// FindByOrderID returns a payment by order ID.
func (r *Repository) FindByOrderID(ctx context.Context, orderID string) (*Payment, error) {
	p := &Payment{}
	err := r.db.QueryRowxContext(ctx,
		`SELECT id, order_id, provider, provider_id, status, amount, currency FROM payments WHERE order_id = $1`, orderID,
	).StructScan(p)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apierrors.ErrNotFound
		}
		return nil, err
	}
	return p, nil
}

// FindByProviderID looks up a payment by the external provider transaction ID.
func (r *Repository) FindByProviderID(ctx context.Context, providerID string) (*Payment, error) {
	p := &Payment{}
	err := r.db.QueryRowxContext(ctx,
		`SELECT id, order_id, provider, provider_id, status, amount, currency FROM payments WHERE provider_id = $1`, providerID,
	).StructScan(p)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apierrors.ErrNotFound
		}
		return nil, err
	}
	return p, nil
}

// UpdateStatus changes a payment's status.
func (r *Repository) UpdateStatus(ctx context.Context, id, status string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE payments SET status = $1, updated_at = now() WHERE id = $2`, status, id,
	)
	return err
}
