package handlers

import (
	"context"

	"ecommerce/backend/models"
)

type OrderService interface {
	PurchaseProduct(ctx context.Context, sku, customerID, idempotencyKey string, quantity int, expectedPrice float64) (*models.Order, error)
	ListOrders(ctx context.Context) ([]models.Order, error)
}
