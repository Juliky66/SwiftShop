package service

import (
	"context"

	"golang.org/x/crypto/bcrypt"

	apierrors "shopkuber/shared/errors"
	userrepo "shopkuber/auth/internal/user/repository"
	tokensvc "shopkuber/auth/internal/token/service"
)

// RegisterRequest holds registration input.
type RegisterRequest struct {
	Email    string
	Phone    string
	Password string
	FullName string
}

// LoginRequest holds login input.
type LoginRequest struct {
	Email    string
	Password string
}

// UserResponse is returned after successful auth operations.
type UserResponse struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	Phone    string `json:"phone,omitempty"`
	FullName string `json:"full_name"`
	Role     string `json:"role"`
}

// Service handles user business logic.
type Service struct {
	repo      *userrepo.Repository
	tokensSvc *tokensvc.Service
}

// New creates a new user service.
func New(repo *userrepo.Repository, tokensSvc *tokensvc.Service) *Service {
	return &Service{repo: repo, tokensSvc: tokensSvc}
}

// Register creates a new user account and returns a token pair.
func (s *Service) Register(ctx context.Context, req RegisterRequest) (*UserResponse, *tokensvc.TokenPair, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, nil, err
	}

	u, err := s.repo.Create(ctx, req.Email, req.Phone, string(hash), req.FullName)
	if err != nil {
		return nil, nil, err
	}

	pair, err := s.tokensSvc.Issue(ctx, u.ID, u.Role)
	if err != nil {
		return nil, nil, err
	}

	return toResponse(u), pair, nil
}

// Login authenticates a user and returns a token pair.
func (s *Service) Login(ctx context.Context, req LoginRequest) (*UserResponse, *tokensvc.TokenPair, error) {
	u, err := s.repo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, nil, apierrors.Wrap(apierrors.ErrUnauthorized, "invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(req.Password)); err != nil {
		return nil, nil, apierrors.Wrap(apierrors.ErrUnauthorized, "invalid credentials")
	}

	if !u.IsActive {
		return nil, nil, apierrors.Wrap(apierrors.ErrForbidden, "account is disabled")
	}

	pair, err := s.tokensSvc.Issue(ctx, u.ID, u.Role)
	if err != nil {
		return nil, nil, err
	}

	return toResponse(u), pair, nil
}

// Refresh rotates a refresh token and issues a new pair.
func (s *Service) Refresh(ctx context.Context, rawToken string) (*tokensvc.TokenPair, error) {
	stored, err := s.tokensSvc.FindByRaw(ctx, rawToken)
	if err != nil {
		return nil, apierrors.ErrUnauthorized
	}

	u, err := s.repo.FindByID(ctx, stored.UserID)
	if err != nil {
		return nil, apierrors.ErrUnauthorized
	}

	return s.tokensSvc.IssueForUser(ctx, rawToken, u.ID, u.Role)
}

// Logout revokes the supplied refresh token.
func (s *Service) Logout(ctx context.Context, rawToken string) error {
	return s.tokensSvc.Revoke(ctx, rawToken)
}

// Me returns the profile of the authenticated user.
func (s *Service) Me(ctx context.Context, userID string) (*UserResponse, error) {
	u, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return toResponse(u), nil
}

// UpdateProfile updates the user's name and phone.
func (s *Service) UpdateProfile(ctx context.Context, userID, fullName, phone string) (*UserResponse, error) {
	u, err := s.repo.UpdateProfile(ctx, userID, fullName, phone)
	if err != nil {
		return nil, err
	}
	return toResponse(u), nil
}

// ChangePassword verifies the old password and sets a new one.
func (s *Service) ChangePassword(ctx context.Context, userID, oldPassword, newPassword string) error {
	u, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		return err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(oldPassword)); err != nil {
		return apierrors.Wrap(apierrors.ErrUnauthorized, "incorrect current password")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	if err := s.repo.UpdatePassword(ctx, userID, string(hash)); err != nil {
		return err
	}

	// Invalidate all sessions after password change
	return s.tokensSvc.RevokeAll(ctx, userID)
}

func toResponse(u *userrepo.User) *UserResponse {
	resp := &UserResponse{
		ID:       u.ID,
		Email:    u.Email,
		FullName: u.FullName,
		Role:     u.Role,
	}
	if u.Phone != nil {
		resp.Phone = *u.Phone
	}
	return resp
}
