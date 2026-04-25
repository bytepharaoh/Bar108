package handlers

import (
	"bar108/internal/db"
	"bar108/internal/services"
	"errors"
	"github.com/gin-gonic/gin"
	"net/http"
)

type UserHandler struct {
	service services.UserService
}

func NewUserHandler(service services.UserService) *UserHandler {
	return &UserHandler{service: service}
}

type createUserRequest struct {
	Name         string `json:"name"          binding:"required"`
	Phone        string `json:"phone"         binding:"required"`
	Email        string `json:"email"         binding:"required"`
	PasswordHash string `json:"password_hash" binding:"required"`
}

type updateUserRequest struct {
	Name         string `json:"name"          binding:"required"`
	Phone        string `json:"phone"         binding:"required"`
	Email        string `json:"email"         binding:"required"`
	PasswordHash string `json:"password_hash" binding:"required"`
	BonusPoints  int32  `json:"bonus_points"`
	IsActive     bool   `json:"is_active"`
}

type updateBonusPointsRequest struct {
	BonusPoints int32 `json:"bonus_points" binding:"required"`
}

func (h *UserHandler) GetAllUsers(c *gin.Context) {
	users, err := h.service.GetAllUsers(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get users"})
		return
	}
	if users == nil {
		users = []db.User{}
	}

	c.JSON(http.StatusOK, gin.H{"data": users})
}

func (h *UserHandler) GetActiveUsers(c *gin.Context) {
	users, err := h.service.GetActiveUsers(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get active users"})
		return
	}
	if users == nil {
		users = []db.User{}
	}

	c.JSON(http.StatusOK, gin.H{"data": users})
}

func (h *UserHandler) GetUserByID(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	user, err := h.service.GetUserByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, services.ErrInvalidUserID) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
			return
		}
		if errors.Is(err, services.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": user})
}

func (h *UserHandler) CreateUser(c *gin.Context) {
	var req createUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	arg := db.CreateUserParams{
		Name:         req.Name,
		Phone:        req.Phone,
		Email:        req.Email,
		PasswordHash: req.PasswordHash,
	}
	user, err := h.service.CreateUser(c.Request.Context(), arg)
	if err != nil {
		if errors.Is(err, services.ErrEmptyUserName) ||
			errors.Is(err, services.ErrEmptyUserPhone) ||
			errors.Is(err, services.ErrEmptyUserEmail) ||
			errors.Is(err, services.ErrInvalidUserEmail) ||
			errors.Is(err, services.ErrEmptyPasswordHash) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": user})
}

func (h *UserHandler) UpdateUser(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var req updateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	arg := db.UpdateUserParams{
		ID:           id,
		Name:         req.Name,
		Phone:        req.Phone,
		Email:        req.Email,
		PasswordHash: req.PasswordHash,
		BonusPoints:  req.BonusPoints,
		IsActive:     req.IsActive,
	}
	user, err := h.service.UpdateUser(c.Request.Context(), arg)
	if err != nil {
		if errors.Is(err, services.ErrInvalidUserID) ||
			errors.Is(err, services.ErrEmptyUserName) ||
			errors.Is(err, services.ErrEmptyUserPhone) ||
			errors.Is(err, services.ErrEmptyUserEmail) ||
			errors.Is(err, services.ErrInvalidUserEmail) ||
			errors.Is(err, services.ErrEmptyPasswordHash) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if errors.Is(err, services.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update user"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": user})
}

func (h *UserHandler) UpdateUserBonusPoints(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var req updateBonusPointsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	user, err := h.service.UpdateUserBonusPoints(c.Request.Context(), id, req.BonusPoints)
	if err != nil {
		if errors.Is(err, services.ErrInvalidUserID) ||
			errors.Is(err, services.ErrNegativeBonusPoints) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if errors.Is(err, services.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update bonus points"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": user})
}

func (h *UserHandler) ActivateUser(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	user, err := h.service.ActivateUser(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, services.ErrInvalidUserID) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
			return
		}
		if errors.Is(err, services.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		if errors.Is(err, services.ErrUserAlreadyActive) {
			// 409 Conflict — the resource is already in the requested state
			c.JSON(http.StatusConflict, gin.H{"error": "user is already active"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to activate user"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": user})
}

func (h *UserHandler) DeactivateUser(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	user, err := h.service.DeactivateUser(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, services.ErrInvalidUserID) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
			return
		}
		if errors.Is(err, services.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		if errors.Is(err, services.ErrUserAlreadyInactive) {
			c.JSON(http.StatusConflict, gin.H{"error": "user is already inactive"})
			return
		}
		if errors.Is(err, services.ErrUserHasActiveOrders) {
			// 422 Unprocessable Entity — request is valid but can't be executed
			// due to business logic (user has active orders)
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "user has active orders"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to deactivate user"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": user})
}
