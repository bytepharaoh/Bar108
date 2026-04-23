package main

import (
	"bar108/config"
	"database/sql"
	"log"
	"net/http"
	 _ "github.com/lib/pq"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Step 1 — Load .env file into environment variables
	// This must happen BEFORE config.Load() reads them
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, reading from environment directly")
	}
	// Step 2 — Load and validate config
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)

	}
	// Step 3 — Connect to PostgreSQL
	db, err := sql.Open("postgres", cfg.DB.DSN())
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("Error closing database: %v", err)

		}
	}()
	// Step 4 — Ping the database to verify the connection is real
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	log.Println("Successfully connected to database")

	db.SetMaxOpenConns(25) // max 25 simultaneous connections
	db.SetMaxIdleConns(5)  // keep 5 connections warm even when idle

	router := gin.Default()

	// Define a simple health check route
	router.GET("/ping", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})
	log.Printf("Starting server on port %s", cfg.AppPort)
	if err := router.Run(":" + cfg.AppPort); err != nil {
		log.Fatalf("Failed to start the server: %v", err)
	}
}
