package server

import (
	"context"
	"crypto/rsa"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/jmoiron/sqlx"

	sharedmw "shopkuber/shared/middleware"
	tokenrepo "shopkuber/auth/internal/token/repository"
	tokensvc "shopkuber/auth/internal/token/service"
	userhandler "shopkuber/auth/internal/user/handler"
	userrepo "shopkuber/auth/internal/user/repository"
	usersvc "shopkuber/auth/internal/user/service"
)

// Config holds server configuration.
type Config struct {
	Port       string
	PrivateKey *rsa.PrivateKey
	PublicKey  *rsa.PublicKey
	AccessTTL  time.Duration
	RefreshTTL time.Duration
}

// Server is the HTTP server for the auth service.
type Server struct {
	httpServer *http.Server
}

// New wires all dependencies and builds the HTTP server.
func New(db *sqlx.DB, cfg Config) *Server {
	// Repositories
	tRepo := tokenrepo.New(db)
	uRepo := userrepo.New(db)

	// Services
	tSvc := tokensvc.New(tRepo, cfg.PrivateKey, cfg.PublicKey, cfg.AccessTTL, cfg.RefreshTTL)
	uSvc := usersvc.New(uRepo, tSvc)

	// Handlers
	h := userhandler.New(uSvc)

	// Router
	r := chi.NewRouter()
	r.Use(chiMiddleware.Recoverer)
	r.Use(sharedmw.Logger)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Accept", "Authorization", "Content-Type"},
	}))

	r.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	authMiddleware := sharedmw.Authenticate(cfg.PublicKey)

	r.Route("/api/v1/auth", func(r chi.Router) {
		r.Post("/register", h.Register)
		r.Post("/login", h.Login)
		r.Post("/refresh", h.Refresh)

		r.Group(func(r chi.Router) {
			r.Use(authMiddleware)
			r.Post("/logout", h.Logout)
			r.Get("/me", h.Me)
			r.Put("/me", h.UpdateMe)
			r.Put("/me/password", h.ChangePassword)
		})
	})

	return &Server{
		httpServer: &http.Server{
			Addr:         fmt.Sprintf(":%s", cfg.Port),
			Handler:      r,
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
			IdleTimeout:  60 * time.Second,
		},
	}
}

// Start begins listening for connections.
func (s *Server) Start() error {
	slog.Info("auth service listening", "addr", s.httpServer.Addr)
	return s.httpServer.ListenAndServe()
}

// Shutdown gracefully stops the server.
func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}
