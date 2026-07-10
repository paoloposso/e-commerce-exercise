package services

import (
	"context"
	"errors"
	"testing"

	"ntd/backend/models"
)

type mockProductRepository struct {
	listFn     func(ctx context.Context, query, category string) ([]models.Product, error)
	getByIDFn  func(ctx context.Context, id string) (*models.Product, error)
	getBySKUFn func(ctx context.Context, sku string) (*models.Product, error)
	createFn   func(ctx context.Context, p *models.Product) error
	updateFn   func(ctx context.Context, id string, p *models.Product) error
	deleteFn   func(ctx context.Context, id string) error
}

func (m *mockProductRepository) List(ctx context.Context, query, category string) ([]models.Product, error) {
	return m.listFn(ctx, query, category)
}
func (m *mockProductRepository) GetByID(ctx context.Context, id string) (*models.Product, error) {
	return m.getByIDFn(ctx, id)
}
func (m *mockProductRepository) GetBySKU(ctx context.Context, sku string) (*models.Product, error) {
	return m.getBySKUFn(ctx, sku)
}
func (m *mockProductRepository) Create(ctx context.Context, p *models.Product) error {
	return m.createFn(ctx, p)
}
func (m *mockProductRepository) Update(ctx context.Context, id string, p *models.Product) error {
	return m.updateFn(ctx, id, p)
}
func (m *mockProductRepository) Delete(ctx context.Context, id string) error {
	return m.deleteFn(ctx, id)
}

func TestCreateProduct_ServiceValidations(t *testing.T) {
	mockRepo := &mockProductRepository{
		getBySKUFn: func(ctx context.Context, sku string) (*models.Product, error) {
			if sku == "SKU-EXISTS" {
				return &models.Product{ID: "existing-uuid", SKU: "SKU-EXISTS"}, nil
			}
			return nil, nil
		},
		createFn: func(ctx context.Context, p *models.Product) error {
			return nil
		},
	}

	service := NewProductService(mockRepo)

	p1 := &models.Product{SKU: "SKU-VAL", Name: "  "}
	err := service.CreateProduct(context.Background(), p1)
	if !errors.Is(err, ErrInvalidInput) {
		t.Errorf("Expected ErrInvalidInput for empty name, got %v", err)
	}

	p2 := &models.Product{SKU: "SKU-VAL", Name: "Test Product", Price: -10.00}
	err = service.CreateProduct(context.Background(), p2)
	if !errors.Is(err, ErrInvalidInput) {
		t.Errorf("Expected ErrInvalidInput for negative price, got %v", err)
	}

	p3 := &models.Product{SKU: "SKU-VAL", Name: "Test Product", Stock: -5}
	err = service.CreateProduct(context.Background(), p3)
	if !errors.Is(err, ErrInvalidInput) {
		t.Errorf("Expected ErrInvalidInput for negative stock, got %v", err)
	}

	p4 := &models.Product{SKU: "SKU-EXISTS", Name: "New Product"}
	err = service.CreateProduct(context.Background(), p4)
	if !errors.Is(err, ErrSKUDuplicate) {
		t.Errorf("Expected ErrSKUDuplicate for existing SKU, got %v", err)
	}

	p5 := &models.Product{SKU: "SKU-NEW-STUFF", Name: "Unique Product", Price: 15.50, Stock: 10, WeightKg: 0.5}
	err = service.CreateProduct(context.Background(), p5)
	if err != nil {
		t.Errorf("Expected successful product creation, got error: %v", err)
	}
	if p5.ID == "" {
		t.Errorf("Expected product to have a UUID assigned, got empty string")
	}
}

func TestDeleteProduct_ServiceNotFound(t *testing.T) {
	mockRepo := &mockProductRepository{
		deleteFn: func(ctx context.Context, id string) error {
			return ErrNotFound
		},
	}

	service := NewProductService(mockRepo)

	err := service.DeleteProduct(context.Background(), "missing-id")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("Expected ErrNotFound for missing ID delete, got %v", err)
	}
}
