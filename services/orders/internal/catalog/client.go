// Package catalog provides an HTTP client for communicating with the catalog service.
package catalog

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// VariantInfo holds variant + product data returned by the catalog service.
type VariantInfo struct {
	Variant struct {
		ID        string  `json:"id"`
		ProductID string  `json:"product_id"`
		SKU       string  `json:"sku"`
		Price     float64 `json:"price"`
		Stock     int     `json:"stock"`
	} `json:"variant"`
	Product struct {
		ID       string `json:"id"`
		Name     string `json:"name"`
		SellerID string `json:"seller_id"`
	} `json:"product"`
}

// Client is an HTTP client for the catalog service.
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// New creates a new catalog client.
func New(baseURL string) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

// GetVariant fetches variant + product info from the catalog service.
func (c *Client) GetVariant(ctx context.Context, variantID string) (*VariantInfo, error) {
	url := fmt.Sprintf("%s/api/v1/internal/variants/%s", c.baseURL, variantID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("catalog: unexpected status %d for variant %s", resp.StatusCode, variantID)
	}

	var info VariantInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, err
	}
	return &info, nil
}

// ReserveStock decrements stock for a variant in the catalog service.
func (c *Client) ReserveStock(ctx context.Context, variantID string, quantity int) error {
	url := fmt.Sprintf("%s/api/v1/internal/variants/%s/reserve", c.baseURL, variantID)
	body := fmt.Sprintf(`{"quantity":%d}`, quantity)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, strings.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNoContent {
		return nil
	}
	return fmt.Errorf("catalog: stock reserve failed with status %d", resp.StatusCode)
}
