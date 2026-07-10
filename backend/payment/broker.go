package payment

import "context"

var ErrDeclined = errDeclined("payment declined by provider")

type errDeclined string

func (e errDeclined) Error() string { return string(e) }

type Request struct {
	OrderID        string
	SKU            string
	Quantity       int
	TotalPrice     float64
	IdempotencyKey string
}

type Result struct {
	TransactionID string
	Status        string
}

type Broker interface {
	Charge(ctx context.Context, req Request) (*Result, error)
}
