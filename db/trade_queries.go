package db

import (	
	"log"

	"golang-order-matching-system/models"
)

// CreateTrade inserts a new trade into the database
func CreateTrade(trade *models.Trade) error {
	query := `
		INSERT INTO trades (symbol, buy_order_id, sell_order_id, price, quantity, created_at)
		VALUES (?, ?, ?, ?, ?, ?)`
	_, err := DB.Exec(query,
		trade.Symbol,
		trade.BuyOrderID,
		trade.SellOrderID,
		trade.Price,
		trade.Quantity,
		trade.CreatedAt)
	if err != nil {
		log.Printf("Failed to create trade: %v", err)
		return err
	}
	return nil
}

// GetTrades retrieves trades, optionally filtered by symbol
func GetTrades(symbol string) ([]models.Trade, error) {
	var trades []models.Trade
	query := `
		SELECT id, symbol, buy_order_id, sell_order_id, price, quantity, created_at
		FROM trades`
	if symbol != "" {
		query += ` WHERE symbol = ?`
	}
	// Conditionally pass the symbol argument
	if symbol != "" {
		rows, err := DB.Query(query, symbol)
		if err != nil {
			log.Printf("Failed to get trades: %v", err)
			return nil, err
		}
		defer rows.Close()

		for rows.Next() {
			var trade models.Trade
			var createdAtBytes []byte
			err := rows.Scan(&trade.ID, &trade.Symbol, &trade.BuyOrderID, &trade.SellOrderID, &trade.Price, &trade.Quantity, &createdAtBytes)
			if err != nil {
				log.Printf("Failed to scan trade: %v", err)
				return nil, err
			}
			trade.CreatedAt, err = parseTime(createdAtBytes)
			if err != nil {
				log.Printf("Failed to parse created_at: %v", err)
				return nil, err
			}
			trades = append(trades, trade)
		}
	} else {
		rows, err := DB.Query(query)
		if err != nil {
			log.Printf("Failed to get trades: %v", err)
			return nil, err
		}
		defer rows.Close()

		for rows.Next() {
			var trade models.Trade
			var createdAtBytes []byte
			err := rows.Scan(&trade.ID, &trade.Symbol, &trade.BuyOrderID, &trade.SellOrderID, &trade.Price, &trade.Quantity, &createdAtBytes)
			if err != nil {
				log.Printf("Failed to scan trade: %v", err)
				return nil, err
			}
			trade.CreatedAt, err = parseTime(createdAtBytes)
			if err != nil {
				log.Printf("Failed to parse created_at: %v", err)
				return nil, err
			}
			trades = append(trades, trade)
		}
	}

	return trades, nil
}