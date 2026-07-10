package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"strings"

	"ntd/backend/models"
)

var (
	ErrNotFound          = errors.New("product not found")
	ErrSKUDuplicate      = errors.New("product sku already exists")
	ErrInvalidInput      = errors.New("invalid product input data")
	ErrInsufficientStock = errors.New("insufficient stock available")
)

// ProductService encapsulates catalog business validations and CSV import.
type ProductService struct {
	repo ProductRepository
}

// NewProductService creates a new ProductService with the given repository.
func NewProductService(repo ProductRepository) *ProductService {
	return &ProductService{repo: repo}
}

// ListProducts queries the catalog filtering by search terms or categories.
func (s *ProductService) ListProducts(ctx context.Context, query, category string) ([]models.Product, error) {
	return s.repo.List(ctx, query, category)
}

// CreateProduct validates and persists a new product.
func (s *ProductService) CreateProduct(ctx context.Context, p *models.Product) error {
	p.SKU = strings.TrimSpace(p.SKU)
	p.Name = strings.TrimSpace(p.Name)

	if p.SKU == "" {
		return fmt.Errorf("%w: SKU is required", ErrInvalidInput)
	}
	if p.Name == "" {
		return fmt.Errorf("%w: product name is required", ErrInvalidInput)
	}
	if p.Price < 0 {
		return fmt.Errorf("%w: price cannot be negative", ErrInvalidInput)
	}
	if p.Stock < 0 {
		return fmt.Errorf("%w: stock cannot be negative", ErrInvalidInput)
	}
	if p.WeightKg < 0 {
		return fmt.Errorf("%w: weight cannot be negative", ErrInvalidInput)
	}

	existing, err := s.repo.GetBySKU(ctx, p.SKU)
	if err != nil {
		return err
	}
	if existing != nil {
		return ErrSKUDuplicate
	}

	p.ID = GenerateUUID()
	return s.repo.Create(ctx, p)
}

// UpdateProduct validates and applies changes to an existing product.
func (s *ProductService) UpdateProduct(ctx context.Context, id string, p *models.Product) error {
	p.SKU = strings.TrimSpace(p.SKU)
	p.Name = strings.TrimSpace(p.Name)

	if id == "" {
		return fmt.Errorf("%w: product ID is required", ErrInvalidInput)
	}
	if p.SKU == "" {
		return fmt.Errorf("%w: SKU is required", ErrInvalidInput)
	}
	if p.Name == "" {
		return fmt.Errorf("%w: product name is required", ErrInvalidInput)
	}
	if p.Price < 0 {
		return fmt.Errorf("%w: price cannot be negative", ErrInvalidInput)
	}
	if p.Stock < 0 {
		return fmt.Errorf("%w: stock cannot be negative", ErrInvalidInput)
	}
	if p.WeightKg < 0 {
		return fmt.Errorf("%w: weight cannot be negative", ErrInvalidInput)
	}

	existing, err := s.repo.GetBySKU(ctx, p.SKU)
	if err != nil {
		return err
	}
	if existing != nil && existing.ID != id {
		return ErrSKUDuplicate
	}

	err = s.repo.Update(ctx, id, p)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		}
		return err
	}
	return nil
}

// DeleteProduct removes a catalog product by its unique ID.
func (s *ProductService) DeleteProduct(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("%w: product ID is required", ErrInvalidInput)
	}

	err := s.repo.Delete(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		}
		return err
	}
	return nil
}

// ImportProducts parses a CSV reader and upserts products into the catalog.
func (s *ProductService) ImportProducts(ctx context.Context, reader io.Reader) (*ImportReport, error) {
	return ImportProductsFromCSV(reader, s.repo)
}
