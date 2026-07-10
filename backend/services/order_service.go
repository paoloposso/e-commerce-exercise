package services

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"ecommerce/backend/models"
	"ecommerce/backend/payment"
)

const maxPurchaseRetries = 3

var (
	ErrPaymentDeclined     = errors.New("payment declined")
	ErrConcurrencyConflict = errors.New("purchase failed due to concurrent stock updates — please try again")
	ErrPriceChanged        = errors.New("price has changed since last viewed")
)

var errCASConflict = errors.New("cas conflict: product version changed during checkout")

type OrderService struct {
	repo    OrderRepository
	payment payment.Broker
}

func NewOrderService(repo OrderRepository, broker payment.Broker) *OrderService {
	return &OrderService{repo: repo, payment: broker}
}

func (s *OrderService) PurchaseProduct(ctx context.Context, sku, customerID, idempotencyKey string, quantity int, expectedPrice float64) (*models.Order, error) {
	sku = strings.TrimSpace(sku)
	idempotencyKey = strings.TrimSpace(idempotencyKey)

	if sku == "" {
		return nil, fmt.Errorf("%w: SKU is required", ErrInvalidInput)
	}
	if quantity <= 0 {
		return nil, fmt.Errorf("%w: quantity must be greater than zero", ErrInvalidInput)
	}
	if idempotencyKey == "" {
		return nil, fmt.Errorf("%w: idempotency_key is required to prevent duplicate orders", ErrInvalidInput)
	}

	existing, err := s.repo.GetOrderByIdempotencyKey(ctx, idempotencyKey)
	if err != nil {
		return nil, fmt.Errorf("failed to check idempotency key: %w", err)
	}
	if existing != nil {
		return existing, nil
	}

	for attempt := 1; attempt <= maxPurchaseRetries; attempt++ {
		order, err := s.tryPurchase(ctx, sku, customerID, idempotencyKey, quantity, expectedPrice)
		if err == nil {
			return order, nil
		}

		if errors.Is(err, ErrInsufficientStock) || errors.Is(err, ErrPaymentDeclined) || errors.Is(err, ErrNotFound) || errors.Is(err, ErrPriceChanged) {
			return nil, err
		}

		if errors.Is(err, errCASConflict) {
			if attempt < maxPurchaseRetries {
				select {
				case <-ctx.Done():
					return nil, ctx.Err()
				case <-time.After(time.Duration(attempt*50) * time.Millisecond):
					continue
				}
			}
			return nil, ErrConcurrencyConflict
		}

		return nil, err
	}

	return nil, ErrConcurrencyConflict
}

func (s *OrderService) tryPurchase(ctx context.Context, sku, customerID, idempotencyKey string, quantity int, expectedPrice float64) (*models.Order, error) {
	product, err := s.repo.GetBySKU(ctx, sku)
	if err != nil {
		return nil, err
	}
	if product == nil {
		return nil, ErrNotFound
	}

	if product.Price != expectedPrice {
		return nil, fmt.Errorf("%w: expected %.2f, got %.2f", ErrPriceChanged, expectedPrice, product.Price)
	}

	if product.Stock < quantity {
		return nil, fmt.Errorf("%w: requested %d, available %d", ErrInsufficientStock, quantity, product.Stock)
	}

	ok, err := s.repo.TryDecrementStock(ctx, sku, quantity, product.Version)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, errCASConflict
	}

	totalPrice := product.Price * float64(quantity)
	orderID := GenerateUUID()

	payResult, err := s.payment.Charge(ctx, payment.Request{
		OrderID:        orderID,
		SKU:            sku,
		Quantity:       quantity,
		TotalPrice:     totalPrice,
		IdempotencyKey: idempotencyKey,
	})

	if err != nil || (payResult != nil && payResult.Status != "approved") {
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = s.repo.RestoreStock(cleanupCtx, sku, quantity)

		if errors.Is(err, payment.ErrDeclined) || (payResult != nil && payResult.Status == "declined") {
			return nil, ErrPaymentDeclined
		}
		if err != nil {
			return nil, fmt.Errorf("payment error: %w", err)
		}
		return nil, ErrPaymentDeclined
	}

	order := &models.Order{
		ID:             orderID,
		CustomerID:     customerID,
		SKU:            sku,
		Quantity:       quantity,
		TotalPrice:     totalPrice,
		IdempotencyKey: idempotencyKey,
		CreatedAt:      time.Now(),
	}

	finalizationCtx, finalCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer finalCancel()

	if err := s.repo.CreateOrder(finalizationCtx, order); err != nil {
		_ = s.repo.RestoreStock(finalizationCtx, sku, quantity)
		return nil, fmt.Errorf("failed to record order after payment: %w", err)
	}

	return order, nil
}

func (s *OrderService) ListOrders(ctx context.Context) ([]models.Order, error) {
	return s.repo.ListOrders(ctx)
}
