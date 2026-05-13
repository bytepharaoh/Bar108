package server

import (
	"bar108/config"
	"bar108/internal/handlers"
	"context"
	"database/sql"
	"os"
	"os/signal"
	"syscall"

	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type Server struct {
	httpServer *http.Server
	router     *gin.Engine
}

func New(cfg *config.Config, db *sql.DB) *Server {
	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(gin.Logger())
	// Trust only localhost proxy — fixes the Gin warning you saw.
	// In production you'd set this to your load balancer's IP.
	if err := router.SetTrustedProxies([]string{"127.0.0.1"}); err != nil {
		log.Fatalf("server: failed to set trusted proxies: %v", err)
	}
	s := &Server{
		router: router,
		httpServer: &http.Server{
			Addr:         ":" + cfg.AppPort,
			Handler:      router,
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
			IdleTimeout:  60 * time.Second,
		},
	}
	s.setupRoutes(db)
	return s
}

// setupRoutes creates all layers (repo → service → handler)

func (s *Server) setupRoutes(db *sql.DB) {
	menuRepo := newMenuRepository(db)
	userRepo := newUserRepository(db)
	menuSvc := newMenuService(menuRepo)
	userSvc := newUserService(userRepo)
	menuHandler := handlers.NewMenuHandler(menuSvc)
	userHandler := handlers.NewUserHandler(userSvc)
	s.router.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"message": "pong",
		})
	})
	menu := s.router.Group("/menu")
	{
		menu.GET("", menuHandler.GetAllMenuItems)
		menu.GET("/:id", menuHandler.GetMenuItemByID)
		menu.POST("", menuHandler.CreateMenuItem)
		menu.PUT("/:id", menuHandler.UpdateMenuItem)
		menu.DELETE("/:id", menuHandler.DeleteMenuItem)
	}
	s.router.GET("/categories", menuHandler.GetAllCategories)
	// User routes
	users := s.router.Group("/users")
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

}
func (s *Server) Run() {
	go func() {
		log.Printf("server: listening on %s", s.httpServer.Addr)
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server: failed to start: %v", err)
		}

	}()
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("server: shutting down gracefully...")
	// Give active requests 5 seconds to finish.
	// After 5 seconds, any remaining connections are forcefully closed.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := s.httpServer.Shutdown(ctx); err != nil {
		log.Printf("server: forced shutdown: %v", err)

	}
	log.Println("server: stopped")

}
