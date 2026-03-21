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
	catalogclient "shopkuber/orders/internal/catalog"
	cartrepo "shopkuber/orders/internal/cart/repository"
	orderhandler "shopkuber/orders/internal/order/handler"
	orderrepo "shopkuber/orders/internal/order/repository"
	ordersvc "shopkuber/orders/internal/order/service"
)

// Config holds server configuration.
type Config struct {
	Port       string
	PublicKey  *rsa.PublicKey
	CatalogURL string
}

// Server wraps the HTTP server.
type Server struct {
	httpServer *http.Server
}

// New wires all dependencies and builds the HTTP server.
func New(db *sqlx.DB, cfg Config) *Server {
	// Repositories
	cRepo := cartrepo.New(db)
	oRepo := orderrepo.New(db)

	// Clients
	catClient := catalogclient.New(cfg.CatalogURL)

	// Services
	svc := ordersvc.New(oRepo, cRepo, catClient)

	// Handlers
	h := orderhandler.New(svc)

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

	r.Route("/api/v1", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware)

			// Cart
			r.Get("/cart", h.GetCart)
			r.Post("/cart/items", h.AddToCart)
			r.Put("/cart/items/{variant_id}", h.UpdateCartItem)
			r.Delete("/cart/items/{variant_id}", h.RemoveFromCart)
			r.Delete("/cart", h.ClearCart)

			// Orders
			r.Post("/orders", h.Checkout)
			r.Get("/orders", h.ListOrders)
			r.Get("/orders/{id}", h.GetOrder)
			r.Post("/orders/{id}/cancel", h.CancelOrder)
		})

		// Internal endpoints (not exposed via ingress)
		r.Route("/internal", func(r chi.Router) {
			r.Get("/orders/verified", internalVerifiedPurchase(svc))
			r.Put("/orders/{id}/status", internalUpdateStatus(svc))
			r.Get("/orders/{id}/amount", internalGetOrderAmount(svc))
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

func internalVerifiedPurchase(svc *ordersvc.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.URL.Query().Get("user_id")
		productID := r.URL.Query().Get("product_id")
		if userID == "" || productID == "" {
			http.Error(w, "user_id and product_id are required", http.StatusBadRequest)
			return
		}
		orderID, err := svc.VerifiedPurchase(r.Context(), userID, productID)
		if err != nil {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"order_id": orderID})
	}
}

func internalGetOrderAmount(svc *ordersvc.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		orderID := chi.URLParam(r, "id")
		amount, err := svc.GetOrderTotalAmount(r.Context(), orderID)
		if err != nil {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]float64{"total_amount": amount})
	}
}

func internalUpdateStatus(svc *ordersvc.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		orderID := chi.URLParam(r, "id")
		var req struct {
			Status string `json:"status"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}
		if err := svc.UpdateStatus(r.Context(), orderID, req.Status); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

// Start begins listening.
func (s *Server) Start() error {
	slog.Info("orders service listening", "addr", s.httpServer.Addr)
	return s.httpServer.ListenAndServe()
}

// Shutdown gracefully stops the server.
func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}
