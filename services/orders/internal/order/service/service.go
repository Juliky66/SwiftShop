package service

import (
	"context"
	"encoding/json"

	apierrors "shopkuber/shared/errors"
	"shopkuber/shared/pagination"
	catalogclient "shopkuber/orders/internal/catalog"
	cartrepo "shopkuber/orders/internal/cart/repository"
	orderrepo "shopkuber/orders/internal/order/repository"
)

// DeliveryAddress holds shipping details.
type DeliveryAddress struct {
	FullName   string `json:"full_name"`
	Phone      string `json:"phone"`
	City       string `json:"city"`
	Street     string `json:"street"`
	PostalCode string `json:"postal_code"`
}

// OrderDetail holds an order with its items.
type OrderDetail struct {
	*orderrepo.Order
	Items []orderrepo.OrderItem `json:"items"`
}

// Service handles order business logic.
type Service struct {
	orders  *orderrepo.Repository
	carts   *cartrepo.Repository
	catalog *catalogclient.Client
}

// New creates a new order service.
func New(orders *orderrepo.Repository, carts *cartrepo.Repository, catalog *catalogclient.Client) *Service {
	return &Service{orders: orders, carts: carts, catalog: catalog}
}

// GetCart returns the authenticated user's cart with items.
func (s *Service) GetCart(ctx context.Context, userID string) (*cartrepo.Cart, []cartrepo.CartItem, error) {
	cart, err := s.carts.GetOrCreate(ctx, userID)
	if err != nil {
		return nil, nil, err
	}
	items, err := s.carts.Items(ctx, cart.ID)
	if err != nil {
		return nil, nil, err
	}
	return cart, items, nil
}

// AddToCart adds an item to the user's cart, fetching current price from catalog.
func (s *Service) AddToCart(ctx context.Context, userID, variantID string, quantity int) error {
	variantInfo, err := s.catalog.GetVariant(ctx, variantID)
	if err != nil {
		return apierrors.Wrap(apierrors.ErrNotFound, "variant not found in catalog")
	}
	if variantInfo.Variant.Stock < quantity {
		return apierrors.Wrap(apierrors.ErrConflict, "insufficient stock")
	}

	cart, err := s.carts.GetOrCreate(ctx, userID)
	if err != nil {
		return err
	}

	return s.carts.UpsertItem(ctx, cart.ID, variantID, quantity, variantInfo.Variant.Price)
}

// UpdateCartItem updates the quantity of a cart item.
func (s *Service) UpdateCartItem(ctx context.Context, userID, variantID string, quantity int) error {
	if quantity <= 0 {
		return s.RemoveFromCart(ctx, userID, variantID)
	}
	cart, err := s.carts.GetOrCreate(ctx, userID)
	if err != nil {
		return err
	}
	return s.carts.UpdateQuantity(ctx, cart.ID, variantID, quantity)
}

// RemoveFromCart removes an item from the cart.
func (s *Service) RemoveFromCart(ctx context.Context, userID, variantID string) error {
	cart, err := s.carts.GetOrCreate(ctx, userID)
	if err != nil {
		return err
	}
	return s.carts.RemoveItem(ctx, cart.ID, variantID)
}

// ClearCart removes all items from the cart.
func (s *Service) ClearCart(ctx context.Context, userID string) error {
	cart, err := s.carts.GetOrCreate(ctx, userID)
	if err != nil {
		return err
	}
	return s.carts.Clear(ctx, cart.ID)
}

