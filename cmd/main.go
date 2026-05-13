package main

import (
	"bar108/config"
	"bar108/internal/database"
	"bar108/internal/handlers"
	"bar108/internal/repository"
	"bar108/internal/services"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, reading from environment directly")
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// One line — all DB logic lives in the database package
	db := database.Connect(cfg)
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("Error closing database: %v", err)
		}
	}()

	menuRepo := repository.NewMenuRepository(db)
	userRepo := repository.NewUserRepository(db)

	menuSvc := services.NewMenuService(menuRepo)
	userSvc := services.NewUserService(userRepo)

	menuHandler := handlers.NewMenuHandler(menuSvc)
	userHandler := handlers.NewUserHandler(userSvc)

	router := gin.Default()

	router.GET("/ping", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"message": "pong"})
	})

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
		log.Fatalf("Failed to start server: %v", err)
	}
}
