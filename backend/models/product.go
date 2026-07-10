package models

// Product represents the product entity in SQLite database.
type Product struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	SKU         string  `json:"sku"`
	Description string  `json:"description"`
	Category    string  `json:"category"`
	Price       float64 `json:"price"`
	Stock       int     `json:"stock"`
	WeightKg    float64 `json:"weight_kg"`
	// Version is an internal optimistic lock counter and is not exposed via the API.
	Version int `json:"-"`
}
