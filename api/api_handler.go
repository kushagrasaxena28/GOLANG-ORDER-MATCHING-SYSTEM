package api

import (
	"fmt"
	"encoding/json"
	"net/http"
	"strconv"
	"database/sql"
	"golang-order-matching-system/db"
	"golang-order-matching-system/engine"
	"golang-order-matching-system/models"
	"golang-order-matching-system/utils"
	"github.com/gorilla/mux"
	"time"
)

var orderBook *engine.OrderBook

// SetupRoutes sets up the API routes
func SetupRoutes(r *mux.Router) {
	r.HandleFunc("/orders", CreateOrder).Methods("POST")
	r.HandleFunc("/orders/{id}", CancelOrder).Methods("DELETE")
	r.HandleFunc("/orderbook", GetOrderBook).Methods("GET")
	r.HandleFunc("/trades", GetTrades).Methods("GET")
	r.HandleFunc("/orders/{id}/status", UpdateOrderStatus).Methods("PUT")
	r.HandleFunc("/orders/{id}", GetOrder).Methods("GET")
}

// CreateOrder handles POST /orders to place a new order
func CreateOrder(w http.ResponseWriter, r *http.Request) {
	var order models.Order
	if err := json.NewDecoder(r.Body).Decode(&order); err != nil {
		utils.JSONErrorResponse(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Enhanced input validation
	if order.Symbol == "" {
		utils.JSONErrorResponse(w, http.StatusBadRequest, "Symbol is required")
		return
	}
	if order.Quantity <= 0 {
		utils.JSONErrorResponse(w, http.StatusBadRequest, "Quantity must be greater than 0")
		return
	}
	if order.Type == "limit" {
		if order.Price == nil {
			utils.JSONErrorResponse(w, http.StatusBadRequest, "Price is required for limit orders")
			return
		}
		if *order.Price <= 0 {
			utils.JSONErrorResponse(w, http.StatusBadRequest, "Price must be greater than 0 for limit orders")
			return
		}
	}

	order.Status = "open"
	order.RemainingQuantity = order.Quantity
	order.CreatedAt = time.Now()
	order.UpdatedAt = order.CreatedAt

	if orderBook == nil {
		orderBook = engine.NewOrderBook()
	}

	if err := orderBook.MatchOrders(&order); err != nil {
		utils.JSONErrorResponse(w, http.StatusInternalServerError, "Failed to process order")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(order)
}

// CancelOrder handles DELETE /orders/{id} to cancel an order
func CancelOrder(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orderID, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		utils.JSONErrorResponse(w, http.StatusBadRequest, "Invalid order ID")
		return
	}

	order, err := db.GetOrderByID(orderID)
	if err != nil {
		utils.JSONErrorResponse(w, http.StatusInternalServerError, "Failed to retrieve order")
		return
	}
	if order == nil {
		utils.JSONErrorResponse(w, http.StatusNotFound, "Order not found")
		return
	}
	if order.Status == "filled" || order.Status == "canceled" {
		utils.JSONErrorResponse(w, http.StatusBadRequest, "Order cannot be canceled, status is "+order.Status)
		return
	}

	if err := db.CancelOrder(orderID); err != nil {
		utils.JSONErrorResponse(w, http.StatusInternalServerError, "Failed to cancel order")
		return
	}

	w.WriteHeader(http.StatusNoContent) // 204 No Content for successful deletion
}

// GetOrderBook handles GET /orderbook?symbol={symbol} to query the order book
func GetOrderBook(w http.ResponseWriter, r *http.Request) {
	symbol := r.URL.Query().Get("symbol")
	if symbol == "" {
		utils.JSONErrorResponse(w, http.StatusBadRequest, "Symbol is required")
		return
	}

	fullStr := r.URL.Query().Get("full")
	full := fullStr == "true" // Default to false if not provided or invalid

	orders, err := db.GetOrderBook(symbol, full)
	if err != nil {
		utils.JSONErrorResponse(w, http.StatusInternalServerError, "Failed to get order book")
		return
	}

	var bids, asks []models.Order
	for _, order := range orders {
		if order.Side == "buy" {
			bids = append(bids, order)
		} else {
			asks = append(asks, order)
		}
	}

	orderBookResp := struct {
		Symbol string          `json:"symbol"`
		Bids   []models.Order  `json:"bids"`
		Asks   []models.Order  `json:"asks"`
		Full   bool            `json:"full"`
	}{
		Symbol: symbol,
		Bids:   bids,
		Asks:   asks,
		Full:   full,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(orderBookResp)
}

// GetTrades handles GET /trades to retrieve trade history
func GetTrades(w http.ResponseWriter, r *http.Request) {
	symbol := r.URL.Query().Get("symbol") // Optional filter by symbol

	trades, err := db.GetTrades(symbol)
	if err != nil {
		utils.JSONErrorResponse(w, http.StatusInternalServerError, "Failed to get trades")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(trades)
}

// UpdateOrderStatus handles PUT /orders/{id}/status to update order status
func UpdateOrderStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orderID, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		utils.JSONErrorResponse(w, http.StatusBadRequest, "Invalid order ID")
		return
	}

	var req struct {
		Status           string `json:"status"`
		RemainingQuantity int    `json:"remaining_quantity"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.JSONErrorResponse(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	order, err := db.GetOrderByID(orderID)
	if err != nil {
		utils.JSONErrorResponse(w, http.StatusInternalServerError, "Failed to retrieve order")
		return
	}
	if order == nil {
		utils.JSONErrorResponse(w, http.StatusNotFound, "Order not found")
		return
	}

	// Validate status
	validStatuses := map[string]bool{"open": true, "partially_filled": true, "filled": true, "canceled": true}
	if !validStatuses[req.Status] {
		utils.JSONErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("Invalid status: %s, must be one of open, partially_filled, filled, canceled", req.Status))
		return
	}

	// Validate remaining quantity
	if req.RemainingQuantity < 0 || req.RemainingQuantity > order.Quantity {
		utils.JSONErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("Invalid remaining_quantity: %d, must be between 0 and original quantity %d", req.RemainingQuantity, order.Quantity))
		return
	}

	// Define valid state transitions
	validTransitions := map[string]map[string]bool{
		"open":           {"partially_filled": true, "filled": true, "canceled": true},
		"partially_filled": {"filled": true, "canceled": true},
		"filled":          {},
		"canceled":        {},
	}

	// Check if the transition is valid
	if !validTransitions[order.Status][req.Status] {
		utils.JSONErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("Invalid state transition: %s cannot transition to %s", order.Status, req.Status))
		return
	}

	// Additional validation based on status
	switch req.Status {
	case "filled":
		if req.RemainingQuantity != 0 {
			utils.JSONErrorResponse(w, http.StatusBadRequest, "filled status requires remaining_quantity to be 0")
			return
		}
	case "open", "partially_filled":
		if req.RemainingQuantity == 0 {
			utils.JSONErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("%s status requires remaining_quantity greater than 0", req.Status))
			return
		}
	}

	if err := db.UpdateOrderStatus(orderID, req.Status, req.RemainingQuantity); err != nil {
		if err == sql.ErrNoRows {
			utils.JSONErrorResponse(w, http.StatusNotFound, "Order not found")
			return
		}
		utils.JSONErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent) // 204 No Content for successful update
}

// GetOrder handles GET /orders/{id} to retrieve order details
func GetOrder(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orderID, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		utils.JSONErrorResponse(w, http.StatusBadRequest, "Invalid order ID")
		return
	}

	order, err := db.GetOrderByID(orderID)
	if err != nil {
		utils.JSONErrorResponse(w, http.StatusInternalServerError, "Failed to retrieve order")
		return
	}
	if order == nil {
		utils.JSONErrorResponse(w, http.StatusNotFound, "Order not found")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(order)
}