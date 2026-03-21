package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"

	apierrors "shopkuber/shared/errors"
)

// Order is the DB model.
type Order struct {
	ID              string  `db:"id"               json:"id"`
	UserID          string  `db:"user_id"          json:"user_id"`
	Status          string  `db:"status"           json:"status"`
	TotalAmount     float64 `db:"total_amount"     json:"total_amount"`
	DeliveryAddress string  `db:"delivery_address" json:"shipping_address"` // frontend key is shipping_address
	PaymentID       *string `db:"payment_id"       json:"payment_id"`
	PaymentMethod   *string `db:"payment_method"   json:"payment_method"`
}

// OrderItem is a line item in an order.
type OrderItem struct {
	ID          string  `db:"id"           json:"id"`
	OrderID     string  `db:"order_id"     json:"order_id"`
	VariantID   string  `db:"variant_id"   json:"variant_id"`
	SellerID    string  `db:"seller_id"    json:"seller_id"`
	ProductName string  `db:"product_name" json:"product_name"`
	SKU         string  `db:"sku"          json:"sku"`
	Quantity    int     `db:"quantity"     json:"quantity"`
	UnitPrice   float64 `db:"unit_price"   json:"unit_price"`
	TotalPrice  float64 `db:"total_price"  json:"total_price"`
}

// CreateOrderInput holds data for creating an order + items.
type CreateOrderInput struct {
	UserID          string
	TotalAmount     float64
	DeliveryAddress string // JSON string
	Items           []CreateOrderItemInput
}

// CreateOrderItemInput holds data for a single order item.
type CreateOrderItemInput struct {
	VariantID   string
	SellerID    string
	ProductName string
	SKU         string
	Quantity    int
	UnitPrice   float64
}

// Repository handles order persistence.
type Repository struct {
	db *sqlx.DB
}

// New creates a new order repository.
func New(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

// Create inserts an order and its items within a transaction.
func (r *Repository) Create(ctx context.Context, input CreateOrderInput) (*Order, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	order := &Order{}
	err = tx.QueryRowxContext(ctx,
		`INSERT INTO orders (user_id, total_amount, delivery_address)
		 VALUES ($1, $2, $3::jsonb)
		 RETURNING id, user_id, status, total_amount, delivery_address::text, payment_id, payment_method`,
		input.UserID, input.TotalAmount, input.DeliveryAddress,
	).StructScan(order)
	if err != nil {
		return nil, err
	}

	for _, item := range input.Items {
		_, err = tx.ExecContext(ctx,
			`INSERT INTO order_items (order_id, variant_id, seller_id, product_name, sku, quantity, unit_price, total_price)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
			order.ID, item.VariantID, item.SellerID, item.ProductName, item.SKU,
			item.Quantity, item.UnitPrice, float64(item.Quantity)*item.UnitPrice,
		)
		if err != nil {
			return nil, err
		}
	}

	return order, tx.Commit()
}

// FindByID returns an order by ID, ensuring ownership.
func (r *Repository) FindByID(ctx context.Context, orderID, userID string) (*Order, error) {
	o := &Order{}
	err := r.db.QueryRowxContext(ctx,
		`SELECT id, user_id, status, total_amount, delivery_address::text, payment_id, payment_method
		 FROM orders WHERE id = $1 AND user_id = $2`,
		orderID, userID,
	).StructScan(o)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apierrors.ErrNotFound
		}
		return nil, err
	}
	return o, nil
}

// FindByIDAdmin returns an order by ID without ownership check.
func (r *Repository) FindByIDAdmin(ctx context.Context, orderID string) (*Order, error) {
	o := &Order{}
	err := r.db.QueryRowxContext(ctx,
		`SELECT id, user_id, status, total_amount, delivery_address::text, payment_id, payment_method
		 FROM orders WHERE id = $1`, orderID,
	).StructScan(o)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apierrors.ErrNotFound
		}
		return nil, err
	}
	return o, nil
}

// ListByUser returns paginated orders for a user.
func (r *Repository) ListByUser(ctx context.Context, userID string, limit, offset int) ([]Order, int, error) {
	var total int
	if err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM orders WHERE user_id = $1`, userID,
	).Scan(&total); err != nil {
		return nil, 0, err
	}

	var orders []Order
	err := r.db.SelectContext(ctx, &orders,
		`SELECT id, user_id, status, total_amount, delivery_address::text, payment_id, payment_method
		 FROM orders WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		userID, limit, offset,
	)
	return orders, total, err
}

// Items returns all items for an order.
func (r *Repository) Items(ctx context.Context, orderID string) ([]OrderItem, error) {
	var items []OrderItem
	err := r.db.SelectContext(ctx, &items,
		`SELECT id, order_id, variant_id, seller_id, product_name, sku, quantity, unit_price, total_price
		 FROM order_items WHERE order_id = $1`,
		orderID,
	)
	return items, err
}

// UpdateStatus changes the order status.
func (r *Repository) UpdateStatus(ctx context.Context, orderID, status string) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE orders SET status = $1, updated_at = now() WHERE id = $2`,
		status, orderID,
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

// SetPaymentID stores the external payment reference on an order.
func (r *Repository) SetPaymentID(ctx context.Context, orderID, paymentID string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE orders SET payment_id = $1, updated_at = now() WHERE id = $2`,
		paymentID, orderID,
	)
	return err
}

// VerifiedPurchase checks if a user has a delivered order containing a given product.
func (r *Repository) VerifiedPurchase(ctx context.Context, userID, productID string) (string, error) {
	var orderID string
	err := r.db.QueryRowContext(ctx,
		`SELECT o.id FROM orders o
		 JOIN order_items oi ON oi.order_id = o.id
		 JOIN product_variants pv ON pv.id = oi.variant_id
		 WHERE o.user_id = $1 AND pv.product_id = $2 AND o.status = 'delivered'
		 LIMIT 1`,
		userID, productID,
	).Scan(&orderID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", apierrors.ErrNotFound
		}
		return "", err
	}
	return orderID, nil
}

// ListBySellerItems returns orders that contain items from a specific seller.
func (r *Repository) ListBySellerItems(ctx context.Context, sellerID string, limit, offset int) ([]Order, int, error) {
	var total int
	if err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(DISTINCT o.id) FROM orders o
		 JOIN order_items oi ON oi.order_id = o.id
		 WHERE oi.seller_id = $1`, sellerID,
	).Scan(&total); err != nil {
		return nil, 0, err
	}

	var orders []Order
	err := r.db.SelectContext(ctx, &orders,
		`SELECT DISTINCT o.id, o.user_id, o.status, o.total_amount, o.delivery_address::text, o.payment_id, o.payment_method
		 FROM orders o
		 JOIN order_items oi ON oi.order_id = o.id
		 WHERE oi.seller_id = $1
		 ORDER BY o.id DESC LIMIT $2 OFFSET $3`,
		sellerID, limit, offset,
	)
	return orders, total, err
}
