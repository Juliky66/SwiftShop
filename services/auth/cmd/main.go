package main

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/lib/pq"
	"github.com/jmoiron/sqlx"

	"shopkuber/shared/config"
	authserver "shopkuber/auth/internal/server"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	dsn := config.GetString("AUTH_DB_DSN", "postgres://shopkuber:shopkuber@localhost:5432/shopkuber?sslmode=disable")
	port := config.GetString("AUTH_PORT", "8081")
	accessTTL := config.GetDuration("AUTH_JWT_ACCESS_TTL", 15*time.Minute)
	refreshTTL := config.GetDuration("AUTH_JWT_REFRESH_TTL", 30*24*time.Hour)
	privateKeyPath := config.GetString("AUTH_JWT_PRIVATE_KEY_PATH", "./keys/private.pem")
	publicKeyPath := config.GetString("AUTH_JWT_PUBLIC_KEY_PATH", "./keys/public.pem")

	// Connect to database
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		slog.Error("failed to connect to database", "err", err)
		os.Exit(1)
	}
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(5 * time.Minute)
	defer db.Close()

	// Load RSA keys
	privateKey, err := loadPrivateKey(privateKeyPath)
	if err != nil {
		slog.Error("failed to load private key", "err", err)
		os.Exit(1)
	}
	publicKey := &privateKey.PublicKey
	if publicKeyPath != "" {
		if pk, err := loadPublicKey(publicKeyPath); err == nil {
			publicKey = pk
		}
	}

	srv := authserver.New(db, authserver.Config{
		Port:       port,
		PrivateKey: privateKey,
		PublicKey:  publicKey,
		AccessTTL:  accessTTL,
		RefreshTTL: refreshTTL,
	})

	// Graceful shutdown
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
	slog.Info("server stopped")
}

func loadPrivateKey(path string) (*rsa.PrivateKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(data)
	// Try PKCS#1 first (traditional RSA format)
	if key, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
		return key, nil
	}
	// Fall back to PKCS#8 (modern openssl default)
	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	rsaKey, ok := key.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("private key is not RSA")
	}
	return rsaKey, nil
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