// Checkout converts the cart into an order.
func (s *Service) Checkout(ctx context.Context, userID string, addr DeliveryAddress) (*OrderDetail, error) {
	cart, items, err := s.GetCart(ctx, userID)
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, apierrors.Wrap(apierrors.ErrBadRequest, "cart is empty")
	}

	// Build order items, reserving stock for each variant
	var orderItems []orderrepo.CreateOrderItemInput
	var total float64

	for _, item := range items {
		variantInfo, err := s.catalog.GetVariant(ctx, item.VariantID)
		if err != nil {
			return nil, apierrors.Wrap(apierrors.ErrNotFound, "variant "+item.VariantID+" not found")
		}

		if err := s.catalog.ReserveStock(ctx, item.VariantID, item.Quantity); err != nil {
			return nil, apierrors.Wrap(apierrors.ErrConflict, "could not reserve stock for "+variantInfo.Variant.SKU)
		}

		unitPrice := variantInfo.Variant.Price
		orderItems = append(orderItems, orderrepo.CreateOrderItemInput{
			VariantID:   item.VariantID,
			SellerID:    variantInfo.Product.SellerID,
			ProductName: variantInfo.Product.Name,
			SKU:         variantInfo.Variant.SKU,
			Quantity:    item.Quantity,
			UnitPrice:   unitPrice,
		})
		total += unitPrice * float64(item.Quantity)
	}

	addrJSON, err := json.Marshal(addr)
	if err != nil {
		return nil, err
	}

	order, err := s.orders.Create(ctx, orderrepo.CreateOrderInput{
		UserID:          userID,
		TotalAmount:     total,
		DeliveryAddress: string(addrJSON),
		Items:           orderItems,
	})
	if err != nil {
		return nil, err
	}

	// Clear the cart after successful checkout
	_ = s.carts.Clear(ctx, cart.ID)

	orderItemsResult, err := s.orders.Items(ctx, order.ID)
	if err != nil {
		return nil, err
	}

	return &OrderDetail{Order: order, Items: orderItemsResult}, nil
}

// GetOrder returns an order with its items.
func (s *Service) GetOrder(ctx context.Context, orderID, userID string) (*OrderDetail, error) {
	order, err := s.orders.FindByID(ctx, orderID, userID)
	if err != nil {
		return nil, err
	}
	items, err := s.orders.Items(ctx, order.ID)
	if err != nil {
		return nil, err
	}
	return &OrderDetail{Order: order, Items: items}, nil
}

// GetOrderTotalAmount returns only the total_amount for an order (no ownership check, internal use).
func (s *Service) GetOrderTotalAmount(ctx context.Context, orderID string) (float64, error) {
	order, err := s.orders.FindByIDAdmin(ctx, orderID)
	if err != nil {
		return 0, err
	}
	return order.TotalAmount, nil
}

// ListOrders returns paginated order history for a user.
func (s *Service) ListOrders(ctx context.Context, userID string, pg pagination.Request) (pagination.Page[orderrepo.Order], error) {
	orders, total, err := s.orders.ListByUser(ctx, userID, pg.Limit, pg.Offset)
	if err != nil {
		return pagination.Page[orderrepo.Order]{}, err
	}
	return pagination.NewPage(orders, total, pg.Limit, pg.Offset), nil
}

// CancelOrder sets an order to cancelled if it's still in a cancellable state.
func (s *Service) CancelOrder(ctx context.Context, orderID, userID string) error {
	order, err := s.orders.FindByID(ctx, orderID, userID)
	if err != nil {
		return err
	}
	if order.Status != "pending" && order.Status != "confirmed" && order.Status != "processing" {
		return apierrors.Wrap(apierrors.ErrConflict, "order cannot be cancelled at this stage")
	}
	return s.orders.UpdateStatus(ctx, orderID, "cancelled")
}

// UpdateStatus updates an order's status (used by payments webhook).
func (s *Service) UpdateStatus(ctx context.Context, orderID, status string) error {
	return s.orders.UpdateStatus(ctx, orderID, status)
}

// VerifiedPurchase checks if a user has a delivered order for a product (used by reviews).
func (s *Service) VerifiedPurchase(ctx context.Context, userID, productID string) (string, error) {
	return s.orders.VerifiedPurchase(ctx, userID, productID)
}
