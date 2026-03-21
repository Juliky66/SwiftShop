package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"

	apierrors "shopkuber/shared/errors"
)

// Product is the DB model.
type Product struct {
	ID              string  `db:"id"                json:"id"`
	SellerID        string  `db:"seller_id"         json:"seller_id"`
	CategoryID      string  `db:"category_id"       json:"category_id"`
	Name            string  `db:"name"              json:"name"`
	Slug            string  `db:"slug"              json:"slug"`
	Description     *string `db:"description"       json:"description"`
	Brand           *string `db:"brand"             json:"brand"`
	Status          string  `db:"status"            json:"status"`
	Rating          float64 `db:"rating"            json:"rating"`
	ReviewCount     int     `db:"review_count"      json:"review_count"`
	PrimaryImageURL *string `db:"primary_image_url" json:"primary_image_url,omitempty"`
}

// Variant is the DB model for a product variant.
type Variant struct {
	ID         string         `db:"id"         json:"id"`
	ProductID  string         `db:"product_id" json:"product_id"`
	SKU        string         `db:"sku"        json:"sku"`
	Price      float64        `db:"price"      json:"price"`
	OldPrice   *float64       `db:"old_price"  json:"old_price"`
	Stock      int            `db:"stock"      json:"stock"`
	Attributes map[string]any `db:"attributes" json:"attributes"`
}

// Image is the DB model for a product image.
type Image struct {
	ID        string `db:"id"         json:"id"`
	ProductID string `db:"product_id" json:"product_id"`
	URL       string `db:"url"        json:"url"`
	SortOrder int    `db:"sort_order" json:"sort_order"`
}

// SearchFilter holds optional search/filter parameters.
type SearchFilter struct {
	Query      string
	CategoryID string
	SellerID   string
	Brand      string
	PriceMin   *float64
	PriceMax   *float64
	InStock    *bool
	Sort       string // rating | price_asc | price_desc | newest
}

// Repository handles product persistence.
type Repository struct {
	db *sqlx.DB
}

