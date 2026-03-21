package server

import (
	"context"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/jmoiron/sqlx"

	sharedmw "shopkuber/shared/middleware"
	cathandler "shopkuber/catalog/internal/category/handler"
	catrepo "shopkuber/catalog/internal/category/repository"
	catsvc "shopkuber/catalog/internal/category/service"
	prodhandler "shopkuber/catalog/internal/product/handler"
	prodrepo "shopkuber/catalog/internal/product/repository"
	prodsvc "shopkuber/catalog/internal/product/service"
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
	cRepo := catrepo.New(db)
	pRepo := prodrepo.New(db)

	// Services
	catService := catsvc.New(cRepo)
	prodService := prodsvc.New(pRepo, cRepo)

	// Handlers
	ch := cathandler.New(catService)
	ph := prodhandler.New(prodService, catService)

	r := chi.NewRouter()
	r.Use(chiMiddleware.Recoverer)
	r.Use(sharedmw.Logger)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "OPTIONS"},
		AllowedHeaders: []string{"Accept", "Authorization", "Content-Type"},
	}))

	r.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	r.Route("/api/v1", func(r chi.Router) {
		// Public catalog routes
		r.Get("/categories", ch.List)
		r.Get("/categories/{slug}", ch.Get)
		r.Get("/categories/{slug}/products", ph.CategoryProducts)

		r.Get("/products", ph.List)
		r.Get("/products/search", ph.Search)
		r.Get("/products/{slug}", ph.Get)

		// Internal routes (not exposed via ingress, called by other services)
		r.Route("/internal", func(r chi.Router) {
			r.Post("/variants/{id}/reserve", internalReserve(prodService))
			r.Put("/products/{id}/rating", internalUpdateRating(prodService))
			r.Get("/variants/{id}", internalGetVariant(prodService))
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

func internalReserve(svc *prodsvc.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		variantID := chi.URLParam(r, "id")
		var req struct {
			Quantity int `json:"quantity"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}
		if err := svc.ReserveStock(r.Context(), variantID, req.Quantity); err != nil {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func internalUpdateRating(svc *prodsvc.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		productID := chi.URLParam(r, "id")
		var req struct {
			Rating      float64 `json:"rating"`
			ReviewCount int     `json:"review_count"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}
		if err := svc.UpdateRating(r.Context(), productID, req.Rating, req.ReviewCount); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func internalGetVariant(svc *prodsvc.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		variantID := chi.URLParam(r, "id")
		v, p, err := svc.GetVariant(r.Context(), variantID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"variant": v, "product": p})
	}
}

// Start begins listening.
func (s *Server) Start() error {
	slog.Info("catalog service listening", "addr", s.httpServer.Addr)
	return s.httpServer.ListenAndServe()
}

// Shutdown gracefully stops the server.
func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}
