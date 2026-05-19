package main

import (
	"bar108/config"
	"bar108/internal/database"
	jwtpkg "bar108/internal/jwt"
	"bar108/internal/server"
	"log"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("main: no .env file found, reading from environment directly")
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("main: failed to load config: %v", err)
	}

	db := database.Connect(cfg)
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("main: error closing database: %v", err)
		}
	}()

	database.Migrate(db, "./db/migrations")

	jwtManager := jwtpkg.New(cfg.JWT)

	server.New(cfg, db, jwtManager).Run()
}
