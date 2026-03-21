// Package catalog provides an HTTP client for the catalog service (used to update ratings).
package catalog

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// Client calls the catalog service internal API.
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// New creates a new catalog client.
func New(baseURL string) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{Timeout: 5 * time.Second},
	}
}

// UpdateRating sends updated rating info to the catalog service.
func (c *Client) UpdateRating(ctx context.Context, productID string, rating float64, reviewCount int) error {
	url := fmt.Sprintf("%s/api/v1/internal/products/%s/rating", c.baseURL, productID)
	body := fmt.Sprintf(`{"rating":%f,"review_count":%d}`, rating, reviewCount)
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
		return fmt.Errorf("catalog: update rating failed with status %d", resp.StatusCode)
	}
	return nil
}
