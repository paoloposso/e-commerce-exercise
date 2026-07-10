package handlers

import (
	"context"

	"ntd/backend/models"
)

// OrderService defines the checkout operations consumed by the order HTTP handler.
type OrderService interface {
	PurchaseProduct(ctx context.Context, sku, customerID, idempotencyKey string, quantity int) (*models.Order, error)
}
