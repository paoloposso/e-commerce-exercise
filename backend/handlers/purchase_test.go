package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"ecommerce/backend/models"
	"ecommerce/backend/payment"
	"ecommerce/backend/repository"
	"ecommerce/backend/services"
)

type mockOrderService struct {
	purchaseFn func(ctx context.Context, sku, customerID, idempotencyKey string, quantity int, expectedPrice float64) (*models.Order, error)
}

func (m *mockOrderService) PurchaseProduct(ctx context.Context, sku, customerID, idempotencyKey string, quantity int, expectedPrice float64) (*models.Order, error) {
	return m.purchaseFn(ctx, sku, customerID, idempotencyKey, quantity, expectedPrice)
}

func TestPurchaseProduct_Integration(t *testing.T) {
	testDBPath := "test_purchase.db"
	dbHandle, err := repository.ConnectDB(testDBPath)
	if err != nil {
		t.Fatalf("Failed to initialize test SQLite DB: %v", err)
	}
	defer func() {
		_ = dbHandle.Close()
		_ = os.Remove(testDBPath)
	}()

	productRepo := repository.NewSQLiteProductRepository(dbHandle)
	orderService := services.NewOrderService(productRepo, &payment.MockBroker{})
	orderHandler := NewOrderHandler(orderService)

	product := models.Product{
		ID:          "p-test-1",
		Name:        "Test Product",
		SKU:         "SKU-TEST-1",
		Description: "A product for purchase tests",
		Category:    "Test",
		Price:       25.50,
		Stock:       3,
		WeightKg:    0.5,
	}
	_, err = dbHandle.Exec("INSERT INTO products (id, name, sku, description, category, price, stock, weight_kg) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		product.ID, product.Name, product.SKU, product.Description, product.Category, product.Price, product.Stock, product.WeightKg)
	if err != nil {
		t.Fatalf("Failed to seed product: %v", err)
	}

	reqBody := PurchaseRequest{
		SKU:            "SKU-TEST-1",
		CustomerID:     "customer-uuid-001",
		Quantity:       2,
		ExpectedPrice:  25.50,
		IdempotencyKey: "idem-key-001",
	}
	bodyJSON, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/api/purchase", bytes.NewBuffer(bodyJSON))
	rr := httptest.NewRecorder()

	orderHandler.PurchaseProduct(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("Expected status code %d, got %d — body: %s", http.StatusCreated, rr.Code, rr.Body.String())
	}

	var order models.Order
	err = json.Unmarshal(rr.Body.Bytes(), &order)
	if err != nil {
		t.Fatalf("Failed to parse order response: %v", err)
	}

	if order.SKU != "SKU-TEST-1" || order.Quantity != 2 || order.TotalPrice != 51.00 {
		t.Errorf("Order details mismatch: %+v", order)
	}
	if order.CustomerID != "customer-uuid-001" {
		t.Errorf("Expected customer_id 'customer-uuid-001', got %q", order.CustomerID)
	}

	var remainingStock int
	err = dbHandle.QueryRow("SELECT stock FROM products WHERE sku = ?", "SKU-TEST-1").Scan(&remainingStock)
	if err != nil {
		t.Fatalf("Failed to check remaining stock: %v", err)
	}
	if remainingStock != 1 {
		t.Errorf("Expected stock to be 1, got %d", remainingStock)
	}

	bodyJSONReplay, _ := json.Marshal(reqBody)
	reqReplay, _ := http.NewRequest("POST", "/api/purchase", bytes.NewBuffer(bodyJSONReplay))
	rrReplay := httptest.NewRecorder()

	orderHandler.PurchaseProduct(rrReplay, reqReplay)

	if rrReplay.Code != http.StatusCreated {
		t.Errorf("Idempotent replay: expected status %d, got %d", http.StatusCreated, rrReplay.Code)
	}

	var replayOrder models.Order
	_ = json.Unmarshal(rrReplay.Body.Bytes(), &replayOrder)
	if replayOrder.ID != order.ID {
		t.Errorf("Idempotent replay: expected same order ID %q, got %q", order.ID, replayOrder.ID)
	}

	err = dbHandle.QueryRow("SELECT stock FROM products WHERE sku = ?", "SKU-TEST-1").Scan(&remainingStock)
	if err != nil {
		t.Fatalf("Failed to check remaining stock after replay: %v", err)
	}
	if remainingStock != 1 {
		t.Errorf("Idempotent replay: expected stock to remain 1, got %d", remainingStock)
	}

	reqBodyExceed := PurchaseRequest{SKU: "SKU-TEST-1", CustomerID: "customer-uuid-001", Quantity: 2, ExpectedPrice: 25.50, IdempotencyKey: "idem-key-002"}
	bodyJSONExceed, _ := json.Marshal(reqBodyExceed)
	reqExceed, _ := http.NewRequest("POST", "/api/purchase", bytes.NewBuffer(bodyJSONExceed))
	rrExceed := httptest.NewRecorder()

	orderHandler.PurchaseProduct(rrExceed, reqExceed)

	if rrExceed.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, rrExceed.Code)
	}

	reqBodyNoKey := PurchaseRequest{SKU: "SKU-TEST-1", CustomerID: "customer-uuid-001", Quantity: 1, ExpectedPrice: 25.50}
	bodyJSONNoKey, _ := json.Marshal(reqBodyNoKey)
	reqNoKey, _ := http.NewRequest("POST", "/api/purchase", bytes.NewBuffer(bodyJSONNoKey))
	rrNoKey := httptest.NewRecorder()

	orderHandler.PurchaseProduct(rrNoKey, reqNoKey)

	if rrNoKey.Code != http.StatusBadRequest {
		t.Errorf("Missing idempotency key: expected status %d, got %d", http.StatusBadRequest, rrNoKey.Code)
	}
}
