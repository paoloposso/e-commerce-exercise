package services

import (
	"context"
	"ecommerce/backend/models"
)

type OrderRepository interface {
	GetBySKU(ctx context.Context, sku string) (*models.Product, error)
	GetOrderByIdempotencyKey(ctx context.Context, idempotencyKey string) (*models.Order, error)
	TryDecrementStock(ctx context.Context, sku string, quantity, expectedVersion int) (bool, error)
	RestoreStock(ctx context.Context, sku string, quantity int) error
	CreateOrder(ctx context.Context, order *models.Order) error
	ListOrders(ctx context.Context) ([]models.Order, error)
}
