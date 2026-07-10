package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"ntd/backend/repository"
	"ntd/backend/services"
)

func main() {
	csvFilePath := flag.String("file", "data/products_example.csv", "Path to the e-commerce CSV data file to seed")
	dbPath := flag.String("db", "data/ecommerce.db", "Path to the SQLite database file")
	flag.Parse()

	log.Printf("Seeder CLI started: DB=%s, CSV=%s", *dbPath, *csvFilePath)

	if _, err := os.Stat(*csvFilePath); os.IsNotExist(err) {
		log.Fatalf("Error: CSV data file not found at '%s'. Check the path parameter.", *csvFilePath)
	}

	dbHandle, err := repository.ConnectDB(*dbPath)
	if err != nil {
		log.Fatalf("Error: Failed to connect to SQLite: %v", err)
	}
	defer dbHandle.Close()

	productRepo := repository.NewSQLiteProductRepository(dbHandle)
	productService := services.NewProductService(productRepo)

	file, err := os.Open(*csvFilePath)
	if err != nil {
		log.Fatalf("Error: Failed to open CSV file: %v", err)
	}
	defer file.Close()

	report, err := productService.ImportProducts(context.Background(), file)
	if err != nil {
		log.Fatalf("Error: CSV import failed: %v", err)
	}

	fmt.Println("\n==================================================")
	fmt.Println("              DATABASE SEED COMPLETE              ")
	fmt.Println("==================================================")
	fmt.Printf("Total Rows Processed : %d\n", report.TotalRows)
	fmt.Printf("New Products Created : %d\n", report.ImportedRows)
	fmt.Printf("Products Updated     : %d\n", report.UpdatedRows)
	fmt.Printf("Validation Failures  : %d\n", len(report.Errors))
	fmt.Println("==================================================")

	if len(report.Errors) > 0 {
		fmt.Println("\nValidation details for failed rows:")
		for _, e := range report.Errors {
			fmt.Printf("  - Row %d (SKU: %s): %s\n", e.RowNumber, e.SKU, e.Error)
		}
		fmt.Println("==================================================")
	}
}
