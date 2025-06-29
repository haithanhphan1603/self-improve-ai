package main

import (
	"log"
	"os"
	"self-improve-ai/internal/goal"
	"self-improve-ai/internal/journal"
	"self-improve-ai/internal/user"
	"self-improve-ai/pkg/db"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	database := db.InitPostgres()
	database.AutoMigrate(&user.User{}, &goal.Goal{}, &journal.JournalEntry{})

	r := gin.Default()

	user.RegisterRoutes(r, database)
	goal.RegisterRoutes(r, database)
	journal.RegisterRoutes(r, database)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	r.Run(":" + port)
}
