package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"

	apierrors "shopkuber/shared/errors"
)

// Product is the DB model for a seller-owned product.
type Product struct {
	ID          string  `db:"id"           json:"id"`
	SellerID    string  `db:"seller_id"    json:"seller_id"`
	CategoryID  string  `db:"category_id"  json:"category_id"`
	Name        string  `db:"name"         json:"name"`
	Slug        string  `db:"slug"         json:"slug"`
	Description *string `db:"description"  json:"description"`
	Brand       *string `db:"brand"        json:"brand"`
	Status      string  `db:"status"       json:"status"`
	Rating      float64 `db:"rating"       json:"rating"`
	ReviewCount int     `db:"review_count" json:"review_count"`
}

// Variant is the DB model for a product variant.
type Variant struct {
	ID         string   `db:"id"         json:"id"`
	ProductID  string   `db:"product_id" json:"product_id"`
	SKU        string   `db:"sku"        json:"sku"`
	Price      float64  `db:"price"      json:"price"`
	OldPrice   *float64 `db:"old_price"  json:"old_price"`
	Stock      int      `db:"stock"      json:"stock"`
	Attributes string   `db:"attributes" json:"attributes"` // raw JSON string
}

// Image is the DB model for a product image.
type Image struct {
	ID        string `db:"id"         json:"id"`
	ProductID string `db:"product_id" json:"product_id"`
	URL       string `db:"url"        json:"url"`
	SortOrder int    `db:"sort_order" json:"sort_order"`
}

// Repository handles seller product persistence.
type Repository struct {
	db *sqlx.DB
}

// New creates a new seller product repository.
func New(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

// Create creates a new product draft owned by a seller.
func (r *Repository) Create(ctx context.Context, sellerID, categoryID, name, slug string, description, brand *string) (*Product, error) {
	p := &Product{}
	err := r.db.QueryRowxContext(ctx,
		`INSERT INTO products (seller_id, category_id, name, slug, description, brand)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 RETURNING id, seller_id, category_id, name, slug, description, brand, status, rating, review_count`,
		sellerID, categoryID, name, slug, description, brand,
	).StructScan(p)
	if err != nil {
		if containsStr(err.Error(), "duplicate key") {
			return nil, apierrors.Wrap(apierrors.ErrConflict, "slug already in use")
		}
		return nil, err
	}
	return p, nil
}

// FindByID finds a product by ID, optionally scoped to a seller.
func (r *Repository) FindByID(ctx context.Context, id, sellerID string) (*Product, error) {
	p := &Product{}
	query := `SELECT id, seller_id, category_id, name, slug, description, brand, status, rating, review_count FROM products WHERE id = $1`
	args := []any{id}
	if sellerID != "" {
		query += ` AND seller_id = $2`
		args = append(args, sellerID)
	}
	err := r.db.QueryRowxContext(ctx, query, args...).StructScan(p)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apierrors.ErrNotFound
		}
		return nil, err
	}
	return p, nil
}

// Update updates a product's editable fields.
func (r *Repository) Update(ctx context.Context, id, sellerID, categoryID, name, slug string, description, brand *string) (*Product, error) {
	p := &Product{}
	err := r.db.QueryRowxContext(ctx,
		`UPDATE products
		 SET category_id = $1, name = $2, slug = $3, description = $4, brand = $5, updated_at = now()
		 WHERE id = $6 AND seller_id = $7
		 RETURNING id, seller_id, category_id, name, slug, description, brand, status, rating, review_count`,
		categoryID, name, slug, description, brand, id, sellerID,
	).StructScan(p)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apierrors.ErrNotFound
		}
		return nil, err
	}
	return p, nil
}

// Publish sets a product to published if it has at least one variant and one image.
func (r *Repository) Publish(ctx context.Context, id, sellerID string) error {
	var variantCount, imageCount int
	_ = r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM product_variants WHERE product_id = $1`, id).Scan(&variantCount)
	_ = r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM product_images WHERE product_id = $1`, id).Scan(&imageCount)
	if variantCount == 0 {
		return apierrors.Wrap(apierrors.ErrBadRequest, "product must have at least one variant")
	}
	if imageCount == 0 {
		return apierrors.Wrap(apierrors.ErrBadRequest, "product must have at least one image")
	}

	res, err := r.db.ExecContext(ctx,
		`UPDATE products SET status = 'published', updated_at = now() WHERE id = $1 AND seller_id = $2`,
		id, sellerID,
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

// Archive sets a product to archived.
func (r *Repository) Archive(ctx context.Context, id, sellerID string) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE products SET status = 'archived', updated_at = now() WHERE id = $1 AND seller_id = $2`,
		id, sellerID,
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

// List returns paginated products for a seller.
func (r *Repository) List(ctx context.Context, sellerID string, limit, offset int) ([]Product, int, error) {
	var total int
	_ = r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM products WHERE seller_id = $1`, sellerID).Scan(&total)

	var products []Product
	err := r.db.SelectContext(ctx, &products,
		`SELECT id, seller_id, category_id, name, slug, description, brand, status, rating, review_count
		 FROM products WHERE seller_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		sellerID, limit, offset,
	)
	return products, total, err
}

// AddVariant adds a variant to a product.
func (r *Repository) AddVariant(ctx context.Context, productID, sku string, price float64, oldPrice *float64, stock int, attributesJSON string) (*Variant, error) {
	v := &Variant{}
	err := r.db.QueryRowxContext(ctx,
		`INSERT INTO product_variants (product_id, sku, price, old_price, stock, attributes)
		 VALUES ($1, $2, $3, $4, $5, $6::jsonb)
		 RETURNING id, product_id, sku, price, old_price, stock, attributes::text`,
		productID, sku, price, oldPrice, stock, attributesJSON,
	).Scan(&v.ID, &v.ProductID, &v.SKU, &v.Price, &v.OldPrice, &v.Stock, &v.Attributes)
	if err != nil {
		if containsStr(err.Error(), "duplicate key") {
			return nil, apierrors.Wrap(apierrors.ErrConflict, "SKU already in use")
		}
		return nil, err
	}
	return v, nil
}

// UpdateVariant updates variant price/stock.
func (r *Repository) UpdateVariant(ctx context.Context, variantID, productID string, price float64, oldPrice *float64, stock int) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE product_variants SET price = $1, old_price = $2, stock = $3 WHERE id = $4 AND product_id = $5`,
		price, oldPrice, stock, variantID, productID,
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

// DeleteVariant removes a variant.
func (r *Repository) DeleteVariant(ctx context.Context, variantID, productID string) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM product_variants WHERE id = $1 AND product_id = $2`, variantID, productID,
	)
	return err
}

// AddImage adds an image URL to a product.
func (r *Repository) AddImage(ctx context.Context, productID, url string, sortOrder int) (*Image, error) {
	img := &Image{}
	err := r.db.QueryRowxContext(ctx,
		`INSERT INTO product_images (product_id, url, sort_order) VALUES ($1, $2, $3)
		 RETURNING id, product_id, url, sort_order`,
		productID, url, sortOrder,
	).StructScan(img)
	return img, err
}

// DeleteImage removes an image.
func (r *Repository) DeleteImage(ctx context.Context, imageID, productID string) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM product_images WHERE id = $1 AND product_id = $2`, imageID, productID,
	)
	return err
}

func containsStr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
