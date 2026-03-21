package main

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/lib/pq"
	"github.com/jmoiron/sqlx"

	"shopkuber/shared/config"
	ordersserver "shopkuber/orders/internal/server"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	db, err := sqlx.Connect("postgres", config.GetString("ORDERS_DB_DSN", "postgres://shopkuber:shopkuber@localhost:5432/shopkuber?sslmode=disable"))
	if err != nil {
		slog.Error("db connect failed", "err", err)
		os.Exit(1)
	}
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(5 * time.Minute)
	defer db.Close()

	publicKey, err := loadPublicKey(config.GetString("ORDERS_JWT_PUBLIC_KEY_PATH", "./keys/public.pem"))
	if err != nil {
		slog.Error("failed to load public key", "err", err)
		os.Exit(1)
	}

	srv := ordersserver.New(db, ordersserver.Config{
		Port:       config.GetString("ORDERS_PORT", "8083"),
		PublicKey:  publicKey,
		CatalogURL: config.GetString("ORDERS_CATALOG_URL", "http://localhost:8082"),
	})

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := srv.Start(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "err", err)
			os.Exit(1)
		}
	}()

	<-quit
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctx)
	slog.Info("server stopped")
}

func loadPublicKey(path string) (*rsa.PublicKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(data)
	key, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	return key.(*rsa.PublicKey), nil
}
