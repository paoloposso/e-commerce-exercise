package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"ntd/backend/models"
	"ntd/backend/services"
)

type mockProductService struct {
	listProductsFn  func(ctx context.Context, query, category string) ([]models.Product, error)
	createProductFn func(ctx context.Context, p *models.Product) error
	updateProductFn func(ctx context.Context, id string, p *models.Product) error
	deleteProductFn func(ctx context.Context, id string) error
	importProductsFn func(ctx context.Context, reader io.Reader) (*services.ImportReport, error)
}

func (m *mockProductService) ListProducts(ctx context.Context, query, category string) ([]models.Product, error) {
	return m.listProductsFn(ctx, query, category)
}
func (m *mockProductService) CreateProduct(ctx context.Context, p *models.Product) error {
	return m.createProductFn(ctx, p)
}
func (m *mockProductService) UpdateProduct(ctx context.Context, id string, p *models.Product) error {
	return m.updateProductFn(ctx, id, p)
}
func (m *mockProductService) DeleteProduct(ctx context.Context, id string) error {
	return m.deleteProductFn(ctx, id)
}
func (m *mockProductService) ImportProducts(ctx context.Context, reader io.Reader) (*services.ImportReport, error) {
	return m.importProductsFn(ctx, reader)
}

func TestListProducts_Mocked(t *testing.T) {
	mockService := &mockProductService{
		listProductsFn: func(ctx context.Context, query, category string) ([]models.Product, error) {
			return []models.Product{
				{ID: "m-1", Name: "Mock Product 1", SKU: "SKU-M1", Price: 19.99, Stock: 10, WeightKg: 0.1},
				{ID: "m-2", Name: "Mock Product 2", SKU: "SKU-M2", Price: 29.99, Stock: 20, WeightKg: 0.2},
			}, nil
		},
	}

	productHandler := NewProductHandler(mockService)

	req, _ := http.NewRequest("GET", "/api/products", nil)
	rr := httptest.NewRecorder()

	productHandler.ListAndSearchProducts(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, rr.Code)
	}

	var products []models.Product
	err := json.Unmarshal(rr.Body.Bytes(), &products)
	if err != nil {
		t.Fatalf("Failed to decode response payload: %v", err)
	}

	if len(products) != 2 {
		t.Errorf("Expected 2 products, got %d", len(products))
	}

	if products[0].Name != "Mock Product 1" || products[1].SKU != "SKU-M2" {
		t.Errorf("Decoded list details mismatch: %+v", products)
	}
}

func TestCreateProduct_Mocked_ValidationError(t *testing.T) {
	mockService := &mockProductService{
		createProductFn: func(ctx context.Context, p *models.Product) error {
			return services.ErrInvalidInput
		},
	}
	productHandler := NewProductHandler(mockService)

	p := models.Product{
		Name:  "Invalid Product",
		SKU:   "SKU-INVALID",
		Price: -5.00,
		Stock: 10,
	}
	body, _ := json.Marshal(p)
	req, _ := http.NewRequest("POST", "/api/products", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()

	productHandler.CreateProduct(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, rr.Code)
	}
}
