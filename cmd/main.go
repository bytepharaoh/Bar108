package main

import (
	"bar108/config"
	"bar108/internal/database"
	"bar108/internal/server"
	"log"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	// 1 — Load .env file
	// Must happen first — before anything reads environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("main: no .env file found, reading from environment directly")
	}

	// 2 — Load and validate config
	// Fails fast if required env vars are missing
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("main: failed to load config: %v", err)
	}

	// 3 — Connect to database
	// Fails fast if DB is unreachable
	db := database.Connect(cfg)
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("main: error closing database: %v", err)
		}
	}()

	// 4 — Create and run the server
	// Blocks here until Ctrl+C or SIGTERM
	// Graceful shutdown happens inside Run()
	server.New(cfg, db).Run()
}
