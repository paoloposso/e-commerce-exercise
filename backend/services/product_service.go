package services

import (
	"context"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strings"

	"ntd/backend/models"
)

var (
	ErrNotFound          = errors.New("product not found")
	ErrSKUDuplicate      = errors.New("product sku already exists")
	ErrInvalidInput      = errors.New("invalid product input data")
	ErrInsufficientStock = errors.New("insufficient stock available")
)

type ProductService struct {
	repo ProductRepository
}

func NewProductService(repo ProductRepository) *ProductService {
	return &ProductService{repo: repo}
}

func (s *ProductService) ListProducts(ctx context.Context, query, category string) ([]models.Product, error) {
	return s.repo.List(ctx, query, category)
}

func (s *ProductService) ImportProducts(ctx context.Context, reader io.Reader) (*ImportReport, error) {
	return ImportProductsFromCSV(ctx, reader, s.repo)
}

var skuRegex = regexp.MustCompile(`^[a-zA-Z0-9-_]+$`)

func validateProduct(p *models.Product) error {
	if strings.TrimSpace(p.Name) == "" {
		return fmt.Errorf("%w: product name is required", ErrInvalidInput)
	}
	if strings.TrimSpace(p.SKU) == "" {
		return fmt.Errorf("%w: SKU is required", ErrInvalidInput)
	}
	if !skuRegex.MatchString(strings.TrimSpace(p.SKU)) {
		return fmt.Errorf("%w: SKU can only contain alphanumeric characters, hyphens, and underscores", ErrInvalidInput)
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
	return nil
}

func (s *ProductService) CreateProduct(ctx context.Context, p *models.Product) error {
	if err := validateProduct(p); err != nil {
		return err
	}
	p.Name = strings.TrimSpace(p.Name)
	p.SKU = strings.TrimSpace(p.SKU)
	p.Category = strings.TrimSpace(p.Category)
	p.Description = strings.TrimSpace(p.Description)

	existing, err := s.repo.GetBySKU(ctx, p.SKU)
	if err != nil {
		return err
	}
	if existing != nil {
		return ErrSKUDuplicate
	}
	return s.repo.Create(ctx, p)
}

func (s *ProductService) UpdateProduct(ctx context.Context, id string, p *models.Product) error {
	if err := validateProduct(p); err != nil {
		return err
	}
	p.Name = strings.TrimSpace(p.Name)
	p.SKU = strings.TrimSpace(p.SKU)
	p.Category = strings.TrimSpace(p.Category)
	p.Description = strings.TrimSpace(p.Description)

	existing, err := s.repo.GetBySKU(ctx, p.SKU)
	if err != nil {
		return err
	}
	if existing != nil && existing.ID != id {
		return ErrSKUDuplicate
	}
	return s.repo.Update(ctx, id, p)
}

func (s *ProductService) DeleteProduct(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}
