package services

import (
	"context"
	"ecommerce/backend/models"
)

type OrderRepository interface {
	GetOrderByIdempotencyKey(ctx context.Context, idempotencyKey string) (*models.Order, error)
	CreateOrder(ctx context.Context, order *models.Order) error
	ListOrders(ctx context.Context) ([]models.Order, error)
}
