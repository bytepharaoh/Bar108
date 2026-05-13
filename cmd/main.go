package main

import (
	"bar108/config"
	"bar108/internal/handlers"
	"bar108/internal/repository"
	"bar108/internal/services"
	"database/sql"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
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

	menuRepo := repository.NewMenuRepository(db)
	userRepo := repository.NewUserRepository(db)

	menSvc := services.NewMenuService(menuRepo)
	userSvc := services.NewUserService(userRepo)

	menuHandler := handlers.NewMenuHandler(menSvc)
	userHandler := handlers.NewUserHandler(userSvc)

	router := gin.Default()

	// !Define a simple health check route
	router.GET("/ping", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})
	// !Define the main endpoints to program
	menu := router.Group("/menu")
	{

		menu.GET("", menuHandler.GetAllMenuItems)
		menu.GET("/:id", menuHandler.GetMenuItemByID)
		menu.POST("", menuHandler.CreateMenuItem)
		menu.PUT("/:id", menuHandler.UpdateMenuItem)
		menu.DELETE("/:id", menuHandler.DeleteMenuItem)
	}
	router.GET("/categories", menuHandler.GetAllCategories)

	users := router.Group("/users")
	{
		users.GET("", userHandler.GetAllUsers)
		users.GET("/active", userHandler.GetActiveUsers)
		users.GET("/:id", userHandler.GetUserByID)
		users.POST("", userHandler.CreateUser)
		users.PUT("/:id", userHandler.UpdateUser)
		users.PATCH("/:id/bonus", userHandler.UpdateUserBonusPoints)
		users.PATCH("/:id/activate", userHandler.ActivateUser)
		users.PATCH("/:id/deactivate", userHandler.DeactivateUser)

	}

	log.Printf("Starting server on port %s", cfg.AppPort)
	if err := router.Run(":" + cfg.AppPort); err != nil {
		log.Fatalf("Failed to start the server: %v", err)
	}
}
