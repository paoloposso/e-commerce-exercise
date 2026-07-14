package services

import (
	"context"
	"crypto/rand"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"

	"ecommerce/backend/models"
)

type RowError struct {
	RowNumber int    `json:"row_number"`
	SKU       string `json:"sku,omitempty"`
	Error     string `json:"error"`
}

type ImportReport struct {
	TotalRows int        `json:"total_rows"`
	Errors    []RowError `json:"errors"`
}

func GenerateUUID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

func ImportProductsFromCSV(ctx context.Context, reader io.Reader, repo ProductImporterStore) (*ImportReport, error) {
	csvReader := csv.NewReader(reader)
	csvReader.LazyQuotes = true

	header, err := csvReader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV header: %w", err)
	}

	headerMap := make(map[string]int)
	for i, h := range header {
		headerMap[strings.ToLower(strings.TrimSpace(h))] = i
	}

	required := []string{"name", "sku", "description", "category", "price", "stock", "weight_kg"}
	for _, req := range required {
		if _, exists := headerMap[req]; !exists {
			return nil, fmt.Errorf("missing required header field: %s", req)
		}
	}

	report := &ImportReport{
		Errors: []RowError{},
	}

	const batchSize = 500
	var batch []models.Product

	flushBatch := func() error {
		if len(batch) == 0 {
			return nil
		}
		if err := repo.ImportProducts(ctx, batch); err != nil {
			return fmt.Errorf("bulk import failed: %w", err)
		}
		batch = batch[:0]
		return nil
	}

	rowNum := 1
	for {
		if err := ctx.Err(); err != nil {
			return report, fmt.Errorf("import aborted: %w", err)
		}

		record, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		rowNum++
		if err != nil {
			if errors.Is(err, csv.ErrFieldCount) && len(record) == 1 && strings.TrimSpace(record[0]) == "" {
				continue
			}
			report.Errors = append(report.Errors, RowError{
				RowNumber: rowNum,
				Error:     fmt.Sprintf("Failed to read row structure: %v", err),
			})
			continue
		}

		report.TotalRows++

		isEmpty := true
		for _, val := range record {
			if strings.TrimSpace(val) != "" {
				isEmpty = false
				break
			}
		}
		if isEmpty {
			report.TotalRows--
			continue
		}

		skuVal := strings.TrimSpace(record[headerMap["sku"]])
		nameVal := strings.TrimSpace(record[headerMap["name"]])
		descVal := strings.TrimSpace(record[headerMap["description"]])
		catVal := strings.TrimSpace(record[headerMap["category"]])
		priceVal := strings.TrimSpace(record[headerMap["price"]])
		stockVal := strings.TrimSpace(record[headerMap["stock"]])
		weightVal := strings.TrimSpace(record[headerMap["weight_kg"]])

		if skuVal == "" {
			report.Errors = append(report.Errors, RowError{
				RowNumber: rowNum,
				Error:     "SKU is a required field and cannot be empty",
			})
			continue
		}

		if nameVal == "" {
			report.Errors = append(report.Errors, RowError{
				RowNumber: rowNum,
				SKU:       skuVal,
				Error:     "Product Name is required and cannot be empty or only whitespace",
			})
			continue
		}

		cleanPriceVal := strings.TrimSpace(strings.ReplaceAll(priceVal, "$", ""))

		var price float64
		if strings.ToLower(cleanPriceVal) == "free" {
			price = 0.0
		} else {
			p, err := strconv.ParseFloat(cleanPriceVal, 64)
			if err != nil {
				report.Errors = append(report.Errors, RowError{
					RowNumber: rowNum,
					SKU:       skuVal,
					Error:     fmt.Sprintf("Invalid price format '%s'", priceVal),
				})
				continue
			}
			if p < 0 {
				report.Errors = append(report.Errors, RowError{
					RowNumber: rowNum,
					SKU:       skuVal,
					Error:     fmt.Sprintf("Price cannot be negative: %f", p),
				})
				continue
			}
			price = p
		}

		s, err := strconv.Atoi(stockVal)
		if err != nil {
			report.Errors = append(report.Errors, RowError{
				RowNumber: rowNum,
				SKU:       skuVal,
				Error:     fmt.Sprintf("Invalid stock format '%s'", stockVal),
			})
			continue
		}
		if s < 0 {
			report.Errors = append(report.Errors, RowError{
				RowNumber: rowNum,
				SKU:       skuVal,
				Error:     fmt.Sprintf("Stock quantity cannot be negative: %d", s),
			})
			continue
		}

		var weight float64
		if weightVal != "" {
			w, err := strconv.ParseFloat(weightVal, 64)
			if err != nil {
				report.Errors = append(report.Errors, RowError{
					RowNumber: rowNum,
					SKU:       skuVal,
					Error:     fmt.Sprintf("Invalid weight format '%s'", weightVal),
				})
				continue
			}
			if w < 0 {
				report.Errors = append(report.Errors, RowError{
					RowNumber: rowNum,
					SKU:       skuVal,
					Error:     fmt.Sprintf("Weight cannot be negative: %f", w),
				})
				continue
			}
			weight = w
		}

		batch = append(batch, models.Product{
			Name:        nameVal,
			SKU:         skuVal,
			Description: descVal,
			Category:    catVal,
			Price:       price,
			Stock:       s,
			WeightKg:    weight,
		})

		if len(batch) >= batchSize {
			if err := flushBatch(); err != nil {
				return nil, err
			}
		}
	}

	if err := flushBatch(); err != nil {
		return nil, err
	}

	log.Printf("CSV Import completed. Total rows: %d, Errors: %d", report.TotalRows, len(report.Errors))

	return report, nil
}
