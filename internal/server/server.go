package server

import (
	"bar108/config"
	"bar108/internal/handlers"
	jwtpkg "bar108/internal/jwt"
	"bar108/internal/middleware"
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

type Server struct {
	httpServer *http.Server
	router     *gin.Engine
}

func New(cfg *config.Config, db *sql.DB, jwtManager *jwtpkg.Manager) *Server {
	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(gin.Logger())

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

	// Pass jwtManager into setupRoutes
	s.setupRoutes(db, jwtManager)
	return s
}

func (s *Server) setupRoutes(db *sql.DB, jwtManager *jwtpkg.Manager) {
	menuRepo := newMenuRepository(db)
	userRepo := newUserRepository(db)
	orderRepo := newOrderRepository(db)

	menuSvc := newMenuService(menuRepo)
	userSvc := newUserService(userRepo)
	orderSvc := newOrderService(orderRepo)
	authSvc := newAuthService(userRepo, jwtManager)

	menuHandler := handlers.NewMenuHandler(menuSvc)
	userHandler := handlers.NewUserHandler(userSvc)
	orderHandler := newOrderHandler(orderSvc)
	authHandler := newAuthHandler(authSvc)

	// Health check — always public
	s.router.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "message": "pong"})
	})

	// ── Public routes (no auth) ──────────────────────────
	s.router.GET("/menu", menuHandler.GetAllMenuItems)
	s.router.GET("/menu/:id", menuHandler.GetMenuItemByID)
	s.router.GET("/categories", menuHandler.GetAllCategories)

	s.router.POST("/auth/register", authHandler.Register)
	s.router.POST("/auth/login", authHandler.Login)

	// ── Authenticated routes ─────────────────────────────
	authed := s.router.Group("")
	authed.Use(middleware.AuthMiddleware(jwtManager))
	{
		// Any logged-in user
		authed.POST("/orders", orderHandler.PlaceOrder)
		authed.GET("/orders/:id", orderHandler.GetOrderByID)
		authed.GET("/orders/:id/track", orderHandler.GetOrderStatusHistory)
		authed.GET("/orders/:id/items", orderHandler.GetOrderItems)
		authed.PATCH("/orders/:id/cancel", orderHandler.CancelOrder)
		authed.GET("/users/:id/orders", orderHandler.GetOrdersByUserID)
		authed.GET("/users/:id", userHandler.GetUserByID)
		authed.PUT("/users/:id", userHandler.UpdateUser)
		authed.PATCH("/users/:id/bonus", userHandler.UpdateUserBonusPoints)

		// ── Admin only ───────────────────────────────────
		admin := authed.Group("")
		admin.Use(middleware.AdminMiddleware())
		{
			admin.GET("/orders", orderHandler.GetAllOrders)
			admin.GET("/orders/pending", orderHandler.GetPendingOrders)
			admin.PATCH("/orders/:id/status", orderHandler.UpdateOrderStatus)
			admin.PATCH("/orders/:id/courier", orderHandler.AssignCourier)

			admin.POST("/menu", menuHandler.CreateMenuItem)
			admin.PUT("/menu/:id", menuHandler.UpdateMenuItem)
			admin.DELETE("/menu/:id", menuHandler.DeleteMenuItem)

			admin.GET("/users", userHandler.GetAllUsers)
			admin.GET("/users/active", userHandler.GetActiveUsers)
			admin.PATCH("/users/:id/activate", userHandler.ActivateUser)
			admin.PATCH("/users/:id/deactivate", userHandler.DeactivateUser)

			admin.GET("/couriers", orderHandler.GetAllCouriers)
			admin.GET("/couriers/available", orderHandler.GetAvailableCouriers)
			admin.GET("/couriers/:id", orderHandler.GetCourierByID)
			admin.PATCH("/couriers/:id/status", orderHandler.UpdateCourierStatus)
		}
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
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.httpServer.Shutdown(ctx); err != nil {
		log.Printf("server: forced shutdown: %v", err)
	}
	log.Println("server: stopped")
}
