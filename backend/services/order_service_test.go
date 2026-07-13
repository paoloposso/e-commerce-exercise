package services

import (
	"context"
	"errors"
	"testing"

	"ecommerce/backend/models"
	"ecommerce/backend/payment"
)

type mockOrderRepository struct {
	getBySKUFn              func(ctx context.Context, sku string) (*models.Product, error)
	getOrderByIdempotencyFn func(ctx context.Context, key string) (*models.Order, error)
	tryDecrementStockFn     func(ctx context.Context, sku string, quantity, expectedVersion int) (bool, error)
	restoreStockFn          func(ctx context.Context, sku string, quantity int) error
	createOrderFn           func(ctx context.Context, order *models.Order) error
	listOrdersFn            func(ctx context.Context) ([]models.Order, error)
}

func (m *mockOrderRepository) GetBySKU(ctx context.Context, sku string) (*models.Product, error) {
	return m.getBySKUFn(ctx, sku)
}
func (m *mockOrderRepository) GetOrderByIdempotencyKey(ctx context.Context, key string) (*models.Order, error) {
	if m.getOrderByIdempotencyFn != nil {
		return m.getOrderByIdempotencyFn(ctx, key)
	}
	return nil, nil
}
func (m *mockOrderRepository) TryDecrementStock(ctx context.Context, sku string, quantity, expectedVersion int) (bool, error) {
	if m.tryDecrementStockFn != nil {
		return m.tryDecrementStockFn(ctx, sku, quantity, expectedVersion)
	}
	return true, nil
}
func (m *mockOrderRepository) RestoreStock(ctx context.Context, sku string, quantity int) error {
	if m.restoreStockFn != nil {
		return m.restoreStockFn(ctx, sku, quantity)
	}
	return nil
}
func (m *mockOrderRepository) CreateOrder(ctx context.Context, order *models.Order) error {
	if m.createOrderFn != nil {
		return m.createOrderFn(ctx, order)
	}
	return nil
}
func (m *mockOrderRepository) ListOrders(ctx context.Context) ([]models.Order, error) {
	if m.listOrdersFn != nil {
		return m.listOrdersFn(ctx)
	}
	return nil, nil
}

func defaultBroker() payment.Broker {
	return &payment.MockBroker{}
}

func TestPurchaseProduct_PaymentDeclined(t *testing.T) {
	product := &models.Product{ID: "p1", SKU: "SKU-001", Price: 10.00, Stock: 5, Version: 1}

	restoreCallCount := 0
	mockRepo := &mockOrderRepository{
		getBySKUFn: func(ctx context.Context, sku string) (*models.Product, error) {
			return product, nil
		},
		tryDecrementStockFn: func(ctx context.Context, sku string, quantity, expectedVersion int) (bool, error) {
			return true, nil
		},
		restoreStockFn: func(ctx context.Context, sku string, quantity int) error {
			restoreCallCount++
			return nil
		},
	}

	declinedBroker := &payment.MockBroker{
		ChargeFunc: func(ctx context.Context, req payment.Request) (*payment.Result, error) {
			return nil, payment.ErrDeclined
		},
	}

	service := NewOrderService(mockRepo, declinedBroker)
	_, err := service.PurchaseProduct(context.Background(), "SKU-001", "customer-1", "idem-key-decline", 2, 10.00)
	if !errors.Is(err, ErrPaymentDeclined) {
		t.Errorf("Expected ErrPaymentDeclined, got %v", err)
	}
	if restoreCallCount != 1 {
		t.Errorf("Expected RestoreStock to be called once, got %d", restoreCallCount)
	}
}

func TestPurchaseProduct_CASConflictExhaustRetries(t *testing.T) {
	product := &models.Product{ID: "p1", SKU: "SKU-001", Price: 10.00, Stock: 5, Version: 1}

	attempts := 0
	mockRepo := &mockOrderRepository{
		getBySKUFn: func(ctx context.Context, sku string) (*models.Product, error) {
			return product, nil
		},
		tryDecrementStockFn: func(ctx context.Context, sku string, quantity, expectedVersion int) (bool, error) {
			attempts++
			return false, nil
		},
	}

	service := NewOrderService(mockRepo, defaultBroker())
	_, err := service.PurchaseProduct(context.Background(), "SKU-001", "customer-1", "idem-key-cas", 2, 10.00)
	if !errors.Is(err, ErrConcurrencyConflict) {
		t.Errorf("Expected ErrConcurrencyConflict, got %v", err)
	}
	if attempts != maxPurchaseRetries {
		t.Errorf("Expected %d CAS attempts, got %d", maxPurchaseRetries, attempts)
	}
}

