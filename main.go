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
    err := godotenv.Load()
    if err != nil {
        log.Println("No .env file found, using system env vars")
    }

    err = db.InitDB()
    if err != nil {
        log.Fatalf("Failed to initialize database: %v", err)
    }
    defer db.CloseDB()

    router := mux.NewRouter()
    api.SetupRoutes(router)

    port := os.Getenv("PORT")
    if port == "" {
        port = "8080" // Default port if not specified
    }
    log.Printf("Server starting on port %s", port)
    if err := http.ListenAndServe(":"+port, router); err != nil {
        log.Fatalf("Server failed to start: %v", err)
    }
}