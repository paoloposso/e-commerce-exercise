package services

import (
	"context"
	"errors"
	"testing"

	"ecommerce/backend/models"
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

func TestProductService_CreateProduct(t *testing.T) {
	t.Run("valid input", func(t *testing.T) {
		repo := &mockProductRepository{
			getBySKUFn: func(ctx context.Context, sku string) (*models.Product, error) {
				return nil, nil
			},
			createFn: func(ctx context.Context, p *models.Product) error {
				return nil
			},
		}
		s := NewProductService(repo)
		err := s.CreateProduct(context.Background(), &models.Product{
			Name:  "Test",
			SKU:   "TEST-1",
			Price: 10.0,
			Stock: 5,
		})
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("duplicate sku", func(t *testing.T) {
		repo := &mockProductRepository{
			getBySKUFn: func(ctx context.Context, sku string) (*models.Product, error) {
				return &models.Product{ID: "existing-id", SKU: sku}, nil
			},
		}
		s := NewProductService(repo)
		err := s.CreateProduct(context.Background(), &models.Product{
			Name:  "Test",
			SKU:   "TEST-1",
			Price: 10.0,
			Stock: 5,
		})
		if !errors.Is(err, ErrSKUDuplicate) {
			t.Fatalf("expected ErrSKUDuplicate, got %v", err)
		}
	})

	t.Run("invalid input", func(t *testing.T) {
		repo := &mockProductRepository{}
		s := NewProductService(repo)
		err := s.CreateProduct(context.Background(), &models.Product{
			Name:  "",
			SKU:   "TEST-1",
			Price: 10.0,
			Stock: 5,
		})
		if !errors.Is(err, ErrInvalidInput) {
			t.Fatalf("expected ErrInvalidInput, got %v", err)
		}
	})
}

func TestProductService_UpdateProduct(t *testing.T) {
	t.Run("valid update", func(t *testing.T) {
		repo := &mockProductRepository{
			getBySKUFn: func(ctx context.Context, sku string) (*models.Product, error) {
				return &models.Product{ID: "prod-1", SKU: sku}, nil
			},
			updateFn: func(ctx context.Context, id string, p *models.Product) error {
				return nil
			},
		}
		s := NewProductService(repo)
		err := s.UpdateProduct(context.Background(), "prod-1", &models.Product{
			Name:  "Updated name",
			SKU:   "TEST-1",
			Price: 10.0,
			Stock: 5,
		})
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("duplicate SKU on different product", func(t *testing.T) {
		repo := &mockProductRepository{
			getBySKUFn: func(ctx context.Context, sku string) (*models.Product, error) {
				return &models.Product{ID: "prod-2", SKU: sku}, nil
			},
		}
		s := NewProductService(repo)
		err := s.UpdateProduct(context.Background(), "prod-1", &models.Product{
			Name:  "Updated name",
			SKU:   "TEST-1",
			Price: 10.0,
			Stock: 5,
		})
		if !errors.Is(err, ErrSKUDuplicate) {
			t.Fatalf("expected ErrSKUDuplicate, got %v", err)
		}
	})
}
