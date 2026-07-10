package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"ntd/backend/models"
	"ntd/backend/services"
)

// ProductHandler coordinates HTTP routing for the product catalog domain.
type ProductHandler struct {
	service ProductService
}

// NewProductHandler creates a new ProductHandler with the given ProductService.
func NewProductHandler(service ProductService) *ProductHandler {
	return &ProductHandler{service: service}
}

func writeJSONError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// ListAndSearchProducts handles product listing and keyword/category filters.
func (h *ProductHandler) ListAndSearchProducts(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	category := r.URL.Query().Get("category")

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	products, err := h.service.ListProducts(ctx, q, category)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Failed to query products catalog")
		return
	}

	writeJSON(w, http.StatusOK, products)
}

// CreateProduct handles inserting a new product.
func (h *ProductHandler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	var p models.Product
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid JSON payload")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	err := h.service.CreateProduct(ctx, &p)
	if err != nil {
		if errors.Is(err, services.ErrInvalidInput) || errors.Is(err, services.ErrSKUDuplicate) {
			writeJSONError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSONError(w, http.StatusInternalServerError, "Failed to create product")
		return
	}

	writeJSON(w, http.StatusCreated, p)
}

// UpdateProduct updates product attributes for a specific ID.
func (h *ProductHandler) UpdateProduct(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeJSONError(w, http.StatusBadRequest, "Product ID is required")
		return
	}

	var p models.Product
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid JSON payload")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	err := h.service.UpdateProduct(ctx, id, &p)
	if err != nil {
		if errors.Is(err, services.ErrInvalidInput) || errors.Is(err, services.ErrSKUDuplicate) {
			writeJSONError(w, http.StatusBadRequest, err.Error())
			return
		}
		if errors.Is(err, services.ErrNotFound) {
			writeJSONError(w, http.StatusNotFound, err.Error())
			return
		}
		writeJSONError(w, http.StatusInternalServerError, "Failed to update product")
		return
	}

	p.ID = id
	writeJSON(w, http.StatusOK, p)
}

// DeleteProduct removes a product by ID.
func (h *ProductHandler) DeleteProduct(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeJSONError(w, http.StatusBadRequest, "Product ID is required")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	err := h.service.DeleteProduct(ctx, id)
	if err != nil {
		if errors.Is(err, services.ErrInvalidInput) {
			writeJSONError(w, http.StatusBadRequest, err.Error())
			return
		}
		if errors.Is(err, services.ErrNotFound) {
			writeJSONError(w, http.StatusNotFound, err.Error())
			return
		}
		writeJSONError(w, http.StatusInternalServerError, "Failed to delete product")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ImportProducts parses an uploaded CSV file and imports products into the catalog.
func (h *ProductHandler) ImportProducts(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Failed to parse multipart form")
		return
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Failed to retrieve uploaded file from 'file' field")
		return
	}
	defer file.Close()

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	report, err := h.service.ImportProducts(ctx, file)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, report)
}
