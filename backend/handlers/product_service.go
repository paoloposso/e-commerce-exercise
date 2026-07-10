package handlers

import (
	"context"
	"io"

	"ecommerce/backend/models"
	"ecommerce/backend/services"
)

type ProductService interface {
	ListProducts(ctx context.Context, query, category string) ([]models.Product, error)
	ImportProducts(ctx context.Context, reader io.Reader) (*services.ImportReport, error)
	CreateProduct(ctx context.Context, p *models.Product) error
	UpdateProduct(ctx context.Context, id string, p *models.Product) error
	DeleteProduct(ctx context.Context, id string) error
}
