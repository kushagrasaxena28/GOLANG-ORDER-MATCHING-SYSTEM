package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)


// InitDB initializes the database connection
func InitDB() error {
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, using system env vars")
	}

	// Get database URL from environment
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL is not set")
		return fmt.Errorf("DATABASE_URL is not set")
	}

	// Open database connection
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
		}
	}
}