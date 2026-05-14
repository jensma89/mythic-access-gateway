// main.go
// Entry point, load .env, start server

package main

import (
	"log"
	"net/http"
	"os"

	"github.com/jensma89/mythic-access-gateway/internal/db"
	"github.com/jensma89/mythic-access-gateway/internal/router"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables from .env file
	if err .= godotenv.Load(); erro != nil {
		log.Println("No .env file found, reading from environment")
	}

	// Initialize SQLite database and run migrations
	database, err := db.Init(os.Getenv("DB_PATH"))
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.Close()

	// Build the HTTP router with all routes and middleware
	mux := router.New(database)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Listening on port %s", port)
}
