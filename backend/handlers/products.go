package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"ecommerce/backend/models"
	"ecommerce/backend/services"
)

type ProductHandler struct {
	service ProductService
}

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

func (h *ProductHandler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	var p models.Product
	err := json.NewDecoder(r.Body).Decode(&p)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	err = h.service.CreateProduct(ctx, &p)
	if err != nil {
		if errors.Is(err, services.ErrSKUDuplicate) {
			writeJSONError(w, http.StatusConflict, err.Error())
			return
		}
		if errors.Is(err, services.ErrInvalidInput) {
			writeJSONError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSONError(w, http.StatusInternalServerError, "Failed to create product")
		return
	}

	writeJSON(w, http.StatusCreated, p)
}

func (h *ProductHandler) UpdateProduct(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeJSONError(w, http.StatusBadRequest, "Missing product ID")
		return
	}

	var p models.Product
	err := json.NewDecoder(r.Body).Decode(&p)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	err = h.service.UpdateProduct(ctx, id, &p)
	if err != nil {
		if errors.Is(err, services.ErrSKUDuplicate) {
			writeJSONError(w, http.StatusConflict, err.Error())
			return
		}
		if errors.Is(err, services.ErrInvalidInput) {
			writeJSONError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSONError(w, http.StatusInternalServerError, "Failed to update product")
		return
	}

	writeJSON(w, http.StatusOK, p)
}

func (h *ProductHandler) DeleteProduct(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeJSONError(w, http.StatusBadRequest, "Missing product ID")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	err := h.service.DeleteProduct(ctx, id)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Failed to delete product")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}
