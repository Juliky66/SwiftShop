// Package orders provides an HTTP client for the orders service.
package orders

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// Client calls the orders service internal API.
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// New creates a new orders client.
func New(baseURL string) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{Timeout: 5 * time.Second},
	}
}

// VerifiedPurchase checks if a user has a delivered order for a product.
// Returns the order ID if found, or an error if not.
func (c *Client) VerifiedPurchase(ctx context.Context, userID, productID string) (string, error) {
	url := fmt.Sprintf("%s/api/v1/internal/orders/verified?user_id=%s&product_id=%s", c.baseURL, userID, productID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return "", fmt.Errorf("no verified purchase")
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("orders: unexpected status %d", resp.StatusCode)
	}
	var result struct {
		OrderID string `json:"order_id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	return result.OrderID, nil
}

// UpdateOrderStatus sets an order's status (used after payment webhook).
func (c *Client) UpdateOrderStatus(ctx context.Context, orderID, status string) error {
	url := fmt.Sprintf("%s/api/v1/internal/orders/%s/status", c.baseURL, orderID)
	body := fmt.Sprintf(`{"status":%q}`, status)
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, strings.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("orders: update status failed with status %d", resp.StatusCode)
	}
	return nil
}

// GetOrderAmount returns the total amount for an order via the internal endpoint.
func (c *Client) GetOrderAmount(ctx context.Context, orderID, userID string) (float64, error) {
	url := fmt.Sprintf("%s/api/v1/internal/orders/%s/amount?user_id=%s", c.baseURL, orderID, userID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return 0, err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("orders: could not get order %s", orderID)
	}
	var result struct {
		TotalAmount float64 `json:"total_amount"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, err
	}
	return result.TotalAmount, nil
}
