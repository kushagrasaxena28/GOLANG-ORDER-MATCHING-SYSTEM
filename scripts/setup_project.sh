#!/bin/bash

# Exit on any error
set -e

# Load environment variables from .env
set -a
source .env 2>/dev/null || { echo "Error: .env file not found or unreadable"; exit 1; }
set +a

# Check if MySQL is installed and accessible
if ! command -v mysql &> /dev/null; then
    echo "Error: MySQL client is not installed. Please install it first."
    exit 1
fi

# Check if MySQL server is running
if ! mysqladmin ping -u kushagra -p"${MYSQL_PASSWORD}" --silent; then
    echo "Error: MySQL server is not running or credentials are incorrect. Check your .env file."
    exit 1
fi

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "Error: Go is not installed. Please install Go version 1.22 or higher."
    exit 1
fi

# Create database and tables
echo "Setting up database 'order_matching'..."
mysql -u kushagra -p"${MYSQL_PASSWORD}" < scripts/schema.sql || { echo "Error: Failed to set up database"; exit 1; }

# Install Go dependencies
echo "Installing Go dependencies..."
go mod tidy || { echo "Error: Failed to install dependencies"; exit 1; }

# Start the server on the specified port
PORT=${PORT:-8080}  # Default to 8080 if not set
echo "Starting server on port $PORT..."
PORT=$PORT go run main.go &>/dev/null &

echo "Setup complete. Server is running on port $PORT. Use Ctrl+C to stop."