package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"

	apierrors "shopkuber/shared/errors"
)

// Cart is the DB model.
type Cart struct {
	ID     string `db:"id"      json:"id"`
	UserID string `db:"user_id" json:"user_id"`
}

// CartItem is a line item in the cart.
type CartItem struct {
	ID            string  `db:"id"             json:"id"`
	CartID        string  `db:"cart_id"        json:"cart_id"`
	VariantID     string  `db:"variant_id"     json:"variant_id"`
	Quantity      int     `db:"quantity"       json:"quantity"`
	PriceSnapshot float64 `db:"price_snapshot" json:"price_snapshot"`
}

// Repository handles cart persistence.
type Repository struct {
	db *sqlx.DB
}

// New creates a new cart repository.
func New(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

// GetOrCreate returns the cart for a user, creating it if it doesn't exist.
func (r *Repository) GetOrCreate(ctx context.Context, userID string) (*Cart, error) {
	c := &Cart{}
	err := r.db.QueryRowxContext(ctx,
		`INSERT INTO carts (user_id) VALUES ($1)
		 ON CONFLICT (user_id) DO UPDATE SET updated_at = now()
		 RETURNING id, user_id`, userID,
	).StructScan(c)
	return c, err
}

// Items returns all items in a cart.
func (r *Repository) Items(ctx context.Context, cartID string) ([]CartItem, error) {
	var items []CartItem
	err := r.db.SelectContext(ctx, &items,
		`SELECT id, cart_id, variant_id, quantity, price_snapshot FROM cart_items WHERE cart_id = $1`,
		cartID,
	)
	return items, err
}

// UpsertItem adds or updates a cart item. Uses ON CONFLICT to update quantity.
func (r *Repository) UpsertItem(ctx context.Context, cartID, variantID string, quantity int, price float64) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO cart_items (cart_id, variant_id, quantity, price_snapshot)
		 VALUES ($1, $2, $3, $4)
		 ON CONFLICT (cart_id, variant_id) DO UPDATE
		   SET quantity = $3, price_snapshot = $4`,
		cartID, variantID, quantity, price,
	)
	return err
}

// UpdateQuantity sets the quantity of an existing cart item.
func (r *Repository) UpdateQuantity(ctx context.Context, cartID, variantID string, quantity int) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE cart_items SET quantity = $1 WHERE cart_id = $2 AND variant_id = $3`,
		quantity, cartID, variantID,
	)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return apierrors.ErrNotFound
	}
	return nil
}

// RemoveItem deletes a specific item from the cart.
func (r *Repository) RemoveItem(ctx context.Context, cartID, variantID string) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM cart_items WHERE cart_id = $1 AND variant_id = $2`,
		cartID, variantID,
	)
	return err
}

// Clear removes all items from the cart.
func (r *Repository) Clear(ctx context.Context, cartID string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM cart_items WHERE cart_id = $1`, cartID)
	return err
}

// FindItem finds a specific cart item.
func (r *Repository) FindItem(ctx context.Context, cartID, variantID string) (*CartItem, error) {
	item := &CartItem{}
	err := r.db.QueryRowxContext(ctx,
		`SELECT id, cart_id, variant_id, quantity, price_snapshot FROM cart_items WHERE cart_id = $1 AND variant_id = $2`,
		cartID, variantID,
	).StructScan(item)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apierrors.ErrNotFound
		}
		return nil, err
	}
	return item, nil
}
