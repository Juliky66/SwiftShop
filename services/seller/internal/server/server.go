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
	prodrepo "shopkuber/seller/internal/product/repository"
	sellerhandler "shopkuber/seller/internal/seller/handler"
	sellerrepo "shopkuber/seller/internal/seller/repository"
	sellersvc "shopkuber/seller/internal/seller/service"
)

// Config holds server configuration.
type Config struct {
	Port      string
	PublicKey *rsa.PublicKey
}

// Server wraps the HTTP server.
type Server struct {
	httpServer *http.Server
}

// New wires all dependencies and builds the HTTP server.
func New(db *sqlx.DB, cfg Config) *Server {
	// Repositories
	sRepo := sellerrepo.New(db)
	pRepo := prodrepo.New(db)

	// Services + handlers
	svc := sellersvc.New(sRepo, pRepo)
	h := sellerhandler.New(svc)

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
	requireSeller := sharedmw.RequireRole("seller")
	requireAdmin := sharedmw.RequireRole("admin")

	r.Route("/api/v1", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware)

			// Any authenticated user can apply to become a seller
			r.Post("/seller/register", h.Register)

			// Seller-only endpoints
			r.Group(func(r chi.Router) {
				r.Use(requireSeller)

				r.Get("/seller/profile", h.GetProfile)
				r.Put("/seller/profile", h.UpdateProfile)

				r.Get("/seller/products", h.ListProducts)
				r.Post("/seller/products", h.CreateProduct)
				r.Get("/seller/products/{id}", h.GetProduct)
				r.Put("/seller/products/{id}", h.UpdateProduct)
				r.Delete("/seller/products/{id}", h.ArchiveProduct)
				r.Post("/seller/products/{id}/publish", h.PublishProduct)

				r.Post("/seller/products/{id}/variants", h.AddVariant)
				r.Put("/seller/products/{id}/variants/{vid}", h.UpdateVariant)
				r.Delete("/seller/products/{id}/variants/{vid}", h.DeleteVariant)

				r.Post("/seller/products/{id}/images", h.AddImage)
				r.Delete("/seller/products/{id}/images/{iid}", h.DeleteImage)
			})

			// Admin-only endpoints
			r.Group(func(r chi.Router) {
				r.Use(requireAdmin)

				r.Get("/admin/sellers", h.AdminListSellers)
				r.Put("/admin/sellers/{id}/approve", h.AdminUpdateSellerStatus("approved"))
				r.Put("/admin/sellers/{id}/block", h.AdminUpdateSellerStatus("blocked"))
			})
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

// Start begins listening.
func (s *Server) Start() error {
	slog.Info("seller service listening", "addr", s.httpServer.Addr)
	return s.httpServer.ListenAndServe()
}

// Shutdown gracefully stops the server.
func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}
