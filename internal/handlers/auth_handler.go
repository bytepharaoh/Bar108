package handlers

import (
	"bar108/internal/apperror"
	"bar108/internal/services"
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
)

type authSvc interface {
	Register(ctx context.Context, input services.RegisterInput) (services.AuthResult, error)
	Login(ctx context.Context, input services.LoginInput) (services.AuthResult, error)
}

type AuthHandler struct {
	service authSvc
}

func NewAuthHandler(service authSvc) *AuthHandler {
	return &AuthHandler{service: service}
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
