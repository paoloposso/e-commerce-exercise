package services

import (
	"context"
	"ecommerce/backend/models"
)

type ProductImporterStore interface {
	ImportProducts(ctx context.Context, products []models.Product) (imported int, updated int, err error)
}

type ProductRepository interface {
	ProductImporterStore
	List(ctx context.Context, query, category string) ([]models.Product, error)
	GetByID(ctx context.Context, id string) (*models.Product, error)
	GetBySKU(ctx context.Context, sku string) (*models.Product, error)
	Create(ctx context.Context, p *models.Product) error
	Update(ctx context.Context, id string, p *models.Product) error
	Delete(ctx context.Context, id string) error
	TryDecrementStock(ctx context.Context, sku string, quantity, expectedVersion int) (bool, error)
	RestoreStock(ctx context.Context, sku string, quantity int) error
}

