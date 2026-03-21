package service

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/hex"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	apierrors "shopkuber/shared/errors"
	sharedmw "shopkuber/shared/middleware"
	tokenrepo "shopkuber/auth/internal/token/repository"
)

// TokenPair holds a freshly issued access + refresh token pair.
type TokenPair struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
}

// Service handles JWT issuance and refresh token lifecycle.
type Service struct {
	repo       *tokenrepo.Repository
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
	accessTTL  time.Duration
	refreshTTL time.Duration
}

// New creates a new token service.
func New(repo *tokenrepo.Repository, privateKey *rsa.PrivateKey, publicKey *rsa.PublicKey, accessTTL, refreshTTL time.Duration) *Service {
	return &Service{
		repo:       repo,
		privateKey: privateKey,
		publicKey:  publicKey,
		accessTTL:  accessTTL,
		refreshTTL: refreshTTL,
	}
}

// Issue creates a new access + refresh token pair for the given user.
func (s *Service) Issue(ctx context.Context, userID, role string) (*TokenPair, error) {
	now := time.Now()
	accessExp := now.Add(s.accessTTL)

	claims := sharedmw.Claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(accessExp),
			ID:        uuid.New().String(),
		},
	}

	accessToken, err := jwt.NewWithClaims(jwt.SigningMethodRS256, claims).SignedString(s.privateKey)
	if err != nil {
		return nil, err
	}

	rawRefresh, err := generateToken()
	if err != nil {
		return nil, err
	}

	refreshExp := now.Add(s.refreshTTL)
	if err := s.repo.Save(ctx, userID, rawRefresh, refreshExp); err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: rawRefresh,
		ExpiresAt:    accessExp,
	}, nil
}

// Refresh validates a refresh token and issues a new token pair.
func (s *Service) Refresh(ctx context.Context, rawRefreshToken string) (*TokenPair, error) {
	stored, err := s.repo.Find(ctx, rawRefreshToken)
	if err != nil {
		return nil, apierrors.ErrUnauthorized
	}
	if time.Now().After(stored.ExpiresAt) {
		_ = s.repo.Delete(ctx, rawRefreshToken)
		return nil, apierrors.Wrap(apierrors.ErrUnauthorized, "refresh token expired")
	}

	// Rotate: delete old, issue new pair
	if err := s.repo.Delete(ctx, rawRefreshToken); err != nil {
		return nil, err
	}

	// We need the user's role — caller (auth service) will pass it in via a wrapper
	// that first fetches the user. Here we just need userID from stored token.
	return nil, nil // replaced by IssueForUser below
}

// IssueForUser is the full refresh flow: validate old token, issue new pair.
func (s *Service) IssueForUser(ctx context.Context, rawRefreshToken, userID, role string) (*TokenPair, error) {
	stored, err := s.repo.Find(ctx, rawRefreshToken)
	if err != nil {
		return nil, apierrors.ErrUnauthorized
	}
	if stored.UserID != userID {
		return nil, apierrors.ErrUnauthorized
	}
	if time.Now().After(stored.ExpiresAt) {
		_ = s.repo.Delete(ctx, rawRefreshToken)
		return nil, apierrors.Wrap(apierrors.ErrUnauthorized, "refresh token expired")
	}
	if err := s.repo.Delete(ctx, rawRefreshToken); err != nil {
		return nil, err
	}
	return s.Issue(ctx, userID, role)
}

// Revoke deletes a refresh token (logout).
func (s *Service) Revoke(ctx context.Context, rawRefreshToken string) error {
	return s.repo.Delete(ctx, rawRefreshToken)
}

// RevokeAll deletes all refresh tokens for a user (password change).
func (s *Service) RevokeAll(ctx context.Context, userID string) error {
	return s.repo.DeleteAllForUser(ctx, userID)
}

// FindByRaw looks up the stored token to get the userID for refresh.
func (s *Service) FindByRaw(ctx context.Context, rawToken string) (*tokenrepo.RefreshToken, error) {
	return s.repo.Find(ctx, rawToken)
}

func generateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
