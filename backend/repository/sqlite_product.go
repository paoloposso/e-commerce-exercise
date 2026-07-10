package repository

import (
	"context"
	"crypto/rand"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"ntd/backend/models"
)

type SQLiteProductRepository struct {
	db *sql.DB
}

// NewSQLiteProductRepository constructs a concrete SQLiteProductRepository utilizing SQLite storage.
func NewSQLiteProductRepository(db *sql.DB) *SQLiteProductRepository {
	return &SQLiteProductRepository{db: db}
}

// Helper to generate pseudo-UUIDv4 inside the repository layer
func generateUUID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

func (r *SQLiteProductRepository) List(ctx context.Context, query, category string) ([]models.Product, error) {
	queryStr := "SELECT id, name, sku, description, category, price, stock, weight_kg, version FROM products WHERE 1=1"
	var args []interface{}

	if category != "" {
		queryStr += " AND category = ?"
		args = append(args, category)
	}

	if query != "" {
		queryStr += " AND (name LIKE ? OR sku LIKE ? OR description LIKE ?)"
		likeQuery := "%" + query + "%"
		args = append(args, likeQuery, likeQuery, likeQuery)
	}

	rows, err := r.db.QueryContext(ctx, queryStr, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	products := []models.Product{}
	for rows.Next() {
		var p models.Product
		err := rows.Scan(&p.ID, &p.Name, &p.SKU, &p.Description, &p.Category, &p.Price, &p.Stock, &p.WeightKg, &p.Version)
		if err != nil {
			return nil, err
		}
		products = append(products, p)
	}

	return products, nil
}

func (r *SQLiteProductRepository) GetByID(ctx context.Context, id string) (*models.Product, error) {
	var p models.Product
	err := r.db.QueryRowContext(ctx, "SELECT id, name, sku, description, category, price, stock, weight_kg, version FROM products WHERE id = ?", id).
		Scan(&p.ID, &p.Name, &p.SKU, &p.Description, &p.Category, &p.Price, &p.Stock, &p.WeightKg, &p.Version)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &p, nil
}

func (r *SQLiteProductRepository) GetBySKU(ctx context.Context, sku string) (*models.Product, error) {
	var p models.Product
	err := r.db.QueryRowContext(ctx, "SELECT id, name, sku, description, category, price, stock, weight_kg, version FROM products WHERE sku = ?", sku).
		Scan(&p.ID, &p.Name, &p.SKU, &p.Description, &p.Category, &p.Price, &p.Stock, &p.WeightKg, &p.Version)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &p, nil
}

func (r *SQLiteProductRepository) Create(ctx context.Context, p *models.Product) error {
	if p.ID == "" {
		p.ID = generateUUID()
	}
	_, err := r.db.ExecContext(ctx,
		"INSERT INTO products (id, name, sku, description, category, price, stock, weight_kg) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		p.ID, p.Name, p.SKU, p.Description, p.Category, p.Price, p.Stock, p.WeightKg)
	return err
}

func (r *SQLiteProductRepository) Update(ctx context.Context, id string, p *models.Product) error {
	result, err := r.db.ExecContext(ctx,
		"UPDATE products SET name = ?, sku = ?, description = ?, category = ?, price = ?, stock = ?, weight_kg = ? WHERE id = ?",
		p.Name, p.SKU, p.Description, p.Category, p.Price, p.Stock, p.WeightKg, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *SQLiteProductRepository) Delete(ctx context.Context, id string) error {
	result, err := r.db.ExecContext(ctx, "DELETE FROM products WHERE id = ?", id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// GetOrderByIdempotencyKey returns an existing order for the given key, or (nil, nil) if not found.
func (r *SQLiteProductRepository) GetOrderByIdempotencyKey(ctx context.Context, idempotencyKey string) (*models.Order, error) {
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

// TryDecrementStock performs a Compare-And-Swap stock decrement.
// It updates the row only when the current version matches expectedVersion.
// Returns (true, nil) on success, (false, nil) on a version conflict (retry), or (false, err) on a database error.
func (r *SQLiteProductRepository) TryDecrementStock(ctx context.Context, sku string, quantity, expectedVersion int) (bool, error) {
	result, err := r.db.ExecContext(ctx,
		"UPDATE products SET stock = stock - ?, version = version + 1 WHERE sku = ? AND version = ? AND stock >= ?",
		quantity, sku, expectedVersion, quantity,
	)
	if err != nil {
		return false, fmt.Errorf("failed to decrement stock: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, err
	}
	return rowsAffected == 1, nil
}

// RestoreStock increments stock back by quantity. Called when payment fails after a successful decrement.
func (r *SQLiteProductRepository) RestoreStock(ctx context.Context, sku string, quantity int) error {
	_, err := r.db.ExecContext(ctx,
		"UPDATE products SET stock = stock + ?, version = version + 1 WHERE sku = ?",
		quantity, sku,
	)
	return err
}

// CreateOrder inserts a completed order record into the database.
func (r *SQLiteProductRepository) CreateOrder(ctx context.Context, order *models.Order) error {
	if order.ID == "" {
		order.ID = generateUUID()
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
