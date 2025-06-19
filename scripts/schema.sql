-- Drop tables if they exist (order matters because of FK constraints)
DROP TABLE IF EXISTS trades;
DROP TABLE IF EXISTS orders;

-- Orders Table
CREATE TABLE orders (
    id INT AUTO_INCREMENT PRIMARY KEY,
    symbol VARCHAR(10) NOT NULL,
    side ENUM('buy', 'sell') NOT NULL,
    type ENUM('limit', 'market') NOT NULL,
    price DECIMAL(20,8), -- NULL for market orders
    quantity INT NOT NULL,
    remaining_quantity INT NOT NULL,
    status ENUM('open', 'filled', 'canceled', 'partially_filled') NOT NULL DEFAULT 'open',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

-- Trades Table
CREATE TABLE trades (
    id INT AUTO_INCREMENT PRIMARY KEY,
    symbol VARCHAR(10) NOT NULL,
    buy_order_id INT NOT NULL,
    sell_order_id INT NOT NULL,
    price DECIMAL(20,8) NOT NULL,
    quantity INT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (buy_order_id) REFERENCES orders(id) ON DELETE CASCADE,
    FOREIGN KEY (sell_order_id) REFERENCES orders(id) ON DELETE CASCADE,
    UNIQUE (buy_order_id, sell_order_id)
);

-- Indexes
CREATE INDEX idx_orders_symbol_side_price_time ON orders(symbol, side, price, created_at);
CREATE INDEX idx_orders_status ON orders(status);
CREATE INDEX idx_trades_symbol ON trades(symbol,created_at);
