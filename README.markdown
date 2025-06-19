# Order Matching System

## Overview
This is a Go-based order matching system developed as part of an assignment. The system supports creating, matching, and canceling limit and market orders for a given symbol (e.g., AAPL), with a RESTful API and a MySQL backend. This README provides setup instructions, details of additional features, assumptions, and a summary of the work.

## Setup Instructions

### Prerequisites
- **Go**: Version 1.18 or higher.
- **MySQL**: Server with `mysql-client` installed.
- **Git** (optional for local cloning if not using provided files).

### Installation
1. **Clone or Copy Files**:
   - Place the project files (e.g., `main.go`, `matcher.go`, `models.go`, `db.go`) in a directory (e.g., `golang-order-matching-system`).

2. **Configure Database**:
   - Create a database named `order_matching`:
     ```bash
     mysql -u root -p -e "CREATE DATABASE order_matching;"
     ```
   - Create a user `kushagra` with a password (e.g., `yourpassword`):
     ```bash
     mysql -u root -p -e "CREATE USER 'kushagra'@'localhost' IDENTIFIED BY 'yourpassword'; GRANT ALL PRIVILEGES ON order_matching.* TO 'kushagra'@'localhost'; FLUSH PRIVILEGES;"
     ```
   - Initialize the schema (create `orders` and `trades` tables):
     ```sql
     CREATE TABLE orders (
         id INT AUTO_INCREMENT PRIMARY KEY,
         symbol VARCHAR(10) NOT NULL,
         side VARCHAR(10) NOT NULL,
         type VARCHAR(10) NOT NULL,
         price DECIMAL(12,8),
         quantity INT NOT NULL,
         remaining_quantity INT NOT NULL,
         status VARCHAR(20) NOT NULL,
         created_at DATETIME NOT NULL,
         updated_at DATETIME NOT NULL
     );

     CREATE TABLE trades (
         id INT AUTO_INCREMENT PRIMARY KEY,
         symbol VARCHAR(10) NOT NULL,
         buy_order_id INT NOT NULL,
         sell_order_id INT NOT NULL,
         price DECIMAL(12,8) NOT NULL,
         quantity INT NOT NULL,
         created_at DATETIME NOT NULL
     );
     ```
   - Update `db.go` with the correct password if different.

3. **Run the Application**:
   - Navigate to the project directory:
     ```bash
     cd golang-order-matching-system
     ```
   - Start the server:
     ```bash
     go run main.go
     ```
   - Expected output: `Server starting on port 8080`.

4. **Test the API**:
   - Use a tool like Postman or `curl`:
     - Create an order: `curl -X POST -H "Content-Type: application/json" -d '{"symbol":"AAPL","side":"sell","type":"limit","price":100.00,"quantity":5}' http://localhost:8080/orders`
     - Get order book: `curl http://localhost:8080/orderbook?symbol=AAPL&full=true`
     - Cancel order: `curl -X DELETE http://localhost:8080/orders/1`

## Additional Features Beyond the Assignment
- **Order Book Endpoint**: Added `GET /orderbook?symbol=AAPL&full=true` to retrieve the current state of bids and asks, enhancing visibility beyond the basic assignment requirements.
- **Transaction Safety**: Implemented database transactions (`tx.Commit()`/`tx.Rollback()`) to ensure atomicity in order creation and matching, not explicitly required but added for robustness.
- **Debug Logging**: Included detailed logs (e.g., `Processing order X, matched: true`) to aid troubleshooting, exceeding the minimal logging expectation.
- **Partial Matching Logic**: Enhanced `matchOrders` to handle iterative matching across multiple asks, allowing partial fills across different price levels.

## Assumptions Made
- **Time Zone**: All timestamps are assumed to be in IST (UTC+5:30) as per the project context.
- **Price Precision**: Prices are stored with 8 decimal places (`DECIMAL(12,8)`) to handle fractional values, assuming high precision is needed.
- **Order Types**: Only "limit" and "market" order types are supported, assuming these are the primary types required by the assignment.
- **Single Symbol**: The system focuses on a single symbol (AAPL) per request, assuming a simplified trading environment.
- **No External Data**: No real-time market data is integrated, assuming a controlled test environment.
- **Error Handling**: Basic error handling (e.g., transaction failures) is implemented, assuming the recruiter will test edge cases separately.

## Summary of Work Done
- **Core Functionality**:
  - Implemented REST API endpoints: `POST /orders`, `DELETE /orders/{id}`, `GET /orderbook`.
  - Developed order matching logic in `matcher.go` to pair buy and sell orders based on price and quantity.
  - Integrated MySQL database with `orders` and `trades` tables for persistence.
- **Test Cases Completed**:
  - Case 1: Initial order creation (assumed).
  - Case 2: Market order matching (IDs 41-43).
  - Case 3: Limit order cancellation .
  - Case 4: Partial order matching with market orders (IDs 41-43).
  - Case 5: Partial order matching with limit orders (IDs 47-49).
  - Case 6: Order cancellation.
- **Challenges Addressed**:
  - Fixed unintended cancellation of limit orders when no matches were found (Case 2).
  - Resolved incomplete matching across multiple asks by updating `matchOrders` (Case 5).
  - Ensured transaction integrity during matching and updates.

## Test Cases Reference
Refer to the `test_cases.md` file for a detailed list of test cases, including steps and expected outcomes, to validate the system.
