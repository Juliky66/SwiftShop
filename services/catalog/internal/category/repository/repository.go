package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"

	apierrors "shopkuber/shared/errors"
)

// Category is the DB model for a category.
type Category struct {
	ID        string  `db:"id"         json:"id"`
	ParentID  *string `db:"parent_id"  json:"parent_id"`
	Name      string  `db:"name"       json:"name"`
	Slug      string  `db:"slug"       json:"slug"`
	SortOrder int     `db:"sort_order" json:"sort_order"`
}

// CategoryNode includes its children for tree responses.
type CategoryNode struct {
	Category
	Children []CategoryNode `json:"children,omitempty"`
}

// Repository handles category persistence.
type Repository struct {
	db *sqlx.DB
}

// New creates a new category repository.
func New(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

// Tree returns all categories in flat list ordered for tree construction.
func (r *Repository) Tree(ctx context.Context) ([]Category, error) {
	var cats []Category
	err := r.db.SelectContext(ctx, &cats,
		`WITH RECURSIVE tree AS (
			SELECT id, parent_id, name, slug, sort_order FROM categories WHERE parent_id IS NULL
			UNION ALL
			SELECT c.id, c.parent_id, c.name, c.slug, c.sort_order
			FROM categories c
			INNER JOIN tree t ON c.parent_id = t.id
		)
		SELECT * FROM tree ORDER BY sort_order, name`)
	return cats, err
}

// FindBySlug finds a single category by slug.
func (r *Repository) FindBySlug(ctx context.Context, slug string) (*Category, error) {
	c := &Category{}
	err := r.db.QueryRowxContext(ctx,
		`SELECT id, parent_id, name, slug, sort_order FROM categories WHERE slug = $1`, slug,
	).StructScan(c)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apierrors.ErrNotFound
		}
		return nil, err
	}
	return c, nil
}

// Children returns immediate children of a category.
func (r *Repository) Children(ctx context.Context, parentID string) ([]Category, error) {
	var cats []Category
	err := r.db.SelectContext(ctx, &cats,
		`SELECT id, parent_id, name, slug, sort_order FROM categories WHERE parent_id = $1 ORDER BY sort_order, name`,
		parentID,
	)
	return cats, err
}
