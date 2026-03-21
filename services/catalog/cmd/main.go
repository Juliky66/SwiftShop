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
	catalogserver "shopkuber/catalog/internal/server"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	dsn := config.GetString("CATALOG_DB_DSN", "postgres://shopkuber:shopkuber@localhost:5432/shopkuber?sslmode=disable")
	port := config.GetString("CATALOG_PORT", "8082")
	publicKeyPath := config.GetString("CATALOG_JWT_PUBLIC_KEY_PATH", "./keys/public.pem")

	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		slog.Error("failed to connect to database", "err", err)
		os.Exit(1)
	}
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(5 * time.Minute)
	defer db.Close()

	publicKey, err := loadPublicKey(publicKeyPath)
	if err != nil {
		slog.Error("failed to load public key", "err", err)
		os.Exit(1)
	}

	srv := catalogserver.New(db, catalogserver.Config{
		Port:      port,
		PublicKey: publicKey,
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
	slog.Info("shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("shutdown error", "err", err)
	}
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
