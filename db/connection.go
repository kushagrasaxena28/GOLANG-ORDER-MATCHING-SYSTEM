package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

var DB *sql.DB

// InitDB initializes the database connection
func InitDB() error {
	// Check if DB is already initialized
	if DB != nil {
		log.Println("Database already initialized")
		return nil
	}

	// Get database URL from environment
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL is not set")
		return fmt.Errorf("DATABASE_URL is not set")
	}

	// Open database connection
	var err error
	DB, err = sql.Open("mysql", dbURL)
	if err != nil {
		log.Printf("Failed to open database: %v", err)
		return err
	}

	// Test the connection
	err = DB.Ping()
	if err != nil {
		log.Printf("Failed to ping database: %v", err)
		return err
	}

	log.Println("Database connected successfully")
	return nil
}

// CloseDB closes the database connection
func CloseDB() {
	if DB != nil {
		err := DB.Close()
		if err != nil {
			log.Printf("Failed to close database: %v", err)
		} else {
			log.Println("Database connection closed")
			DB = nil // Reset DB to nil after closing
		}
	}
}