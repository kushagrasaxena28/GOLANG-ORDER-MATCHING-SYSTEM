package main

import (
    "log"
    "net/http"
    "os"

    "golang-order-matching-system/db"    
    "golang-order-matching-system/api" 
    "github.com/gorilla/mux"
    "github.com/joho/godotenv"
)

func main() {
    // Load environment variables from .env file
    err := godotenv.Load()
    if err != nil {
        log.Println("No .env file found, using system env vars")
    }

    // Initialize database connection
    err = db.InitDB()
    if err != nil {
        log.Fatalf("Failed to initialize database: %v", err)
    }
    defer db.CloseDB()

    // Set up HTTP router
    router := mux.NewRouter()
    api.SetupRoutes(router)

    log.Printf("Server starting on port %s", os.Getenv("PORT"))
    if err := http.ListenAndServe(":"+os.Getenv("PORT"), router); err != nil {
        log.Fatalf("Server failed to start: %v", err)
    }
}