package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"ecommerce/backend/models"
	"ecommerce/backend/services"
)

type mockOrderService struct {
	purchaseFn func(ctx context.Context, sku, customerID, idempotencyKey string, quantity int, expectedPrice float64) (*models.Order, error)
	listFn     func(ctx context.Context) ([]models.Order, error)
}

func (m *mockOrderService) PurchaseProduct(ctx context.Context, sku, customerID, idempotencyKey string, quantity int, expectedPrice float64) (*models.Order, error) {
	if m.purchaseFn != nil {
		return m.purchaseFn(ctx, sku, customerID, idempotencyKey, quantity, expectedPrice)
	}
	return nil, nil
}

func (m *mockOrderService) ListOrders(ctx context.Context) ([]models.Order, error) {
	if m.listFn != nil {
		return m.listFn(ctx)
	}
	return nil, nil
}

func TestPurchaseProduct_Success(t *testing.T) {
	mockService := &mockOrderService{
		purchaseFn: func(ctx context.Context, sku, customerID, idempotencyKey string, quantity int, expectedPrice float64) (*models.Order, error) {
			return &models.Order{
				ID:             "order-123",
				SKU:            sku,
				CustomerID:     customerID,
				Quantity:       quantity,
				TotalPrice:     expectedPrice * float64(quantity),
				IdempotencyKey: idempotencyKey,
			}, nil
		},
	}

	handler := NewOrderHandler(mockService)

	reqBody := PurchaseRequest{
		SKU:            "SKU-TEST",
		CustomerID:     "customer-123",
		Quantity:       2,
		ExpectedPrice:  10.50,
		IdempotencyKey: "idem-key-123",
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/api/purchase", bytes.NewReader(bodyBytes))
	w := httptest.NewRecorder()

	handler.PurchaseProduct(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d", http.StatusCreated, w.Code)
	}

	var order models.Order
	if err := json.NewDecoder(w.Body).Decode(&order); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if order.ID != "order-123" || order.TotalPrice != 21.00 {
		t.Errorf("Unexpected order response: %+v", order)
	}
}

func TestPurchaseProduct_InvalidJSON(t *testing.T) {
	handler := NewOrderHandler(&mockOrderService{})

	req := httptest.NewRequest("POST", "/api/purchase", strings.NewReader("invalid-json"))
	w := httptest.NewRecorder()

	handler.PurchaseProduct(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestPurchaseProduct_Errors(t *testing.T) {
	tests := []struct {
		name           string
		serviceErr     error
		expectedStatus int
	}{
		{
			name:           "Invalid Input",
			serviceErr:     services.ErrInvalidInput,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Insufficient Stock",
			serviceErr:     services.ErrInsufficientStock,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Not Found",
			serviceErr:     services.ErrNotFound,
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "Payment Declined",
			serviceErr:     services.ErrPaymentDeclined,
			expectedStatus: http.StatusPaymentRequired,
		},
		{
			name:           "Concurrency Conflict",
			serviceErr:     services.ErrConcurrencyConflict,
			expectedStatus: http.StatusConflict,
		},
		{
			name:           "Price Changed Error Message",
			serviceErr:     errors.New("price has changed"),
			expectedStatus: http.StatusConflict,
		},
		{
			name:           "Internal Server Error",
			serviceErr:     errors.New("something went wrong"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &mockOrderService{
				purchaseFn: func(ctx context.Context, sku, customerID, idempotencyKey string, quantity int, expectedPrice float64) (*models.Order, error) {
					return nil, tt.serviceErr
				},
			}

			handler := NewOrderHandler(mockService)
			reqBody := PurchaseRequest{
				SKU:            "SKU-TEST",
				CustomerID:     "customer-123",
				Quantity:       2,
				ExpectedPrice:  10.50,
				IdempotencyKey: "idem-key-123",
			}
			bodyBytes, _ := json.Marshal(reqBody)

			req := httptest.NewRequest("POST", "/api/purchase", bytes.NewReader(bodyBytes))
			w := httptest.NewRecorder()

			handler.PurchaseProduct(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestListOrders_Success(t *testing.T) {
	orders := []models.Order{
		{ID: "order-1", SKU: "SKU-1", Quantity: 1},
		{ID: "order-2", SKU: "SKU-2", Quantity: 2},
	}

	mockService := &mockOrderService{
		listFn: func(ctx context.Context) ([]models.Order, error) {
			return orders, nil
		},
	}

	handler := NewOrderHandler(mockService)

	req := httptest.NewRequest("GET", "/api/orders", nil)
	w := httptest.NewRecorder()

	handler.ListOrders(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var result []models.Order
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(result) != 2 || result[0].ID != "order-1" || result[1].ID != "order-2" {
		t.Errorf("Unexpected orders response: %+v", result)
	}
}

func TestListOrders_Error(t *testing.T) {
	mockService := &mockOrderService{
		listFn: func(ctx context.Context) ([]models.Order, error) {
			return nil, errors.New("query failed")
		},
	}

	handler := NewOrderHandler(mockService)

	req := httptest.NewRequest("GET", "/api/orders", nil)
	w := httptest.NewRecorder()

	handler.ListOrders(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}