// New creates a new product repository.
func New(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

// Search queries products with optional full-text search and filters.
func (r *Repository) Search(ctx context.Context, f SearchFilter, limit, offset int) ([]Product, int, error) {
	conds := []string{"p.status = 'published'"}
	args := []any{}
	idx := 1

	if f.Query != "" {
		conds = append(conds, fmt.Sprintf(
			"(p.name ILIKE $%d OR p.brand ILIKE $%d OR p.description ILIKE $%d)",
			idx, idx+1, idx+2,
		))
		like := "%" + f.Query + "%"
		args = append(args, like, like, like)
		idx += 3
	}
	if f.CategoryID != "" {
		conds = append(conds, fmt.Sprintf("p.category_id = $%d", idx))
		args = append(args, f.CategoryID)
		idx++
	}
	if f.SellerID != "" {
		conds = append(conds, fmt.Sprintf("p.seller_id = $%d", idx))
		args = append(args, f.SellerID)
		idx++
	}
	if f.Brand != "" {
		conds = append(conds, fmt.Sprintf("p.brand ILIKE $%d", idx))
		args = append(args, "%"+f.Brand+"%")
		idx++
	}
	if f.PriceMin != nil {
		conds = append(conds, fmt.Sprintf("EXISTS (SELECT 1 FROM product_variants v WHERE v.product_id = p.id AND v.price >= $%d)", idx))
		args = append(args, *f.PriceMin)
		idx++
	}
	if f.PriceMax != nil {
		conds = append(conds, fmt.Sprintf("EXISTS (SELECT 1 FROM product_variants v WHERE v.product_id = p.id AND v.price <= $%d)", idx))
		args = append(args, *f.PriceMax)
		idx++
	}
	if f.InStock != nil && *f.InStock {
		conds = append(conds, "EXISTS (SELECT 1 FROM product_variants v WHERE v.product_id = p.id AND v.stock > 0)")
	}

	where := "WHERE " + strings.Join(conds, " AND ")

	orderBy := "p.created_at DESC"
	switch f.Sort {
	case "rating":
		orderBy = "p.rating DESC, p.review_count DESC"
	case "price_asc":
		orderBy = "(SELECT MIN(price) FROM product_variants WHERE product_id = p.id) ASC"
	case "price_desc":
		orderBy = "(SELECT MIN(price) FROM product_variants WHERE product_id = p.id) DESC"
	}

	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM products p %s`, where)
	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	args = append(args, limit, offset)
	dataQuery := fmt.Sprintf(
		`SELECT p.id, p.seller_id, p.category_id, p.name, p.slug, p.description, p.brand, p.status, p.rating, p.review_count,
		        (SELECT url FROM product_images WHERE product_id = p.id ORDER BY sort_order LIMIT 1) AS primary_image_url
		 FROM products p %s ORDER BY %s LIMIT $%d OFFSET $%d`,
		where, orderBy, idx, idx+1,
	)

	var products []Product
	if err := r.db.SelectContext(ctx, &products, dataQuery, args...); err != nil {
		return nil, 0, err
	}
	return products, total, nil
}

// FindBySlug returns a single published product by slug.
func (r *Repository) FindBySlug(ctx context.Context, slug string) (*Product, error) {
	p := &Product{}
	err := r.db.QueryRowxContext(ctx,
		`SELECT id, seller_id, category_id, name, slug, description, brand, status, rating, review_count
		 FROM products WHERE slug = $1`, slug,
	).StructScan(p)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apierrors.ErrNotFound
		}
		return nil, err
	}
	return p, nil
}

// FindByID returns a product by its UUID.
func (r *Repository) FindByID(ctx context.Context, id string) (*Product, error) {
	p := &Product{}
	err := r.db.QueryRowxContext(ctx,
		`SELECT id, seller_id, category_id, name, slug, description, brand, status, rating, review_count
		 FROM products WHERE id = $1`, id,
	).StructScan(p)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apierrors.ErrNotFound
		}
		return nil, err
	}
	return p, nil
}

// Variants returns all variants for a product.
func (r *Repository) Variants(ctx context.Context, productID string) ([]Variant, error) {
	rows, err := r.db.QueryxContext(ctx,
		`SELECT id, product_id, sku, price, old_price, stock, attributes::text FROM product_variants WHERE product_id = $1`,
		productID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var variants []Variant
	for rows.Next() {
		var v Variant
		var attrJSON string
		if err := rows.Scan(&v.ID, &v.ProductID, &v.SKU, &v.Price, &v.OldPrice, &v.Stock, &attrJSON); err != nil {
			return nil, err
		}
		v.Attributes = map[string]any{"raw": attrJSON}
		variants = append(variants, v)
	}
	return variants, nil
}

// FindVariantByID returns a single variant.
func (r *Repository) FindVariantByID(ctx context.Context, variantID string) (*Variant, error) {
	v := &Variant{}
	var attrJSON string
	err := r.db.QueryRowContext(ctx,
		`SELECT id, product_id, sku, price, old_price, stock, attributes::text FROM product_variants WHERE id = $1`,
		variantID,
	).Scan(&v.ID, &v.ProductID, &v.SKU, &v.Price, &v.OldPrice, &v.Stock, &attrJSON)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apierrors.ErrNotFound
		}
		return nil, err
	}
	return v, nil
}

// ReserveStock decrements stock for a variant atomically. Returns ErrConflict if insufficient.
func (r *Repository) ReserveStock(ctx context.Context, variantID string, qty int) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE product_variants SET stock = stock - $1 WHERE id = $2 AND stock >= $1`,
		qty, variantID,
	)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return apierrors.Wrap(apierrors.ErrConflict, "insufficient stock")
	}
	return nil
}

// UpdateRating updates a product's aggregate rating and review count.
func (r *Repository) UpdateRating(ctx context.Context, productID string, rating float64, reviewCount int) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE products SET rating = $1, review_count = $2, updated_at = now() WHERE id = $3`,
		rating, reviewCount, productID,
	)
	return err
}

// Images returns all images for a product.
func (r *Repository) Images(ctx context.Context, productID string) ([]Image, error) {
	var imgs []Image
	err := r.db.SelectContext(ctx, &imgs,
		`SELECT id, product_id, url, sort_order FROM product_images WHERE product_id = $1 ORDER BY sort_order`,
		productID,
	)
	return imgs, err
}
