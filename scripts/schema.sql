-- Drop tables if they exist (order matters because of FK constraints)
DROP TABLE IF EXISTS trades;
DROP TABLE IF EXISTS orders;

CREATE DATABASE IF NOT EXISTS order_matching;
USE order_matching;

CREATE TABLE IF NOT EXISTS orders (
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

CREATE TABLE IF NOT EXISTS trades (
    id INT AUTO_INCREMENT PRIMARY KEY,
    symbol VARCHAR(10) NOT NULL,
    buy_order_id INT NOT NULL,
    sell_order_id INT NOT NULL,
    price DECIMAL(12,8) NOT NULL,
    quantity INT NOT NULL,
    created_at DATETIME NOT NULL,
    FOREIGN KEY (buy_order_id) REFERENCES orders(id),
    FOREIGN KEY (sell_order_id) REFERENCES orders(id)
);

CREATE USER IF NOT EXISTS 'kushagra'@'localhost' IDENTIFIED BY 'yourpassword';
GRANT ALL PRIVILEGES ON order_matching.* TO 'kushagra'@'localhost';
FLUSH PRIVILEGES;

-- Indexes
CREATE INDEX idx_orders_symbol_side_price_time ON orders(symbol, side, price, created_at);
CREATE INDEX idx_orders_status ON orders(status);
CREATE INDEX idx_trades_symbol ON trades(symbol,created_at);