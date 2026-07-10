package services

import (
	"context"
	"ntd/backend/models"
)

// OrderRepository defines the database operations required by the order service.
type OrderRepository interface {
	GetBySKU(ctx context.Context, sku string) (*models.Product, error)
	GetOrderByIdempotencyKey(ctx context.Context, idempotencyKey string) (*models.Order, error)
	TryDecrementStock(ctx context.Context, sku string, quantity, expectedVersion int) (bool, error)
	RestoreStock(ctx context.Context, sku string, quantity int) error
	CreateOrder(ctx context.Context, order *models.Order) error
}
