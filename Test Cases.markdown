# Test Cases for Order Matching System

## Overview
This document outlines the test cases designed to validate the functionality of the order matching system. Each test case includes a description, steps, expected outcome, and actual outcome (to be filled during testing).

## Test Cases

### Case 1: Order Creation
- **Description**: Verify that a new order can be successfully created.
- **Steps**:
  1. Send a POST request to `/orders` with payload: `{"symbol":"AAPL","side":"buy","type":"limit","price":100.00,"quantity":10}`.
  2. Check the response for a 201 status and the order ID.
- **Expected Outcome**: Order is created with status "open" and remaining_quantity = 10.
- **Actual Outcome**: [To be filled]

### Case 2: Market Order Matching
- **Description**: Test matching of a market order with an existing limit order.
- **Steps**:
  1. Create a limit order: `{"symbol":"AAPL","side":"sell","type":"limit","price":100.00,"quantity":5}`.
  2. Send a market order: `{"symbol":"AAPL","side":"buy","type":"market","quantity":5}`.
  3. Check the order book and trades.
- **Expected Outcome**: Both orders are filled, a trade is logged with price 100.00 and quantity 5.
- **Actual Outcome**: [To be filled]

### Case 3: Limit Order Cancellation
- **Description**: Verify that an open limit order can be canceled.
- **Steps**:
  1. Create a limit order: `{"symbol":"AAPL","side":"buy","type":"limit","price":100.00,"quantity":10}`.
  2. Send a DELETE request to `/orders/{id}`.
  3. Check the order status.
- **Expected Outcome**: Order status changes to "canceled" with a 204 response.
- **Actual Outcome**: [To be filled]

### Case 4: Partial Matching with Market Orders
- **Description**: Test partial matching when a market order quantity exceeds an available limit order.
- **Steps**:
  1. Create a limit order: `{"symbol":"AAPL","side":"sell","type":"limit","price":100.00,"quantity":3}`.
  2. Send a market order: `{"symbol":"AAPL","side":"buy","type":"market","quantity":5}`.
  3. Check the order book and trades.
- **Expected Outcome**: Limit order is filled, market order remains open with remaining_quantity = 2.
- **Actual Outcome**: [To be filled]

### Case 5: Partial Matching with Limit Orders
- **Description**: Test partial matching between two limit orders.
- **Steps**:
  1. Create a buy limit order: `{"symbol":"AAPL","side":"buy","type":"limit","price":101.00,"quantity":10}`.
  2. Create a sell limit order: `{"symbol":"AAPL","side":"sell","type":"limit","price":100.00,"quantity":4}`.
  3. Check the order book and trades.
- **Expected Outcome**: Partial match of 4 units at 100.00, both orders updated with remaining_quantity = 6 and 0 respectively.
- **Actual Outcome**: [To be filled]

### Case 6: Order Cancellation
- **Description**: Verify cancellation of a partially filled order.
- **Steps**:
  1. Follow Case 5 to create a partially filled order.
  2. Send a DELETE request to `/orders/{id}` for the buy order.
  3. Check the order status.
- **Expected Outcome**: Order status changes to "canceled" with a 204 response, remaining_quantity preserved.
- **Actual Outcome**: [To be filled]