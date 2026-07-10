package services

import (
	"bytes"
	"context"
	"testing"

	"ntd/backend/models"
)

type mockImporterRepository struct {
	getBySKUFn func(ctx context.Context, sku string) (*models.Product, error)
	createFn   func(ctx context.Context, p *models.Product) error
	updateFn   func(ctx context.Context, id string, p *models.Product) error
}

func (m *mockImporterRepository) GetBySKU(ctx context.Context, sku string) (*models.Product, error) {
	return m.getBySKUFn(ctx, sku)
}
func (m *mockImporterRepository) Create(ctx context.Context, p *models.Product) error {
	return m.createFn(ctx, p)
}
func (m *mockImporterRepository) Update(ctx context.Context, id string, p *models.Product) error {
	return m.updateFn(ctx, id, p)
}

func TestImportProductsFromCSV_Mocked(t *testing.T) {

	createdCount := 0
	updatedCount := 0

	mockRepo := &mockImporterRepository{
		getBySKUFn: func(ctx context.Context, sku string) (*models.Product, error) {
			if sku == "SKU-NEW" {
				return nil, nil
			}
			if sku == "SKU-DUP" {
				return &models.Product{ID: "existing-uuid-1", SKU: "SKU-DUP"}, nil
			}
			return nil, nil
		},
		createFn: func(ctx context.Context, p *models.Product) error {
			createdCount++
			if p.SKU != "SKU-NEW" {
				t.Errorf("Expected Create to be triggered for SKU-NEW, got %s", p.SKU)
			}
			return nil
		},
		updateFn: func(ctx context.Context, id string, p *models.Product) error {
			updatedCount++
			if id != "existing-uuid-1" || p.SKU != "SKU-DUP" {
				t.Errorf("Expected Update to be triggered for SKU-DUP with ID existing-uuid-1, got ID %s SKU %s", id, p.SKU)
			}
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

	if report.ImportedRows != 1 || createdCount != 1 {
		t.Errorf("Expected 1 imported product, got report: %d, create calls: %d", report.ImportedRows, createdCount)
	}

	if report.UpdatedRows != 1 || updatedCount != 1 {
		t.Errorf("Expected 1 updated product, got report: %d, update calls: %d", report.UpdatedRows, updatedCount)
	}

	if len(report.Errors) != 0 {
		t.Errorf("Expected 0 errors, got %d: %+v", len(report.Errors), report.Errors)
	}
}
