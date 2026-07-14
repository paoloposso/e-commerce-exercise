package repository

import (
	"context"
	"crypto/rand"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"ecommerce/backend/models"
)

type dbConn interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

type SQLiteProductRepository struct {
	db dbConn
}

func NewSQLiteProductRepository(db *sql.DB) *SQLiteProductRepository {
	return &SQLiteProductRepository{db: db}
}

func generateUUID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

func (r *SQLiteProductRepository) List(ctx context.Context, query, category string) ([]models.Product, error) {
	queryStr := "SELECT id, name, sku, description, category, price, stock, weight_kg, version FROM products WHERE 1=1"
	var args []any

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

	return products, rows.Err()
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

func (r *SQLiteProductRepository) RestoreStock(ctx context.Context, sku string, quantity int) error {
	_, err := r.db.ExecContext(ctx,
		"UPDATE products SET stock = stock + ?, version = version + 1 WHERE sku = ?",
		quantity, sku,
	)
	return err
}

type sqliteProductRepositoryTx struct {
	*SQLiteProductRepository
	tx *sql.Tx
}

func (txRepo *sqliteProductRepositoryTx) Commit() error {
	return txRepo.tx.Commit()
}

func (txRepo *sqliteProductRepositoryTx) Rollback() error {
	return txRepo.tx.Rollback()
}

func (r *SQLiteProductRepository) beginTx(ctx context.Context) (*sqliteProductRepositoryTx, error) {
	db, ok := r.db.(*sql.DB)
	if !ok {
		return nil, fmt.Errorf("repository db is not a *sql.DB")
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	return &sqliteProductRepositoryTx{
		SQLiteProductRepository: &SQLiteProductRepository{db: tx},
		tx:                      tx,
	}, nil
}

func (r *SQLiteProductRepository) ImportProducts(ctx context.Context, products []models.Product) error {
	if len(products) == 0 {
		return nil
	}

	tx, err := r.beginTx(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	const paramsPerRow = 8
	const chunkSize = 999 / paramsPerRow

	for i := 0; i < len(products); i += chunkSize {
		end := min(i+chunkSize, len(products))
		batch := products[i:end]

		var sb strings.Builder
		sb.WriteString("INSERT INTO products (id, name, sku, description, category, price, stock, weight_kg) VALUES ")
		vals := make([]any, 0, len(batch)*paramsPerRow)

		for j, p := range batch {
			if j > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString("(?, ?, ?, ?, ?, ?, ?, ?)")
			id := p.ID
			if id == "" {
				id = generateUUID()
			}
			vals = append(vals, id, p.Name, p.SKU, p.Description, p.Category, p.Price, p.Stock, p.WeightKg)
		}

		sb.WriteString(` ON CONFLICT(sku) DO UPDATE SET
			name        = excluded.name,
			description = excluded.description,
			category    = excluded.category,
			price       = excluded.price,
			stock       = excluded.stock,
			weight_kg   = excluded.weight_kg,
			version     = version + 1`)

		_, err := tx.db.ExecContext(ctx, sb.String(), vals...)
		if err != nil {
			return fmt.Errorf("upsert failed: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
