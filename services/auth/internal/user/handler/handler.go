package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-playground/validator/v10"

	apierrors "shopkuber/shared/errors"
	sharedmw "shopkuber/shared/middleware"
	tokensvc "shopkuber/auth/internal/token/service"
	usersvc "shopkuber/auth/internal/user/service"
)

var validate = validator.New()

// UserService is the interface the handler depends on.
type UserService interface {
	Register(ctx context.Context, req usersvc.RegisterRequest) (*usersvc.UserResponse, *tokensvc.TokenPair, error)
	Login(ctx context.Context, req usersvc.LoginRequest) (*usersvc.UserResponse, *tokensvc.TokenPair, error)
	Refresh(ctx context.Context, rawToken string) (*tokensvc.TokenPair, error)
	Logout(ctx context.Context, rawToken string) error
	Me(ctx context.Context, userID string) (*usersvc.UserResponse, error)
	UpdateProfile(ctx context.Context, userID, fullName, phone string) (*usersvc.UserResponse, error)
	ChangePassword(ctx context.Context, userID, oldPassword, newPassword string) error
}

// Handler handles HTTP requests for user auth endpoints.
type Handler struct {
	svc UserService
}

// New creates a new auth handler.
func New(svc UserService) *Handler {
	return &Handler{svc: svc}
}

// Register handles POST /api/v1/auth/register
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"    validate:"required,email"`
		Phone    string `json:"phone"`
		Password string `json:"password" validate:"required,min=8"`
		FullName string `json:"full_name" validate:"required"`
	}
	if !bind(w, r, &req) {
		return
	}

	user, pair, err := h.svc.Register(r.Context(), usersvc.RegisterRequest{
		Email:    req.Email,
		Phone:    req.Phone,
		Password: req.Password,
		FullName: req.FullName,
	})
	if err != nil {
		apierrors.Respond(w, err)
		return
	}

	respond(w, http.StatusCreated, map[string]any{
		"user":          user,
		"access_token":  pair.AccessToken,
		"refresh_token": pair.RefreshToken,
		"expires_at":    pair.ExpiresAt,
	})
}

// Login handles POST /api/v1/auth/login
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"    validate:"required,email"`
		Password string `json:"password" validate:"required"`
	}
	if !bind(w, r, &req) {
		return
	}

	user, pair, err := h.svc.Login(r.Context(), usersvc.LoginRequest{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		apierrors.Respond(w, err)
		return
	}

	respond(w, http.StatusOK, map[string]any{
		"user":          user,
		"access_token":  pair.AccessToken,
		"refresh_token": pair.RefreshToken,
		"expires_at":    pair.ExpiresAt,
	})
}

// Refresh handles POST /api/v1/auth/refresh
func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RefreshToken string `json:"refresh_token" validate:"required"`
	}
	if !bind(w, r, &req) {
		return
	}

	pair, err := h.svc.Refresh(r.Context(), req.RefreshToken)
	if err != nil {
		apierrors.Respond(w, err)
		return
	}

	respond(w, http.StatusOK, map[string]any{
		"access_token":  pair.AccessToken,
		"refresh_token": pair.RefreshToken,
		"expires_at":    pair.ExpiresAt,
	})
}

// Logout handles POST /api/v1/auth/logout
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RefreshToken string `json:"refresh_token" validate:"required"`
	}
	if !bind(w, r, &req) {
		return
	}

	if err := h.svc.Logout(r.Context(), req.RefreshToken); err != nil {
		apierrors.Respond(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Me handles GET /api/v1/auth/me
func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	claims, ok := sharedmw.ClaimsFromContext(r.Context())
	if !ok {
		apierrors.Respond(w, apierrors.ErrUnauthorized)
		return
	}

	user, err := h.svc.Me(r.Context(), claims.UserID)
	if err != nil {
		apierrors.Respond(w, err)
		return
	}

	respond(w, http.StatusOK, user)
}

// UpdateMe handles PUT /api/v1/auth/me
func (h *Handler) UpdateMe(w http.ResponseWriter, r *http.Request) {
	claims, ok := sharedmw.ClaimsFromContext(r.Context())
	if !ok {
		apierrors.Respond(w, apierrors.ErrUnauthorized)
		return
	}

	var req struct {
		FullName string `json:"full_name" validate:"required"`
		Phone    string `json:"phone"`
	}
	if !bind(w, r, &req) {
		return
	}

	user, err := h.svc.UpdateProfile(r.Context(), claims.UserID, req.FullName, req.Phone)
	if err != nil {
		apierrors.Respond(w, err)
		return
	}

	respond(w, http.StatusOK, user)
}

// ChangePassword handles PUT /api/v1/auth/me/password
func (h *Handler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	claims, ok := sharedmw.ClaimsFromContext(r.Context())
	if !ok {
		apierrors.Respond(w, apierrors.ErrUnauthorized)
		return
	}

	var req struct {
		OldPassword string `json:"old_password" validate:"required"`
		NewPassword string `json:"new_password" validate:"required,min=8"`
	}
	if !bind(w, r, &req) {
		return
	}

	if err := h.svc.ChangePassword(r.Context(), claims.UserID, req.OldPassword, req.NewPassword); err != nil {
		apierrors.Respond(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ── helpers ───────────────────────────────────────────────────────────────────

func bind(w http.ResponseWriter, r *http.Request, dst any) bool {
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		apierrors.Respond(w, apierrors.Wrap(apierrors.ErrBadRequest, "invalid JSON"))
		return false
	}
	if err := validate.Struct(dst); err != nil {
		apierrors.Respond(w, apierrors.Wrap(apierrors.ErrBadRequest, err.Error()))
		return false
	}
	return true
}

func respond(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
