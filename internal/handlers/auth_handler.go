package handlers

import (
	"bar108/internal/apperror"
	"bar108/internal/cache"
	"bar108/internal/jwt"
	"bar108/internal/middleware"
	"bar108/internal/services"
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type authSvc interface {
	Register(ctx context.Context, input services.RegisterInput) (services.AuthResult, error)
	Login(ctx context.Context, input services.LoginInput) (services.AuthResult, error)
}

type AuthHandler struct {
	service    authSvc
	cache      *cache.Client
	jwtManager *jwt.Manager
}

func NewAuthHandler(service authSvc, cache *cache.Client, jwtManager *jwt.Manager) *AuthHandler {
	return &AuthHandler{
		service:    service,
		cache:      cache,
		jwtManager: jwtManager,
	}
}

type registerRequest struct {
	Name     string `json:"name"     binding:"required"`
	Phone    string `json:"phone"    binding:"required"`
	Email    string `json:"email"    binding:"required"`
	Password string `json:"password" binding:"required"`
}

type loginRequest struct {
	Email    string `json:"email"    binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apperror.Respond(c, apperror.ErrInvalidInput)
		return
	}

	result, err := h.service.Register(c.Request.Context(), services.RegisterInput{
		Name:     req.Name,
		Phone:    req.Phone,
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		apperror.Respond(c, err)
		return
	}

	// 201 Created — new user was created
	c.JSON(http.StatusCreated, gin.H{
		"token": result.Token,
		"user": gin.H{
			"id":    result.User.ID,
			"name":  result.User.Name,
			"email": result.User.Email,
			"role":  result.User.Role,
		},
	})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apperror.Respond(c, apperror.ErrInvalidInput)
		return
	}

	result, err := h.service.Login(c.Request.Context(), services.LoginInput{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		apperror.Respond(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": result.Token,
		"user": gin.H{
			"id":    result.User.ID,
			"email": result.User.Email,
			"role":  result.User.Role,
		},
	})
}

// Logout invalidates the current JWT by storing its JTI in Redis.
func (h *AuthHandler) Logout(c *gin.Context) {
	jti, ok := middleware.GetJTI(c)
	if !ok {
		apperror.Respond(c, apperror.ErrUnauthorized)
		return
	}

	// Get expiry from the token directly
	// Re-parse from Authorization header to get claims
	authHeader := c.GetHeader("Authorization")
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 {
		apperror.Respond(c, apperror.ErrUnauthorized)
		return
	}
	claims, err := h.jwtManager.Verify(parts[1])
	if err != nil {
		apperror.Respond(c, apperror.ErrUnauthorized)
		return
	}

	ttl := time.Until(claims.ExpiresAt.Time)
	if ttl <= 0 {
		// Token already expired — nothing to blacklist
		c.JSON(http.StatusOK, gin.H{"message": "logged out successfully"})
		return
	}

	if err := h.cache.BlacklistToken(c.Request.Context(), jti, ttl); err != nil {
		apperror.Respond(c, apperror.ErrInternal)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "logged out successfully"})
}
