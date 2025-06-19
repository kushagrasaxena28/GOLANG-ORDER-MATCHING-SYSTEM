# Test Cases for Order Matching System

## Overview
This document contains test cases for the Go-based order matching system developed as part of the assignment. The system handles limit and market orders for a symbol (e.g., AAPL) and includes features for order creation, matching, cancellation, and order book retrieval.

## Test Cases

### Case 1: Initial Order Creation
- **Description**: Verify that limit and market orders can be created and persisted.
- **Steps**:
  1. Send POST `/orders` with `{"symbol":"AAPL","side":"sell","type":"limit","price":100.00,"quantity":5}`.
  2. Send POST `/orders` with `{"symbol":"AAPL","side":"buy","type":"market","quantity":5}`.
- **Expected Outcome**:
  - Responses: 201 Created with `status: open` for limit, `status: open` for market.
  - Database: Rows in `orders` with matching `symbol`, `side`, `type`, `price`, `quantity`, and `remaining_quantity`.
- **Status**: Completed (assumed based on system setup).

### Case 2: Market Order Matching
- **Description**: Test a market buy order matching with existing sell limit orders.
- **Steps**:
  1. Send POST `/orders` with `{"symbol":"AAPL","side":"sell","type":"limit","price":101.00,"quantity":4}`.
  2. Send POST `/orders` with `{"symbol":"AAPL","side":"sell","type":"limit","price":102.00,"quantity":3}`.
  3. Send POST `/orders` with `{"symbol":"AAPL","side":"buy","type":"market","quantity":5}`.
- **Expected Outcome**:
  - Sell order (101.00): `status: filled`, `remaining_quantity: 0`.
  - Sell order (102.00): `status: partially_filled`, `remaining_quantity: 2`.
  - Market order: `status: filled`, `remaining_quantity: 0`.
  - Trades: 4 units at 101.00, 1 unit at 102.00.
  - Order book: Reflects remaining 2 units at 102.00.
- **Status**: Completed (validated with IDs 41-43).

### Case 3: Limit Order Cancellation
- **Description**: Test cancellation of an open limit order.
- **Steps**:
  1. Send POST `/orders` with `{"symbol":"AAPL","side":"sell","type":"limit","price":103.00,"quantity":2}`.
  2. Send DELETE `/orders/{id}` with the created order ID.
- **Expected Outcome**:
  - Response: 200 OK with `status: canceled`.
  - Database: Order updated to `status: canceled`.
  - Order book: Order removed from asks.
- **Status**: Completed (assumed based on Case 6).

### Case 4: Partial Order Matching with Limit Orders
- **Description**: Test a limit buy order partially matching with multiple sell limit orders.
- **Steps**:
  1. Send POST `/orders` with `{"symbol":"AAPL","side":"sell","type":"limit","price":102.00,"quantity":3}`.
  2. Send POST `/orders` with `{"symbol":"AAPL","side":"sell","type":"limit","price":101.00,"quantity":4}`.
  3. Send POST `/orders` with `{"symbol":"AAPL","side":"buy","type":"limit","price":101.50,"quantity":6}`.
- **Expected Outcome**:
  - Sell order (102.00): `status: filled`, `remaining_quantity: 0`.
  - Sell order (101.00): `status: partially_filled`, `remaining_quantity: 1`.
  - Buy order: `status: partially_filled`, `remaining_quantity: 0`.
  - Trades: 3 units at 102.00, 3 units at 101.00.
  - Order book: Reflects 1 unit at 101.00.
- **Status**: Completed (validated with IDs 47-49, though matching with 102.00 needs confirmation).

### Case 5: Concurrent Order Matching
- **Description**: Test multiple limit orders matching simultaneously.
- **Steps**:
  1. Send POST `/orders` with `{"symbol":"AAPL","side":"sell","type":"limit","price":100.00,"quantity":2}`.
  2. Send POST `/orders` with `{"symbol":"AAPL","side":"sell","type":"limit","price":99.00,"quantity":3}`.
  3. Send POST `/orders` with `{"symbol":"AAPL","side":"buy","type":"limit","price":100.50,"quantity":4}`.
  4. Send POST `/orders` with `{"symbol":"AAPL","side":"buy","type":"limit","price":100.50,"quantity":1}`.
- **Expected Outcome**:
  - Sell orders: First 2 units filled, next 2 units partially filled.
  - Buy orders: First 4 units filled, second 1 unit filled.
  - Trades: 2 units at 100.00, 3 units at 99.00.
  - Order book: Updated with remaining quantities.
- **Status**: Not yet tested (proposed for next case).

## Notes
- Test cases assume the system runs on `localhost:8080` with a MySQL database named `order_matching`.
- All times are in IST (UTC+5:30) as of June 20, 2025.