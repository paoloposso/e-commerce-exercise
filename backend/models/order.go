package models

import "time"

type Order struct {
	ID             string    `json:"id"`
	CustomerID     string    `json:"customer_id"`
	SKU            string    `json:"sku"`
	Quantity       int       `json:"quantity"`
	TotalPrice     float64   `json:"total_price"`
	IdempotencyKey string    `json:"idempotency_key"`
	CreatedAt      time.Time `json:"created_at"`
}
