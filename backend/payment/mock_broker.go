package payment

import (
	"context"
	"log"
)

type MockBroker struct {
	ChargeFunc func(ctx context.Context, req Request) (*Result, error)
	RefundFunc func(ctx context.Context, transactionID string) error
}

func (m *MockBroker) Charge(ctx context.Context, req Request) (*Result, error) {
	if m.ChargeFunc != nil {
		return m.ChargeFunc(ctx, req)
	}

	log.Printf("[MockPaymentBroker] Approved charge: order=%s sku=%s qty=%d total=%.2f",
		req.OrderID, req.SKU, req.Quantity, req.TotalPrice)

	return &Result{
		TransactionID: "mock-txn-" + req.OrderID,
		Status:        "approved",
	}, nil
}

func (m *MockBroker) Refund(ctx context.Context, transactionID string) error {
	if m.RefundFunc != nil {
		return m.RefundFunc(ctx, transactionID)
	}

	log.Printf("[MockPaymentBroker] Refunded transaction: %s", transactionID)
	return nil
}
