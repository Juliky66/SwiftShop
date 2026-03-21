// Package dto contains shared data transfer objects used for inter-service communication.
package dto

// ProductSnapshot holds a price/name snapshot captured at checkout time.
type ProductSnapshot struct {
	ProductID   string  `json:"product_id"`
	VariantID   string  `json:"variant_id"`
	ProductName string  `json:"product_name"`
	SKU         string  `json:"sku"`
	SellerID    string  `json:"seller_id"`
	UnitPrice   float64 `json:"unit_price"`
}

// StockReserveRequest is used by orders service to reserve stock in catalog.
type StockReserveRequest struct {
	VariantID string `json:"variant_id"`
	Quantity  int    `json:"quantity"`
}

// VerifiedPurchase is returned by orders service to confirm a delivered order.
type VerifiedPurchase struct {
	OrderID   string `json:"order_id"`
	UserID    string `json:"user_id"`
	ProductID string `json:"product_id"`
}

// OrderStatusUpdate is sent by payments service to update an order's status.
type OrderStatusUpdate struct {
	Status string `json:"status"`
}

// ProductRatingUpdate is sent by reviews service after a new review is saved.
type ProductRatingUpdate struct {
	Rating      float64 `json:"rating"`
	ReviewCount int     `json:"review_count"`
}
