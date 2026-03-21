package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	apierrors "shopkuber/shared/errors"
	sharedmw "shopkuber/shared/middleware"
	tokensvc "shopkuber/auth/internal/token/service"
	"shopkuber/auth/internal/user/handler"
	usersvc "shopkuber/auth/internal/user/service"
)

// ── mock ──────────────────────────────────────────────────────────────────────

type mockUserSvc struct {
	registerFn       func(context.Context, usersvc.RegisterRequest) (*usersvc.UserResponse, *tokensvc.TokenPair, error)
	loginFn          func(context.Context, usersvc.LoginRequest) (*usersvc.UserResponse, *tokensvc.TokenPair, error)
	refreshFn        func(context.Context, string) (*tokensvc.TokenPair, error)
	logoutFn         func(context.Context, string) error
	meFn             func(context.Context, string) (*usersvc.UserResponse, error)
	updateProfileFn  func(context.Context, string, string, string) (*usersvc.UserResponse, error)
	changePasswordFn func(context.Context, string, string, string) error
}

func (m *mockUserSvc) Register(ctx context.Context, req usersvc.RegisterRequest) (*usersvc.UserResponse, *tokensvc.TokenPair, error) {
	return m.registerFn(ctx, req)
}
func (m *mockUserSvc) Login(ctx context.Context, req usersvc.LoginRequest) (*usersvc.UserResponse, *tokensvc.TokenPair, error) {
	return m.loginFn(ctx, req)
}
func (m *mockUserSvc) Refresh(ctx context.Context, raw string) (*tokensvc.TokenPair, error) {
	return m.refreshFn(ctx, raw)
}
func (m *mockUserSvc) Logout(ctx context.Context, raw string) error {
	return m.logoutFn(ctx, raw)
}
func (m *mockUserSvc) Me(ctx context.Context, userID string) (*usersvc.UserResponse, error) {
	return m.meFn(ctx, userID)
}
func (m *mockUserSvc) UpdateProfile(ctx context.Context, userID, fullName, phone string) (*usersvc.UserResponse, error) {
	return m.updateProfileFn(ctx, userID, fullName, phone)
}
func (m *mockUserSvc) ChangePassword(ctx context.Context, userID, old, newPwd string) error {
	return m.changePasswordFn(ctx, userID, old, newPwd)
}

// ── helpers ───────────────────────────────────────────────────────────────────

func fakePair() *tokensvc.TokenPair {
	return &tokensvc.TokenPair{
		AccessToken:  "access-token",
		RefreshToken: "refresh-token",
		ExpiresAt:    time.Now().Add(time.Hour),
	}
}

func fakeUser() *usersvc.UserResponse {
	return &usersvc.UserResponse{ID: "user-1", Email: "alice@example.com", FullName: "Alice", Role: "buyer"}
}

func jsonBody(t *testing.T, v any) *bytes.Buffer {
	t.Helper()
	b, err := json.Marshal(v)
	require.NoError(t, err)
	return bytes.NewBuffer(b)
}

func withClaims(r *http.Request, userID string) *http.Request {
	ctx := context.WithValue(r.Context(), sharedmw.ClaimsKey, &sharedmw.Claims{UserID: userID, Role: "buyer"})
	return r.WithContext(ctx)
}

func callHandler(fn http.HandlerFunc, req *http.Request) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	fn(rec, req)
	return rec
}

// ── Register ──────────────────────────────────────────────────────────────────

