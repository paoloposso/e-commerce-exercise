package repository

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/glebarez/go-sqlite"
)

func ConnectDB(dbPath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open SQLite database: %w", err)
	}

	if err = db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("failed to ping SQLite database: %w", err)
	}

	log.Printf("Successfully connected to SQLite database at: %s", dbPath)

	err = initializeSchema(db)
	if err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("failed to initialize SQLite schema: %w", err)
	}

	return db, nil
}

func initializeSchema(db *sql.DB) error {
	productsTable := `
	CREATE TABLE IF NOT EXISTS products (
		id          TEXT PRIMARY KEY,
		name        TEXT NOT NULL,
		sku         TEXT UNIQUE NOT NULL,
		description TEXT,
		category    TEXT,
		price       REAL NOT NULL,
		stock       INTEGER NOT NULL,
		weight_kg   REAL NOT NULL,
		version     INTEGER NOT NULL DEFAULT 0
	);`

	_, err := db.Exec(productsTable)
	if err != nil {
		return fmt.Errorf("error creating products table: %w", err)
	}

	ordersTable := `
	CREATE TABLE IF NOT EXISTS orders (
		id              TEXT PRIMARY KEY,
		customer_id     TEXT NOT NULL DEFAULT '',
		sku             TEXT NOT NULL,
		quantity        INTEGER NOT NULL,
		total_price     REAL NOT NULL,
		idempotency_key TEXT UNIQUE NOT NULL,
		created_at      DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	_, err = db.Exec(ordersTable)
	if err != nil {
		return fmt.Errorf("error creating orders table: %w", err)
	}

	log.Println("Database schemas initialized successfully (products & orders tables created)")
	return nil
}
