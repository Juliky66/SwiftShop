package service

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"

	apierrors "shopkuber/shared/errors"
	"shopkuber/reviews/internal/payment/gateway"
)

// newTestService builds a Service with nil repo/orders (safe for tests that
// only exercise logic that does NOT touch the database).
func newTestService(secret string) *Service {
	return New(nil, &gateway.MockGateway{}, nil, secret, "mock")
}

// ── mapProviderStatus ─────────────────────────────────────────────────────────

func TestMapProviderStatus(t *testing.T) {
	cases := []struct {
		input    string
		expected string
	}{
		{"succeeded", "succeeded"},
		{"success", "succeeded"},
		{"paid", "succeeded"},
		{"failed", "failed"},
		{"cancelled", "failed"},
		{"canceled", "failed"},
		{"refunded", "refunded"},
		{"pending", "pending"},
		{"unknown", "pending"},
		{"", "pending"},
	}

	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			assert.Equal(t, tc.expected, mapProviderStatus(tc.input))
		})
	}
}

// ── verifySignature / HandleWebhook ───────────────────────────────────────────

func makeHMAC(body []byte, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	return hex.EncodeToString(mac.Sum(nil))
}

func TestHandleWebhook_InvalidSignature(t *testing.T) {
	svc := newTestService("supersecret")
	body := []byte(`{"ProviderPaymentID":"pay-1","Status":"paid"}`)

	err := svc.HandleWebhook(context.Background(), body, "wrong-signature", WebhookEvent{})

	assert.ErrorIs(t, err, apierrors.ErrUnauthorized)
}

func TestHandleWebhook_EmptySignature(t *testing.T) {
	svc := newTestService("supersecret")
	body := []byte(`{"ProviderPaymentID":"pay-1","Status":"paid"}`)

	err := svc.HandleWebhook(context.Background(), body, "", WebhookEvent{})

	assert.ErrorIs(t, err, apierrors.ErrUnauthorized)
}

func TestVerifySignature_Valid(t *testing.T) {
	secret := "my-webhook-secret"
	svc := newTestService(secret)
	body := []byte(`{"ProviderPaymentID":"pay-42","Status":"succeeded"}`)
	sig := makeHMAC(body, secret)

	// A valid HMAC passes the signature check and proceeds to the repo call.
	// With a nil repo the DB call panics — assert.Panics confirms the HMAC
	// check itself did NOT reject the request (ErrUnauthorized would have
	// returned before the repo call, not panicked).
	assert.Panics(t, func() {
		_ = svc.HandleWebhook(context.Background(), body, sig, WebhookEvent{ProviderPaymentID: "pay-42", Status: "succeeded"})
	}, "valid signature should pass auth check and reach repo (nil DB panics)")
}