func TestRegister_InvalidJSON(t *testing.T) {
	h := handler.New(&mockUserSvc{})
	rec := callHandler(h.Register, httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString("{bad")))
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestRegister_MissingFields(t *testing.T) {
	h := handler.New(&mockUserSvc{})
	// email is missing
	body := jsonBody(t, map[string]string{"password": "secret123", "full_name": "Alice"})
	rec := callHandler(h.Register, httptest.NewRequest(http.MethodPost, "/", body))
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestRegister_PasswordTooShort(t *testing.T) {
	h := handler.New(&mockUserSvc{})
	body := jsonBody(t, map[string]string{"email": "a@b.com", "password": "short", "full_name": "Alice"})
	rec := callHandler(h.Register, httptest.NewRequest(http.MethodPost, "/", body))
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestRegister_Success(t *testing.T) {
	svc := &mockUserSvc{
		registerFn: func(_ context.Context, _ usersvc.RegisterRequest) (*usersvc.UserResponse, *tokensvc.TokenPair, error) {
			return fakeUser(), fakePair(), nil
		},
	}
	h := handler.New(svc)

	body := jsonBody(t, map[string]string{"email": "alice@example.com", "password": "password123", "full_name": "Alice"})
	rec := callHandler(h.Register, httptest.NewRequest(http.MethodPost, "/", body))

	assert.Equal(t, http.StatusCreated, rec.Code)
	var resp map[string]any
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
	assert.Contains(t, resp, "access_token")
	assert.Contains(t, resp, "refresh_token")
}

func TestRegister_EmailConflict(t *testing.T) {
	svc := &mockUserSvc{
		registerFn: func(_ context.Context, _ usersvc.RegisterRequest) (*usersvc.UserResponse, *tokensvc.TokenPair, error) {
			return nil, nil, apierrors.ErrConflict
		},
	}
	h := handler.New(svc)

	body := jsonBody(t, map[string]string{"email": "taken@example.com", "password": "password123", "full_name": "Alice"})
	rec := callHandler(h.Register, httptest.NewRequest(http.MethodPost, "/", body))

	assert.Equal(t, http.StatusConflict, rec.Code)
}

// ── Login ─────────────────────────────────────────────────────────────────────

func TestLogin_InvalidJSON(t *testing.T) {
	h := handler.New(&mockUserSvc{})
	rec := callHandler(h.Login, httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString("{bad")))
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestLogin_Success(t *testing.T) {
	svc := &mockUserSvc{
		loginFn: func(_ context.Context, _ usersvc.LoginRequest) (*usersvc.UserResponse, *tokensvc.TokenPair, error) {
			return fakeUser(), fakePair(), nil
		},
	}
	h := handler.New(svc)

	body := jsonBody(t, map[string]string{"email": "alice@example.com", "password": "password123"})
	rec := callHandler(h.Login, httptest.NewRequest(http.MethodPost, "/", body))

	assert.Equal(t, http.StatusOK, rec.Code)
	var resp map[string]any
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
	assert.Contains(t, resp, "access_token")
}

func TestLogin_InvalidCredentials(t *testing.T) {
	svc := &mockUserSvc{
		loginFn: func(_ context.Context, _ usersvc.LoginRequest) (*usersvc.UserResponse, *tokensvc.TokenPair, error) {
			return nil, nil, apierrors.ErrUnauthorized
		},
	}
	h := handler.New(svc)

	body := jsonBody(t, map[string]string{"email": "alice@example.com", "password": "wrongpass"})
	rec := callHandler(h.Login, httptest.NewRequest(http.MethodPost, "/", body))

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

// ── Logout ────────────────────────────────────────────────────────────────────

func TestLogout_MissingToken(t *testing.T) {
	h := handler.New(&mockUserSvc{})
	body := jsonBody(t, map[string]string{}) // refresh_token missing
	rec := callHandler(h.Logout, httptest.NewRequest(http.MethodPost, "/", body))
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestLogout_Success(t *testing.T) {
	svc := &mockUserSvc{
		logoutFn: func(_ context.Context, _ string) error { return nil },
	}
	h := handler.New(svc)

	body := jsonBody(t, map[string]string{"refresh_token": "tok"})
	rec := callHandler(h.Logout, httptest.NewRequest(http.MethodPost, "/", body))

	assert.Equal(t, http.StatusNoContent, rec.Code)
}

// ── Me (protected) ────────────────────────────────────────────────────────────

func TestMe_NoAuthContext(t *testing.T) {
	h := handler.New(&mockUserSvc{})
	// no claims injected into context
	rec := callHandler(h.Me, httptest.NewRequest(http.MethodGet, "/", nil))
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestMe_Success(t *testing.T) {
	svc := &mockUserSvc{
		meFn: func(_ context.Context, _ string) (*usersvc.UserResponse, error) {
			return fakeUser(), nil
		},
	}
	h := handler.New(svc)

	req := withClaims(httptest.NewRequest(http.MethodGet, "/", nil), "user-1")
	rec := callHandler(h.Me, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var resp usersvc.UserResponse
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
	assert.Equal(t, "user-1", resp.ID)
}

// ── ChangePassword (protected) ────────────────────────────────────────────────

func TestChangePassword_NoAuthContext(t *testing.T) {
	h := handler.New(&mockUserSvc{})
	rec := callHandler(h.ChangePassword, httptest.NewRequest(http.MethodPut, "/", nil))
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestChangePassword_Success(t *testing.T) {
	svc := &mockUserSvc{
		changePasswordFn: func(_ context.Context, _, _, _ string) error { return nil },
	}
	h := handler.New(svc)

	body := jsonBody(t, map[string]string{"old_password": "oldpass1", "new_password": "newpass1"})
	req := withClaims(httptest.NewRequest(http.MethodPut, "/", body), "user-1")
	rec := callHandler(h.ChangePassword, req)

	assert.Equal(t, http.StatusNoContent, rec.Code)
}

func TestChangePassword_WrongOldPassword(t *testing.T) {
	svc := &mockUserSvc{
		changePasswordFn: func(_ context.Context, _, _, _ string) error {
			return apierrors.ErrUnauthorized
		},
	}
	h := handler.New(svc)

	body := jsonBody(t, map[string]string{"old_password": "wrong", "new_password": "newpass1"})
	req := withClaims(httptest.NewRequest(http.MethodPut, "/", body), "user-1")
	rec := callHandler(h.ChangePassword, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

// ── UpdateMe (protected) ──────────────────────────────────────────────────────

func TestUpdateMe_NoAuthContext(t *testing.T) {
	h := handler.New(&mockUserSvc{})
	rec := callHandler(h.UpdateMe, httptest.NewRequest(http.MethodPut, "/", nil))
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestUpdateMe_MissingFullName(t *testing.T) {
	h := handler.New(&mockUserSvc{})
	body := jsonBody(t, map[string]string{"phone": "+7999"}) // full_name required
	req := withClaims(httptest.NewRequest(http.MethodPut, "/", body), "user-1")
	rec := callHandler(h.UpdateMe, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}
