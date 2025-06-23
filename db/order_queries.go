package db

import (
	"database/sql"
	"fmt"
	"log"
	"time"
	"golang-order-matching-system/models"
)


// CreateOrderTx inserts a new order within a transaction
func CreateOrderTx(order *models.Order, tx *sql.Tx) error {
	query := `
		INSERT INTO orders (symbol, side, type, price, quantity, remaining_quantity, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`
	result, err := tx.Exec(query,
		order.Symbol,
		order.Side,
		order.Type,
		order.Price,
		order.Quantity,
		order.Quantity, // Initial remaining_quantity equals quantity
		order.Status,
		order.CreatedAt,
		order.UpdatedAt)
	if err != nil {
		log.Printf("Failed to create order: %v", err)
		return err
	}
	id, err := result.LastInsertId()
	if err != nil {
		log.Printf("Failed to get last insert ID: %v", err)
		return err
	}
	order.ID = id
	return nil
}

// UpdateOrderTx updates an order within a transaction
func UpdateOrderTx(order *models.Order, tx *sql.Tx) error {
	query := `
		UPDATE orders 
		SET remaining_quantity = ?, status = ?, updated_at = ?
		WHERE id = ?`

	// Add fallback for non-transactional usage with nil check for DB
	if tx == nil {
		if DB == nil {
			log.Printf("Database connection is nil")
			return fmt.Errorf("database connection is nil")
		}
		_, err := DB.Exec(query,
			order.RemainingQuantity,
			order.Status,
			order.UpdatedAt,
			order.ID)
		if err != nil {
			log.Printf("Failed to update order (non-tx): %v", err)
			return err
		}
		return nil
	}

	// If tx is provided, use it
	_, err := tx.Exec(query,
		order.RemainingQuantity,
		order.Status,
		order.UpdatedAt,
		order.ID)
	if err != nil {
		log.Printf("Failed to update order (tx): %v", err)
		return err
	}
	return nil
}

// GetOrderByID retrieves an order by its ID
func GetOrderByID(orderID int64) (*models.Order, error) {
	order := &models.Order{}
	query := `
		SELECT id, symbol, side, type, price, quantity, remaining_quantity, status, created_at, updated_at
		FROM orders WHERE id = ?`
	var createdAtBytes, updatedAtBytes []byte
	err := DB.QueryRow(query, orderID).Scan(
		&order.ID,
		&order.Symbol,
		&order.Side,
		&order.Type,
		&order.Price,
		&order.Quantity,
		&order.RemainingQuantity,
		&order.Status,
		&createdAtBytes,
		&updatedAtBytes)
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		log.Printf("Failed to get order: %v", err)
		return nil, err
	}
	order.CreatedAt, err = parseTime(createdAtBytes)
	if err != nil {
		log.Printf("Failed to parse created_at: %v", err)
		return nil, err
	}
	order.UpdatedAt, err = parseTime(updatedAtBytes)
	if err != nil {
		log.Printf("Failed to parse updated_at: %v", err)
		return nil, err
	}
	return order, nil
}

// UpdateOrderStatus updates the status of an existing order with transition validation
func UpdateOrderStatus(orderID int64, status string, remainingQuantity int) error {
	order, err := GetOrderByID(orderID)
	if err != nil {
		return err
	}
	if order == nil {
		return sql.ErrNoRows
	}

	validStatuses := map[string]bool{"open": true, "partially_filled": true, "filled": true, "canceled": true}
	if !validStatuses[status] {
		return fmt.Errorf("invalid status: %s", status)
	}

	if remainingQuantity < 0 || remainingQuantity > order.Quantity {
		return fmt.Errorf("invalid remaining_quantity: %d", remainingQuantity)
	}

	validTransitions := map[string]map[string]bool{
		"open":           {"partially_filled": true, "filled": true, "canceled": true},
		"partially_filled": {"filled": true, "canceled": true},
		"filled":          {},
		"canceled":        {},
	}

	if !validTransitions[order.Status][status] {
		return fmt.Errorf("invalid transition from %s to %s", order.Status, status)
	}

	switch status {
	case "filled":
		if remainingQuantity != 0 {
			return fmt.Errorf("filled requires remaining_quantity 0")
		}
	case "open", "partially_filled":
		if remainingQuantity == 0 {
			return fmt.Errorf("%s requires remaining_quantity > 0", status)
		}
	}

	order.Status = status
	order.RemainingQuantity = remainingQuantity
	order.UpdatedAt = time.Now()
	return UpdateOrderTx(order, nil) // Use nil tx for non-transactional update
}

// CancelOrder updates the order status to "canceled"
func CancelOrder(orderID int64) error {
    order, err := GetOrderByID(orderID)
    if err != nil {
        return err
    }
    if order == nil {
        return sql.ErrNoRows
    }

    // Check if the order is not cancellable
    if order.Status == "filled" || order.Status == "canceled" {
        return fmt.Errorf("order %d cannot be canceled, status is %s", orderID, order.Status)
    }

    order.Status = "canceled"
    order.UpdatedAt = time.Now()
    return UpdateOrderTx(order, nil)
}

// GetOrderBook retrieves the current order book for a symbol, optionally with full list
func GetOrderBook(symbol string, full bool) ([]models.Order, error) {
	var orders []models.Order
	query := `
		SELECT id, symbol, side, type, price, quantity, remaining_quantity, status, created_at, updated_at
		FROM orders 
		WHERE symbol = ? AND status IN ('open', 'partially_filled')`
	if !full {
		query += ` ORDER BY price DESC, created_at ASC LIMIT 10`
	}
	rows, err := DB.Query(query, symbol)
	if err != nil {
		log.Printf("Failed to get order book: %v", err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var order models.Order
		var price *float64
		var createdAtBytes, updatedAtBytes []byte
		if err := rows.Scan(&order.ID, &order.Symbol, &order.Side, &order.Type, &price,
			&order.Quantity, &order.RemainingQuantity, &order.Status, &createdAtBytes, &updatedAtBytes); err != nil {
			log.Printf("Failed to scan order: %v", err)
			return nil, err
		}
		order.Price = price
		order.CreatedAt, err = parseTime(createdAtBytes)
		if err != nil {
			return nil, err
		}
		order.UpdatedAt, err = parseTime(updatedAtBytes)
		if err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}
	return orders, nil
}
