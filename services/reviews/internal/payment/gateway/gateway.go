// Package gateway defines the payment provider interface and implementations.
package gateway

import "context"

// PaymentRequest is the input for creating a payment.
type PaymentRequest struct {
	OrderID     string
	Amount      float64
	Currency    string
	Description string
	ReturnURL   string
}

// PaymentResponse is returned when a payment is created.
type PaymentResponse struct {
	ProviderID  string
	Status      string
	ConfirmURL  string // redirect URL for the user
}

// Gateway is the interface that payment providers must implement.
type Gateway interface {
	CreatePayment(ctx context.Context, req PaymentRequest) (*PaymentResponse, error)
	GetStatus(ctx context.Context, providerID string) (string, error)
}

// MockGateway is used in development/testing — always succeeds.
type MockGateway struct{}

// CreatePayment returns a fake successful payment.
func (m *MockGateway) CreatePayment(_ context.Context, req PaymentRequest) (*PaymentResponse, error) {
	return &PaymentResponse{
		ProviderID: "mock-" + req.OrderID,
		Status:     "pending",
		ConfirmURL: "http://localhost/mock-pay?order=" + req.OrderID,
	}, nil
}

// GetStatus always returns succeeded for mock payments.
func (m *MockGateway) GetStatus(_ context.Context, _ string) (string, error) {
	return "succeeded", nil
}
