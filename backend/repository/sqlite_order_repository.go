package repository

import (
	"context"
	"crypto/rand"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"ecommerce/backend/models"
)

type SQLiteOrderRepository struct {
	db dbConn
}

func NewSQLiteOrderRepository(db *sql.DB) *SQLiteOrderRepository {
	return &SQLiteOrderRepository{db: db}
}

func generateUUIDForOrder() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

func (r *SQLiteOrderRepository) GetOrderByIdempotencyKey(ctx context.Context, idempotencyKey string) (*models.Order, error) {
	var o models.Order
	err := r.db.QueryRowContext(ctx,
		"SELECT id, customer_id, sku, quantity, total_price, idempotency_key, created_at FROM orders WHERE idempotency_key = ?",
		idempotencyKey,
	).Scan(&o.ID, &o.CustomerID, &o.SKU, &o.Quantity, &o.TotalPrice, &o.IdempotencyKey, &o.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &o, nil
}

func (r *SQLiteOrderRepository) ListOrders(ctx context.Context) ([]models.Order, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT id, customer_id, sku, quantity, total_price, idempotency_key, created_at FROM orders ORDER BY created_at DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []models.Order
	for rows.Next() {
		var o models.Order
		err := rows.Scan(&o.ID, &o.CustomerID, &o.SKU, &o.Quantity, &o.TotalPrice, &o.IdempotencyKey, &o.CreatedAt)
		if err != nil {
			return nil, err
		}
		orders = append(orders, o)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return orders, nil
}

func (r *SQLiteOrderRepository) CreateOrder(ctx context.Context, order *models.Order) error {
	if order.ID == "" {
		order.ID = generateUUIDForOrder()
	}
	if order.CreatedAt.IsZero() {
		order.CreatedAt = time.Now()
	}
	_, err := r.db.ExecContext(ctx,
		"INSERT INTO orders (id, customer_id, sku, quantity, total_price, idempotency_key, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)",
		order.ID, order.CustomerID, order.SKU, order.Quantity, order.TotalPrice, order.IdempotencyKey, order.CreatedAt,
	)
	return err
}
