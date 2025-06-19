package models

import "time"

type Order struct {
    ID               int64     `json:"id"`
    Symbol           string    `json:"symbol"`
    Side             string    `json:"side"` // "buy" or "sell"
    Type             string    `json:"type"` // "limit" or "market"
    Price            *float64  `json:"price,omitempty"` // pointer to allow NULL for market orders
    Quantity         int       `json:"quantity"`
    RemainingQuantity int      `json:"remaining_quantity"`
    Status           string    `json:"status"` // "open", "partially_filled", "filled", "canceled"
    CreatedAt        time.Time `json:"created_at"`
    UpdatedAt        time.Time `json:"updated_at"`
}
