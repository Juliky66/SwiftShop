package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"

	apierrors "shopkuber/shared/errors"
)

// User is the database model for a user record.
type User struct {
	ID           string  `db:"id"            json:"id"`
	Email        string  `db:"email"         json:"email"`
	Phone        *string `db:"phone"         json:"phone"`
	PasswordHash string  `db:"password_hash" json:"-"`
	FullName     string  `db:"full_name"     json:"full_name"`
	Role         string  `db:"role"          json:"role"`
	IsActive     bool    `db:"is_active"     json:"is_active"`
}

// Repository handles user persistence.
type Repository struct {
	db *sqlx.DB
}

// New creates a new user repository.
func New(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

// Create inserts a new user record and returns the created user.
func (r *Repository) Create(ctx context.Context, email, phone, passwordHash, fullName string) (*User, error) {
	u := &User{}
	query := `
		INSERT INTO users (email, phone, password_hash, full_name)
		VALUES ($1, NULLIF($2, ''), $3, $4)
		RETURNING id, email, phone, password_hash, full_name, role, is_active`
	err := r.db.QueryRowxContext(ctx, query, email, phone, passwordHash, fullName).StructScan(u)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, apierrors.Wrap(apierrors.ErrConflict, "email already registered")
		}
		return nil, err
	}
	return u, nil
}

// FindByEmail looks up a user by email address.
func (r *Repository) FindByEmail(ctx context.Context, email string) (*User, error) {
	u := &User{}
	err := r.db.QueryRowxContext(ctx,
		`SELECT id, email, phone, password_hash, full_name, role, is_active FROM users WHERE email = $1`,
		email,
	).StructScan(u)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apierrors.ErrNotFound
		}
		return nil, err
	}
	return u, nil
}

// FindByID looks up a user by primary key.
func (r *Repository) FindByID(ctx context.Context, id string) (*User, error) {
	u := &User{}
	err := r.db.QueryRowxContext(ctx,
		`SELECT id, email, phone, password_hash, full_name, role, is_active FROM users WHERE id = $1`,
		id,
	).StructScan(u)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apierrors.ErrNotFound
		}
		return nil, err
	}
	return u, nil
}

// UpdateProfile updates the full_name and phone for a user.
func (r *Repository) UpdateProfile(ctx context.Context, id, fullName, phone string) (*User, error) {
	u := &User{}
	err := r.db.QueryRowxContext(ctx,
		`UPDATE users SET full_name = $1, phone = NULLIF($2,''), updated_at = now()
		 WHERE id = $3
		 RETURNING id, email, phone, password_hash, full_name, role, is_active`,
		fullName, phone, id,
	).StructScan(u)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apierrors.ErrNotFound
		}
		return nil, err
	}
	return u, nil
}

// UpdatePassword changes a user's password hash.
func (r *Repository) UpdatePassword(ctx context.Context, id, newHash string) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE users SET password_hash = $1, updated_at = now() WHERE id = $2`,
		newHash, id,
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

// UpdateRole changes a user's role (used by seller service via internal call).
func (r *Repository) UpdateRole(ctx context.Context, id, role string) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE users SET role = $1, updated_at = now() WHERE id = $2`,
		role, id,
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

func isUniqueViolation(err error) bool {
	return err != nil && (err.Error() == `pq: duplicate key value violates unique constraint "users_email_key"` ||
		contains(err.Error(), "duplicate key"))
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(s) > 0 && containsStr(s, sub))
}

func containsStr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