func TestPurchaseProduct_InsufficientStock(t *testing.T) {
	product := &models.Product{ID: "p1", SKU: "SKU-001", Price: 10.00, Stock: 1, Version: 0}

	mockRepo := &mockOrderRepository{
		getBySKUFn: func(ctx context.Context, sku string) (*models.Product, error) {
			return product, nil
		},
	}

	service := NewOrderService(mockRepo, defaultBroker())
	_, err := service.PurchaseProduct(context.Background(), "SKU-001", "customer-1", "idem-key-nostock", 5, 10.00)
	if !errors.Is(err, ErrInsufficientStock) {
		t.Errorf("Expected ErrInsufficientStock, got %v", err)
	}
}

func TestPurchaseProduct_PriceChanged(t *testing.T) {
	product := &models.Product{ID: "p1", SKU: "SKU-001", Price: 15.00, Stock: 5, Version: 1}

	mockRepo := &mockOrderRepository{
		getBySKUFn: func(ctx context.Context, sku string) (*models.Product, error) {
			return product, nil
		},
	}

	service := NewOrderService(mockRepo, defaultBroker())
	_, err := service.PurchaseProduct(context.Background(), "SKU-001", "customer-1", "idem-key-price", 2, 10.00)
	if !errors.Is(err, ErrPriceChanged) {
		t.Errorf("Expected ErrPriceChanged, got %v", err)
	}
}

func TestPurchaseProduct_IdempotencyReplay(t *testing.T) {
	existingOrder := &models.Order{ID: "order-abc", SKU: "SKU-001", Quantity: 2, IdempotencyKey: "idem-key-replay"}

	mockRepo := &mockOrderRepository{
		getBySKUFn: func(ctx context.Context, sku string) (*models.Product, error) {
			return nil, nil
		},
		getOrderByIdempotencyFn: func(ctx context.Context, key string) (*models.Order, error) {
			return existingOrder, nil
		},
	}

	service := NewOrderService(mockRepo, defaultBroker())
	order, err := service.PurchaseProduct(context.Background(), "SKU-001", "customer-1", "idem-key-replay", 2, 10.00)
	if err != nil {
		t.Fatalf("Expected nil error on idempotent replay, got %v", err)
	}
	if order.ID != existingOrder.ID {
		t.Errorf("Expected same order ID %q, got %q", existingOrder.ID, order.ID)
	}
}

func TestPurchaseProduct_CreateOrderFailRefundSuccess(t *testing.T) {
	product := &models.Product{ID: "p1", SKU: "SKU-001", Price: 10.00, Stock: 5, Version: 1}
	dbErr := errors.New("database disk full")

	restoreCalled := false
	mockRepo := &mockOrderRepository{
		getBySKUFn: func(ctx context.Context, sku string) (*models.Product, error) {
			return product, nil
		},
		tryDecrementStockFn: func(ctx context.Context, sku string, quantity, expectedVersion int) (bool, error) {
			return true, nil
		},
		createOrderFn: func(ctx context.Context, order *models.Order) error {
			return dbErr
		},
		restoreStockFn: func(ctx context.Context, sku string, quantity int) error {
			restoreCalled = true
			return nil
		},
	}

	refundCalled := false
	mockBroker := &payment.MockBroker{
		RefundFunc: func(ctx context.Context, transactionID string) error {
			refundCalled = true
			if transactionID == "" {
				t.Error("Expected non-empty transaction ID for refund")
			}
			return nil
		},
	}

	service := NewOrderService(mockRepo, mockBroker)
	_, err := service.PurchaseProduct(context.Background(), "SKU-001", "customer-1", "idem-key-fail-create", 2, 10.00)
	if err == nil {
		t.Fatal("Expected purchase to fail, but it succeeded")
	}

	if !errors.Is(err, ErrRefundedOrderRecordFailed) {
		t.Errorf("Expected error to wrap ErrRefundedOrderRecordFailed, got %v", err)
	}

	if !errors.Is(err, dbErr) {
		t.Errorf("Expected error to wrap underlying database error %v, got %v", dbErr, err)
	}

	if !restoreCalled {
		t.Error("Expected RestoreStock to be called")
	}

	if !refundCalled {
		t.Error("Expected Refund to be called")
	}
}
