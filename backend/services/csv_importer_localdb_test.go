package services

import (
	"bytes"
	"context"
	"os"
	"strings"
	"testing"

	"ecommerce/backend/repository"
)

func TestImportProductsFromCSV_Integration(t *testing.T) {
	testDBPath := "test_ecommerce.db"
	dbHandle, err := repository.ConnectDB(testDBPath)
	if err != nil {
		t.Fatalf("Failed to initialize test SQLite DB: %v", err)
	}
	defer func() {
		_ = dbHandle.Close()
		_ = os.Remove(testDBPath)
	}()

	productRepo := repository.NewSQLiteProductRepository(dbHandle)

	csvData := `name,sku,description,category,price,stock,weight_kg
Standard Product,SKU-001,Standard item description,Electronics,99.99,10,1.5
Product with Symbol,SKU-002,Item priced with dollar symbol,Home,$45.50,150,0.85
Free Product,SKU-003,This is a free item,Gifts,free,250,0.05
Product with Invalid Price,SKU-004,This price format is invalid,Clothing,abc,20,0.2
Product with Negative Stock,SKU-005,Negative inventory check,Sports,10.00,-10,0.45
Product with Empty Name,SKU-006,This product name is empty,,12.50,5,0.3
,SKU-007,Missing name completely,Tools,8.00,2,0.1
Product with Missing SKU, ,No sku value provided,Outdoors,150.00,8,4.2
`

	buf := bytes.NewBufferString(csvData)
	report, err := ImportProductsFromCSV(context.Background(), buf, productRepo)
	if err != nil {
		t.Fatalf("ImportProductsFromCSV failed: %v", err)
	}

	if report.TotalRows != 8 {
		t.Errorf("Expected 8 total rows, got %d", report.TotalRows)
	}

	expectedImported := 4
	if report.ImportedRows != expectedImported {
		t.Errorf("Expected %d imported products, got %d", expectedImported, report.ImportedRows)
	}

	expectedErrors := 4
	if len(report.Errors) != expectedErrors {
		t.Errorf("Expected %d errors, got %d", expectedErrors, len(report.Errors))
	}

	errorChecks := map[int]string{
		5: "Invalid price format",
		6: "Stock quantity cannot be negative",
		8: "Product Name is required",
		9: "SKU is a required field",
	}

	for _, e := range report.Errors {
		expectedMsg, exists := errorChecks[e.RowNumber]
		if !exists {
			t.Errorf("Unexpected error reported on row %d: %s", e.RowNumber, e.Error)
			continue
		}
		if !strings.Contains(e.Error, expectedMsg) {
			t.Errorf("Expected row %d error to contain '%s', but got '%s'", e.RowNumber, expectedMsg, e.Error)
		}
	}
}
