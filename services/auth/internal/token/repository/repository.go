package repository

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"

	apierrors "shopkuber/shared/errors"
)

// RefreshToken is the database model.
type RefreshToken struct {
	ID        string    `db:"id"`
	UserID    string    `db:"user_id"`
	TokenHash string    `db:"token_hash"`
	ExpiresAt time.Time `db:"expires_at"`
	CreatedAt time.Time `db:"created_at"`
}

// Repository handles refresh token persistence.
type Repository struct {
	db *sqlx.DB
}

// New creates a new token repository.
func New(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

// Save stores a new refresh token (hashed).
func (r *Repository) Save(ctx context.Context, userID, rawToken string, expiresAt time.Time) error {
	hash := hashToken(rawToken)
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO refresh_tokens (user_id, token_hash, expires_at) VALUES ($1, $2, $3)`,
		userID, hash, expiresAt,
	)
	return err
}

// Find looks up a refresh token by its raw value.
func (r *Repository) Find(ctx context.Context, rawToken string) (*RefreshToken, error) {
	hash := hashToken(rawToken)
	t := &RefreshToken{}
	err := r.db.QueryRowxContext(ctx,
		`SELECT id, user_id, token_hash, expires_at, created_at FROM refresh_tokens WHERE token_hash = $1`,
		hash,
	).StructScan(t)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apierrors.ErrNotFound
		}
		return nil, err
	}
	return t, nil
}

// Delete removes a refresh token by its raw value (logout).
func (r *Repository) Delete(ctx context.Context, rawToken string) error {
	hash := hashToken(rawToken)
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM refresh_tokens WHERE token_hash = $1`, hash,
	)
	return err
}

// DeleteAllForUser removes all refresh tokens for a user (e.g., on password change).
func (r *Repository) DeleteAllForUser(ctx context.Context, userID string) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM refresh_tokens WHERE user_id = $1`, userID,
	)
	return err
}

func hashToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return fmt.Sprintf("%x", sum)
}
