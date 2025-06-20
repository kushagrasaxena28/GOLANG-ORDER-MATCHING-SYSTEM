package engine

import (
	"database/sql"
	"fmt"
	"log"
	"sort"
	"time"

	"golang-order-matching-system/db"
	"golang-order-matching-system/models"
	"github.com/go-sql-driver/mysql" // Required for MySQLError
)

// OrderStatus constants for maintainability
const (
	OrderStatusOpen           = "open"
	OrderStatusPartiallyFilled = "partially_filled"
	OrderStatusFilled         = "filled"
	OrderStatusCanceled       = "canceled"
)

// OrderBook manages the in-memory order book for matching
type OrderBook struct {
	Orders map[string][]*models.Order
}

// NewOrderBook creates a new order book instance
func NewOrderBook() *OrderBook {
	return &OrderBook{
		Orders: make(map[string][]*models.Order),
	}
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// matchOrders performs the core matching logic between bids and asks within a transaction
func (ob *OrderBook) matchOrders(bids, asks []*models.Order, tx *sql.Tx) bool {
	matched := false
	processedPairs := make(map[string]bool)

	for len(bids) > 0 && len(asks) > 0 {
		bid := bids[0]
		ask := asks[0]

		if bid.Type == "market" || ask.Type == "market" {
			quantity := min(bid.RemainingQuantity, ask.RemainingQuantity)
			if quantity > 0 {
				bid.RemainingQuantity -= quantity
				ask.RemainingQuantity -= quantity
				updateOrderStatus(bid)
				updateOrderStatus(ask)

				if err := db.UpdateOrderTx(bid, tx); err != nil {
					log.Printf("Failed to update bid order %d: %v", bid.ID, err)
					return false
				}
				if err := db.UpdateOrderTx(ask, tx); err != nil {
					log.Printf("Failed to update ask order %d: %v", ask.ID, err)
					return false
				}

				if err := logTrade(bid, ask, *ask.Price, quantity, tx); err != nil {
					log.Printf("Failed to log trade for orders %d and %d: %v", bid.ID, ask.ID, err)
					return false
				}

				matched = true
			}
			if bid.RemainingQuantity == 0 {
				bids = bids[1:]
			}
			if ask.RemainingQuantity == 0 {
				asks = asks[1:]
			} else {
				asks[0] = ask
			}
			continue
		}

		if bid.Price != nil && ask.Price != nil && *bid.Price >= *ask.Price {
			pairKey := fmt.Sprintf("%d-%d", bid.ID, ask.ID)
			if !processedPairs[pairKey] {
				quantity := min(bid.RemainingQuantity, ask.RemainingQuantity)
				if quantity > 0 {
					bid.RemainingQuantity -= quantity
					ask.RemainingQuantity -= quantity
					updateOrderStatus(bid)
					updateOrderStatus(ask)

					if err := db.UpdateOrderTx(bid, tx); err != nil {
						log.Printf("Failed to update bid order %d: %v", bid.ID, err)
						return false
					}
					if err := db.UpdateOrderTx(ask, tx); err != nil {
						log.Printf("Failed to update ask order %d: %v", ask.ID, err)
						return false
					}

					if err := logTrade(bid, ask, *ask.Price, quantity, tx); err != nil {
						log.Printf("Failed to log trade for orders %d and %d: %v", bid.ID, ask.ID, err)
						return false
					}

					processedPairs[pairKey] = true
					matched = true
				}
				if ask.RemainingQuantity == 0 {
					asks = asks[1:]
				} else {
					asks[0] = ask
				}
			} else {
				asks = asks[1:]
			}
		} else {
			asks = asks[1:] // Move to next ask if price condition fails
		}
		if bid.RemainingQuantity == 0 {
			bids = bids[1:]
		}
	}
	return matched
}

// updateOrderStatus sets the status based on remaining quantity
func updateOrderStatus(order *models.Order) {
	order.UpdatedAt = time.Now()
	if order.RemainingQuantity == 0 {
		order.Status = OrderStatusFilled
	} else {
		order.Status = OrderStatusPartiallyFilled
	}
}

// logTrade records a trade in the database with duplicate handling
func logTrade(bid, ask *models.Order, price float64, quantity int, tx *sql.Tx) error {
	trade := &models.Trade{
		Symbol:      bid.Symbol,
		BuyOrderID:  bid.ID,
		SellOrderID: ask.ID,
		Price:       price,
		Quantity:    quantity,
		CreatedAt:   time.Now(),
	}
	if err := db.CreateTrade(trade); err != nil {
		if err, ok := err.(*mysql.MySQLError); ok && err.Number == 1062 { // Duplicate entry
			log.Printf("Duplicate trade ignored: BuyOrderID=%d, SellOrderID=%d, Error: %v", bid.ID, ask.ID, err)
			return nil
		}
		return err
	}
	log.Printf("Trade logged: %s, Price: %.2f, Quantity: %d", trade.Symbol, trade.Price, trade.Quantity)
	return nil
}

// MatchOrders processes a new order and attempts to match it with existing orders
func (ob *OrderBook) MatchOrders(newOrder *models.Order) error {
	tx, err := db.DB.Begin()
	if err != nil {
		log.Printf("Failed to begin transaction: %v", err)
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
			log.Printf("Transaction rolled back for order %d due to error: %v", newOrder.ID, err)
		}
	}()

	newOrder.CreatedAt = time.Now()
	newOrder.UpdatedAt = newOrder.CreatedAt
	if err := db.CreateOrderTx(newOrder, tx); err != nil {
		log.Printf("Failed to create order %d: %v", newOrder.ID, err)
		return err
	}

	existingOrders, err := db.GetOrderBook(newOrder.Symbol, true)
	if err != nil {
		log.Printf("Failed to get order book for %s: %v", newOrder.Symbol, err)
		return err
	}

	if ob.Orders[newOrder.Symbol] == nil {
		ob.Orders[newOrder.Symbol] = make([]*models.Order, 0)
	}
	ob.Orders[newOrder.Symbol] = append(ob.Orders[newOrder.Symbol], newOrder)

	var bids, asks []*models.Order
	for _, order := range existingOrders {
		ord := &models.Order{
			ID:               order.ID,
			Symbol:           order.Symbol,
			Side:             order.Side,
			Type:             order.Type,
			Price:            order.Price,
			Quantity:         order.Quantity,
			RemainingQuantity: order.RemainingQuantity,
			Status:           order.Status,
			CreatedAt:        order.CreatedAt,
			UpdatedAt:        order.UpdatedAt,
		}
		if ord.Side == "buy" {
			bids = append(bids, ord)
		} else {
			asks = append(asks, ord)
		}
	}
	if newOrder.Side == "buy" {
		bids = append(bids, newOrder)
	} else {
		asks = append(asks, newOrder)
	}

	sort.Slice(bids, func(i, j int) bool {
		if bids[i].Price != nil && bids[j].Price != nil {
			if *bids[i].Price != *bids[j].Price {
				return *bids[i].Price > *bids[j].Price
			}
		}
		return bids[i].CreatedAt.Before(bids[j].CreatedAt)
	})
	sort.Slice(asks, func(i, j int) bool {
		if asks[i].Price != nil && asks[j].Price != nil {
			if *asks[i].Price != *asks[j].Price {
				return *asks[i].Price < *asks[j].Price
			}
		}
		return asks[i].CreatedAt.Before(asks[j].CreatedAt)
	})

	matched := ob.matchOrders(bids, asks, tx)
	log.Printf("Processed order %d (type: %s), matched: %v, bids: %d, asks: %d", newOrder.ID, newOrder.Type, matched, len(bids), len(asks))

	if !matched && newOrder.Type == "market" && newOrder.RemainingQuantity > 0 {
		log.Printf("No match for market order %d, remaining quantity %d, status remains open", newOrder.ID, newOrder.RemainingQuantity)
	} else if !matched && newOrder.Type == "limit" && newOrder.RemainingQuantity > 0 {
		log.Printf("No match for limit order %d, remaining quantity %d, status remains open", newOrder.ID, newOrder.RemainingQuantity)
	} else if !matched && newOrder.RemainingQuantity > 0 && newOrder.Type != "market" && newOrder.Type != "limit" {
		newOrder.Status = OrderStatusCanceled
		newOrder.UpdatedAt = time.Now()
		if err := db.UpdateOrderTx(newOrder, tx); err != nil {
			log.Printf("Failed to cancel order %d: %v", newOrder.ID, err)
			return err
		}
		log.Printf("No match for order %d, canceled with remaining quantity %d due to invalid type", newOrder.ID, newOrder.RemainingQuantity)
	}

	if err := tx.Commit(); err != nil {
		log.Printf("Failed to commit transaction for order %d: %v", newOrder.ID, err)
		return err
	}
	log.Printf("Transaction committed successfully for order %d", newOrder.ID)
	return nil
}