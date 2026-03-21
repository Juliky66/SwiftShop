package service

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"

	apierrors "shopkuber/shared/errors"
	"shopkuber/reviews/internal/payment/gateway"
	paymentrepo "shopkuber/reviews/internal/payment/repository"
	ordersclient "shopkuber/reviews/internal/orders"
)

// InitiateResponse is returned after creating a payment.
type InitiateResponse struct {
	PaymentID  string `json:"payment_id"`
	ConfirmURL string `json:"confirm_url"`
	Status     string `json:"status"`
}

// Service handles payment business logic.
type Service struct {
	repo            *paymentrepo.Repository
	gateway         gateway.Gateway
	orders          *ordersclient.Client
	webhookSecret   string
	providerName    string
}

// New creates a new payment service.
func New(repo *paymentrepo.Repository, gw gateway.Gateway, orders *ordersclient.Client, webhookSecret, providerName string) *Service {
	return &Service{
		repo:          repo,
		gateway:       gw,
		orders:        orders,
		webhookSecret: webhookSecret,
		providerName:  providerName,
	}
}

// Initiate creates a payment for an order.
func (s *Service) Initiate(ctx context.Context, orderID, userID string) (*InitiateResponse, error) {
	// Get order amount from orders service
	amount, err := s.orders.GetOrderAmount(ctx, orderID, userID)
	if err != nil {
		return nil, apierrors.Wrap(apierrors.ErrNotFound, "order not found")
	}

	resp, err := s.gateway.CreatePayment(ctx, gateway.PaymentRequest{
		OrderID:     orderID,
		Amount:      amount,
		Currency:    "RUB",
		Description: "ShopKuber order " + orderID,
		ReturnURL:   "https://shopkuber.example.com/orders/" + orderID,
	})
	if err != nil {
		return nil, err
	}

	payment, err := s.repo.Create(ctx, orderID, s.providerName, resp.ProviderID, amount, "RUB")
	if err != nil {
		return nil, err
	}

	// Mark order as processing immediately so duplicate payments are prevented
	_ = s.orders.UpdateOrderStatus(ctx, orderID, "processing")

	return &InitiateResponse{
		PaymentID:  payment.ID,
		ConfirmURL: resp.ConfirmURL,
		Status:     payment.Status,
	}, nil
}

// GetStatus returns the current payment status.
func (s *Service) GetStatus(ctx context.Context, paymentID string) (*paymentrepo.Payment, error) {
	return s.repo.FindByID(ctx, paymentID)
}

// HandleWebhook processes an incoming payment webhook.
// signature is the HMAC-SHA256 hex digest of the raw body, keyed with webhookSecret.
func (s *Service) HandleWebhook(ctx context.Context, rawBody []byte, signature string, event WebhookEvent) error {
	// Verify HMAC signature
	if !s.verifySignature(rawBody, signature) {
		return apierrors.ErrUnauthorized
	}

	payment, err := s.repo.FindByProviderID(ctx, event.ProviderPaymentID)
	if err != nil {
		return err
	}

	newStatus := mapProviderStatus(event.Status)
	if err := s.repo.UpdateStatus(ctx, payment.ID, newStatus); err != nil {
		return err
	}

	// Propagate to orders service
	if newStatus == "succeeded" {
		_ = s.orders.UpdateOrderStatus(ctx, payment.OrderID, "paid")
	} else if newStatus == "failed" || newStatus == "refunded" {
		_ = s.orders.UpdateOrderStatus(ctx, payment.OrderID, "cancelled")
	}

	return nil
}

// WebhookEvent is a parsed incoming event from the payment provider.
type WebhookEvent struct {
	ProviderPaymentID string
	Status            string
}

func (s *Service) verifySignature(body []byte, signature string) bool {
	mac := hmac.New(sha256.New, []byte(s.webhookSecret))
	mac.Write(body)
	expected := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(signature))
}

func mapProviderStatus(providerStatus string) string {
	switch providerStatus {
	case "succeeded", "success", "paid":
		return "succeeded"
	case "failed", "cancelled", "canceled":
		return "failed"
	case "refunded":
		return "refunded"
	default:
		return "pending"
	}
}
