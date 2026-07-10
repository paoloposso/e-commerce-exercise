package services

import (
	"context"
	"ecommerce/backend/models"
)

type ProductRepository interface {
	List(ctx context.Context, query, category string) ([]models.Product, error)
	GetByID(ctx context.Context, id string) (*models.Product, error)
	GetBySKU(ctx context.Context, sku string) (*models.Product, error)
	Create(ctx context.Context, p *models.Product) error
	Update(ctx context.Context, id string, p *models.Product) error
	Delete(ctx context.Context, id string) error
}

type ProductImporterStore interface {
	GetBySKU(ctx context.Context, sku string) (*models.Product, error)
	Create(ctx context.Context, p *models.Product) error
	Update(ctx context.Context, id string, p *models.Product) error
}
