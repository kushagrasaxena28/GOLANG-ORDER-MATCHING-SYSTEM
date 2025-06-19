package models

import "time"

// Trade represents a completed trade between a buy and sell order
type Trade struct {
	ID          int       `json:"id"`
	Symbol      string    `json:"symbol"`
	BuyOrderID  int64     `json:"buy_order_id"`
	SellOrderID int64     `json:"sell_order_id"`
	Price       float64   `json:"price"`
	Quantity    int       `json:"quantity"`
	CreatedAt   time.Time `json:"created_at"`
}