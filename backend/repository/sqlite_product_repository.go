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

	if err := rows.Err(); err != nil {
		return nil, err
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

func (r *SQLiteProductRepository) GetBySKUs(ctx context.Context, skus []string) ([]models.Product, error) {
	if len(skus) == 0 {
		return nil, nil
	}

	chunkSize := 990
	var allProducts []models.Product
	for i := 0; i < len(skus); i += chunkSize {
		end := min(i+chunkSize, len(skus))
		chunk := skus[i:end]

		placeholders := make([]string, len(chunk))
		args := make([]any, len(chunk))
		for j, sku := range chunk {
			placeholders[j] = "?"
			args[j] = sku
		}

		query := fmt.Sprintf(
			"SELECT id, name, sku, description, category, price, stock, weight_kg, version FROM products WHERE sku IN (%s)",
			strings.Join(placeholders, ","),
		)

		rows, err := r.db.QueryContext(ctx, query, args...)
		if err != nil {
			return nil, err
		}

		err = func() error {
			defer rows.Close()
			for rows.Next() {
				var p models.Product
				err := rows.Scan(&p.ID, &p.Name, &p.SKU, &p.Description, &p.Category, &p.Price, &p.Stock, &p.WeightKg, &p.Version)
				if err != nil {
					return err
				}
				allProducts = append(allProducts, p)
			}
			return rows.Err()
		}()
		if err != nil {
			return nil, err
		}
	}

	return allProducts, nil
}

func (r *SQLiteProductRepository) BulkCreate(ctx context.Context, products []*models.Product) error {
	if len(products) == 0 {
		return nil
	}

	batchSize := 100
	for i := 0; i < len(products); i += batchSize {
		end := min(i+batchSize, len(products))
		batch := products[i:end]

		var query strings.Builder
		query.WriteString("INSERT INTO products (id, name, sku, description, category, price, stock, weight_kg) VALUES ")
		vals := []any{}
		for j, p := range batch {
			if p.ID == "" {
				p.ID = generateUUID()
			}
			if j > 0 {
				query.WriteString(", ")
			}
			query.WriteString("(?, ?, ?, ?, ?, ?, ?, ?)")
			vals = append(vals, p.ID, p.Name, p.SKU, p.Description, p.Category, p.Price, p.Stock, p.WeightKg)
		}

		_, err := r.db.ExecContext(ctx, query.String(), vals...)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *SQLiteProductRepository) ImportProducts(ctx context.Context, products []models.Product) (int, int, error) {
	if len(products) == 0 {
		return 0, 0, nil
	}

	tx, err := r.beginTx(ctx)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Extract unique SKUs
	var uniqueSKUs []string
	uniqueSKUsMap := make(map[string]bool)
	for _, p := range products {
		if !uniqueSKUsMap[p.SKU] {
			uniqueSKUsMap[p.SKU] = true
			uniqueSKUs = append(uniqueSKUs, p.SKU)
		}
	}

	// Fetch existing products
	existingList, err := tx.GetBySKUs(ctx, uniqueSKUs)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to fetch existing products: %w", err)
	}

	existingMap := make(map[string]*models.Product)
	for i := range existingList {
		existingMap[existingList[i].SKU] = &existingList[i]
	}

	// Classify and simulate row-by-row counts
	seenSKUs := make(map[string]bool)
	finalProducts := make(map[string]models.Product)
	var insertSKUs []string
	var updateSKUs []string

	importedCount := 0
	updatedCount := 0

	for _, p := range products {
		sku := p.SKU
		existing, inDB := existingMap[sku]

		if !seenSKUs[sku] {
			seenSKUs[sku] = true
			if inDB {
				p.ID = existing.ID
				finalProducts[sku] = p
				updateSKUs = append(updateSKUs, sku)
				updatedCount++
			} else {
				p.ID = generateUUID()
				finalProducts[sku] = p
				insertSKUs = append(insertSKUs, sku)
				importedCount++
			}
		} else {
			p.ID = finalProducts[sku].ID
			finalProducts[sku] = p
			updatedCount++
		}
	}

	// Bulk Create new products
	var productsToInsert []*models.Product
	for _, sku := range insertSKUs {
		p := finalProducts[sku]
		productsToInsert = append(productsToInsert, &p)
	}
	if len(productsToInsert) > 0 {
		err = tx.BulkCreate(ctx, productsToInsert)
		if err != nil {
			return 0, 0, fmt.Errorf("bulk insert failed: %w", err)
		}
	}

	// Update existing products
	for _, sku := range updateSKUs {
		p := finalProducts[sku]
		err = tx.Update(ctx, p.ID, &p)
		if err != nil {
			return 0, 0, fmt.Errorf("update failed for sku %s: %w", sku, err)
		}
	}

	// Commit Transaction
	err = tx.Commit()
	if err != nil {
		return 0, 0, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return importedCount, updatedCount, nil
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
