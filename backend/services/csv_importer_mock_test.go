package services

import (
	"bytes"
	"context"
	"testing"

	"ecommerce/backend/models"
)

type mockImporterRepository struct {
	importProductsFn func(ctx context.Context, products []models.Product) error
}

func (m *mockImporterRepository) ImportProducts(ctx context.Context, products []models.Product) error {
	return m.importProductsFn(ctx, products)
}

func TestImportProductsFromCSV_Mocked(t *testing.T) {
	var importedProducts []models.Product

	mockRepo := &mockImporterRepository{
		importProductsFn: func(ctx context.Context, products []models.Product) error {
			importedProducts = products
			return nil
		},
	}

	csvContent := `name,sku,description,category,price,stock,weight_kg
New Mock Product,SKU-NEW,New mock description,Mocking,19.99,50,0.5
Duplicate Mock Product,SKU-DUP,Updated mock details,Mocking,29.99,100,1.2
`

	buf := bytes.NewBufferString(csvContent)
	report, err := ImportProductsFromCSV(context.Background(), buf, mockRepo)
	if err != nil {
		t.Fatalf("ImportProductsFromCSV failed under mock context: %v", err)
	}

	if report.TotalRows != 2 {
		t.Errorf("Expected 2 processed rows, got %d", report.TotalRows)
	}

	if len(report.Errors) != 0 {
		t.Errorf("Expected 0 errors, got %d: %+v", len(report.Errors), report.Errors)
	}

	if len(importedProducts) != 2 {
		t.Fatalf("Expected mock repo to receive 2 products, got %d", len(importedProducts))
	}

	if importedProducts[0].SKU != "SKU-NEW" || importedProducts[0].Name != "New Mock Product" {
		t.Errorf("Unexpected product at index 0: %+v", importedProducts[0])
	}

	if importedProducts[1].SKU != "SKU-DUP" || importedProducts[1].Name != "Duplicate Mock Product" {
		t.Errorf("Unexpected product at index 1: %+v", importedProducts[1])
	}
}
