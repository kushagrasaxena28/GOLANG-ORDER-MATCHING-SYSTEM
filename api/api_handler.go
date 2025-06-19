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
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Enhanced input validation
	if order.Symbol == "" {
		http.Error(w, "Symbol is required", http.StatusBadRequest)
		return
	}
	if order.Quantity <= 0 {
		http.Error(w, "Quantity must be greater than 0", http.StatusBadRequest)
		return
	}
	if order.Type == "limit" {
		if order.Price == nil {
			http.Error(w, "Price is required for limit orders", http.StatusBadRequest)
			return
		}
		if *order.Price <= 0 {
			http.Error(w, "Price must be greater than 0 for limit orders", http.StatusBadRequest)
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
		http.Error(w, "Failed to process order", http.StatusInternalServerError)
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
		http.Error(w, "Invalid order ID", http.StatusBadRequest)
		return
	}

	order, err := db.GetOrderByID(orderID)
	if err != nil {
		http.Error(w, "Failed to retrieve order", http.StatusInternalServerError)
		return
	}
	if order == nil {
		http.Error(w, "Order not found", http.StatusNotFound)
		return
	}
	if order.Status == "filled" || order.Status == "canceled" {
		http.Error(w, "Order cannot be canceled, status is "+order.Status, http.StatusBadRequest)
		return
	}

	if err := db.CancelOrder(orderID); err != nil {
		http.Error(w, "Failed to cancel order", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent) // 204 No Content for successful deletion
}

// GetOrderBook handles GET /orderbook?symbol={symbol} to query the order book
func GetOrderBook(w http.ResponseWriter, r *http.Request) {
	symbol := r.URL.Query().Get("symbol")
	if symbol == "" {
		http.Error(w, "Symbol is required", http.StatusBadRequest)
		return
	}

	fullStr := r.URL.Query().Get("full")
	full := fullStr == "true" // Default to false if not provided or invalid

	orders, err := db.GetOrderBook(symbol, full)
	if err != nil {
		http.Error(w, "Failed to get order book", http.StatusInternalServerError)
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
		http.Error(w, "Failed to get trades", http.StatusInternalServerError)
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
		http.Error(w, "Invalid order ID", http.StatusBadRequest)
		return
	}

	var req struct {
		Status           string `json:"status"`
		RemainingQuantity int    `json:"remaining_quantity"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	order, err := db.GetOrderByID(orderID)
	if err != nil {
		http.Error(w, "Failed to retrieve order", http.StatusInternalServerError)
		return
	}
	if order == nil {
		http.Error(w, "Order not found", http.StatusNotFound)
		return
	}

	// Validate status
	validStatuses := map[string]bool{"open": true, "partially_filled": true, "filled": true, "canceled": true}
	if !validStatuses[req.Status] {
		http.Error(w, fmt.Sprintf("Invalid status: %s, must be one of open, partially_filled, filled, canceled", req.Status), http.StatusBadRequest)
		return
	}

	// Validate remaining quantity
	if req.RemainingQuantity < 0 || req.RemainingQuantity > order.Quantity {
		http.Error(w, fmt.Sprintf("Invalid remaining_quantity: %d, must be between 0 and original quantity %d", req.RemainingQuantity, order.Quantity), http.StatusBadRequest)
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
		http.Error(w, fmt.Sprintf("Invalid state transition: %s cannot transition to %s", order.Status, req.Status), http.StatusBadRequest)
		return
	}

	// Additional validation based on status
	switch req.Status {
	case "filled":
		if req.RemainingQuantity != 0 {
			http.Error(w, "filled status requires remaining_quantity to be 0", http.StatusBadRequest)
			return
		}
	case "open", "partially_filled":
		if req.RemainingQuantity == 0 {
			http.Error(w, fmt.Sprintf("%s status requires remaining_quantity greater than 0", req.Status), http.StatusBadRequest)
			return
		}
	}

	if err := db.UpdateOrderStatus(orderID, req.Status, req.RemainingQuantity); err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Order not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent) // 204 No Content for successful update
}

// GetOrder handles GET /orders/{id} to retrieve order details
func GetOrder(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orderID, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		http.Error(w, "Invalid order ID", http.StatusBadRequest)
		return
	}

	order, err := db.GetOrderByID(orderID)
	if err != nil {
		http.Error(w, "Failed to retrieve order", http.StatusInternalServerError)
		return
	}
	if order == nil {
		http.Error(w, "Order not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(order)
}