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
	catalogclient "shopkuber/reviews/internal/catalog"
	ordersclient "shopkuber/reviews/internal/orders"
	"shopkuber/reviews/internal/payment/gateway"
	paymenthandler "shopkuber/reviews/internal/payment/handler"
	paymentrepo "shopkuber/reviews/internal/payment/repository"
	paymentsvc "shopkuber/reviews/internal/payment/service"
	reviewhandler "shopkuber/reviews/internal/review/handler"
	reviewrepo "shopkuber/reviews/internal/review/repository"
	reviewsvc "shopkuber/reviews/internal/review/service"
)

// Config holds server configuration.
type Config struct {
	Port            string
	PublicKey       *rsa.PublicKey
	OrdersURL       string
	CatalogURL      string
	PaymentProvider string
	WebhookSecret   string
}

// Server wraps the HTTP server.
type Server struct {
	httpServer *http.Server
}

// New wires all dependencies and builds the HTTP server.
func New(db *sqlx.DB, cfg Config) *Server {
	// Clients
	ordersClient := ordersclient.New(cfg.OrdersURL)
	catalogClient := catalogclient.New(cfg.CatalogURL)

	// Repositories
	rRepo := reviewrepo.New(db)
	pRepo := paymentrepo.New(db)

	// Payment gateway
	var gw gateway.Gateway
	gw = &gateway.MockGateway{}

	// Services
	rSvc := reviewsvc.New(rRepo, ordersClient, catalogClient)
	pSvc := paymentsvc.New(pRepo, gw, ordersClient, cfg.WebhookSecret, cfg.PaymentProvider)

	// Handlers
	rH := reviewhandler.New(rSvc)
	pH := paymenthandler.New(pSvc)

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
		// Public: list reviews
		r.Get("/products/{id}/reviews", rH.List)

		// Payment webhook (no auth, HMAC verified inside service)
		r.Post("/payments/webhook", pH.Webhook)

		r.Group(func(r chi.Router) {
			r.Use(authMiddleware)

			// Reviews
			r.Post("/products/{id}/reviews", rH.Submit)
			r.Put("/reviews/{id}", rH.Update)
			r.Delete("/reviews/{id}", rH.Delete)

			// Payments
			r.Post("/orders/{id}/pay", pH.Initiate)
			r.Get("/payments/{id}", pH.GetStatus)
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
	slog.Info("reviews service listening", "addr", s.httpServer.Addr)
	return s.httpServer.ListenAndServe()
}

// Shutdown gracefully stops the server.
func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}
