package database

import (
	"bar108/config"
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"

	"log"
	"time"
)

func Connect(cfg *config.Config) *sql.DB {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.DB.Host,
		cfg.DB.Port,
		cfg.DB.User,
		cfg.DB.Password,
		cfg.DB.Name,
		cfg.DB.SSLMode)
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("database: failed to open %v", err)
	}
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxIdleTime(5 * time.Minute)
	db.SetConnMaxLifetime(30 * time.Minute)
	if err := db.Ping(); err != nil {
		log.Fatalf("database: failed to ping: %v", err)

	}
	log.Println("database: connected successfully")
	return db

}
func Migrate(db *sql.DB, migrationsDir string) {
	if err := goose.SetDialect("postgres"); err != nil {
		log.Fatalf("database: failed to set dialect: %v", err)
	}

	if err := goose.Up(db, migrationsDir); err != nil {
		log.Fatalf("database: failed to run migrations: %v", err)
	}

	log.Println("database: migrations applied successfully")
}
