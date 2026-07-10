package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"ecommerce/backend/services"
)

type PurchaseRequest struct {
	SKU            string  `json:"sku"`
	CustomerID     string  `json:"customer_id"`
	Quantity       int     `json:"quantity"`
	ExpectedPrice  float64 `json:"expected_price"`
	IdempotencyKey string  `json:"idempotency_key"`
}

type OrderHandler struct {
	service OrderService
}

func NewOrderHandler(service OrderService) *OrderHandler {
	return &OrderHandler{service: service}
}

func (h *OrderHandler) PurchaseProduct(w http.ResponseWriter, r *http.Request) {
	var req PurchaseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid JSON request body")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	order, err := h.service.PurchaseProduct(ctx, req.SKU, req.CustomerID, req.IdempotencyKey, req.Quantity, req.ExpectedPrice)
	if err != nil {
		if errors.Is(err, services.ErrInvalidInput) || errors.Is(err, services.ErrInsufficientStock) {
			writeJSONError(w, http.StatusBadRequest, err.Error())
			return
		}
		if errors.Is(err, services.ErrNotFound) {
			writeJSONError(w, http.StatusNotFound, err.Error())
			return
		}
		if errors.Is(err, services.ErrPaymentDeclined) {
			writeJSONError(w, http.StatusPaymentRequired, err.Error())
			return
		}
		if errors.Is(err, services.ErrConcurrencyConflict) || strings.Contains(err.Error(), "price has changed") {
			writeJSONError(w, http.StatusConflict, err.Error())
			return
		}
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, order)
}

func (h *OrderHandler) ListOrders(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	orders, err := h.service.ListOrders(ctx)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Failed to query orders")
		return
	}

	writeJSON(w, http.StatusOK, orders)
}
