package main

import (
	"log"
	"os"
	"self-improve-ai/internal/ai"
	"self-improve-ai/internal/goal"
	"self-improve-ai/internal/journal"
	"self-improve-ai/internal/user"
	"self-improve-ai/pkg/db"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using system environment variables")
	}

	// Initialize AI client
	if err := ai.InitClient(); err != nil {
		log.Fatalf("Failed to initialize AI client: %v", err)
	}

	// Initialize database
	database := db.InitPostgres()
	if database == nil {
		log.Fatal("Failed to connect to database")
	}

	// Auto-migrate database schemas
	if err := database.AutoMigrate(&user.User{}, &goal.Goal{}, &journal.JournalEntry{}); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	// Initialize Gin router
	r := gin.Default()

	// Register routes
	user.RegisterRoutes(r, database)
	goal.RegisterRoutes(r, database)
	journal.RegisterRoutes(r, database)

	// Determine port
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Start server
	log.Printf("Server starting on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
