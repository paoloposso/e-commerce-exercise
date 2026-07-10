package handlers

import (
	"context"
	"io"

	"ntd/backend/models"
	"ntd/backend/services"
)

// ProductService defines the catalog operations consumed by the product HTTP handler.
type ProductService interface {
	ListProducts(ctx context.Context, query, category string) ([]models.Product, error)
	CreateProduct(ctx context.Context, p *models.Product) error
	UpdateProduct(ctx context.Context, id string, p *models.Product) error
	DeleteProduct(ctx context.Context, id string) error
	ImportProducts(ctx context.Context, reader io.Reader) (*services.ImportReport, error)
}
