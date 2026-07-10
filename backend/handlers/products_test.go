package handlers

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"ntd/backend/models"
	"ntd/backend/services"
)

type mockProductService struct {
	listProductsFn   func(ctx context.Context, query, category string) ([]models.Product, error)
	importProductsFn func(ctx context.Context, reader io.Reader) (*services.ImportReport, error)
	createProductFn  func(ctx context.Context, p *models.Product) error
	updateProductFn  func(ctx context.Context, id string, p *models.Product) error
	deleteProductFn  func(ctx context.Context, id string) error
}

func (m *mockProductService) ListProducts(ctx context.Context, query, category string) ([]models.Product, error) {
	return m.listProductsFn(ctx, query, category)
}
func (m *mockProductService) ImportProducts(ctx context.Context, reader io.Reader) (*services.ImportReport, error) {
	return m.importProductsFn(ctx, reader)
}
func (m *mockProductService) CreateProduct(ctx context.Context, p *models.Product) error {
	if m.createProductFn != nil {
		return m.createProductFn(ctx, p)
	}
	return nil
}
func (m *mockProductService) UpdateProduct(ctx context.Context, id string, p *models.Product) error {
	if m.updateProductFn != nil {
		return m.updateProductFn(ctx, id, p)
	}
	return nil
}
func (m *mockProductService) DeleteProduct(ctx context.Context, id string) error {
	if m.deleteProductFn != nil {
		return m.deleteProductFn(ctx, id)
	}
	return nil
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

func TestCreateProduct_Mocked(t *testing.T) {
	called := false
	mockService := &mockProductService{
		createProductFn: func(ctx context.Context, p *models.Product) error {
			called = true
			if p.Name != "New Product" {
				t.Errorf("Expected Name 'New Product', got '%s'", p.Name)
			}
			return nil
		},
	}

	productHandler := NewProductHandler(mockService)
	body := strings.NewReader(`{"name":"New Product","sku":"NP-1","price":10.5,"stock":5}`)
	req, _ := http.NewRequest("POST", "/api/products", body)
	rr := httptest.NewRecorder()

	productHandler.CreateProduct(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("Expected status code %d, got %d", http.StatusCreated, rr.Code)
	}
	if !called {
		t.Error("Expected createProductFn to be called")
	}
}

func TestUpdateProduct_Mocked(t *testing.T) {
	called := false
	mockService := &mockProductService{
		updateProductFn: func(ctx context.Context, id string, p *models.Product) error {
			called = true
			if id != "test-id" {
				t.Errorf("Expected ID 'test-id', got '%s'", id)
			}
			if p.Name != "Updated Name" {
				t.Errorf("Expected Name 'Updated Name', got '%s'", p.Name)
			}
			return nil
		},
	}

	productHandler := NewProductHandler(mockService)
	body := strings.NewReader(`{"name":"Updated Name","sku":"NP-1","price":10.5,"stock":5}`)
	req, _ := http.NewRequest("PUT", "/api/products/test-id", body)
	req.SetPathValue("id", "test-id")
	rr := httptest.NewRecorder()

	productHandler.UpdateProduct(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, rr.Code)
	}
	if !called {
		t.Error("Expected updateProductFn to be called")
	}
}

func TestDeleteProduct_Mocked(t *testing.T) {
	called := false
	mockService := &mockProductService{
		deleteProductFn: func(ctx context.Context, id string) error {
			called = true
			if id != "test-id" {
				t.Errorf("Expected ID 'test-id', got '%s'", id)
			}
			return nil
		},
	}

	productHandler := NewProductHandler(mockService)
	req, _ := http.NewRequest("DELETE", "/api/products/test-id", nil)
	req.SetPathValue("id", "test-id")
	rr := httptest.NewRecorder()

	productHandler.DeleteProduct(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, rr.Code)
	}
	if !called {
		t.Error("Expected deleteProductFn to be called")
	}
}
