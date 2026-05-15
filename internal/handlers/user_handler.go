package handlers

import (
	"bar108/internal/apperror"
	"bar108/internal/db"
	"bar108/internal/middleware"
	"net/http"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	service userService
}

func NewUserHandler(service userService) *UserHandler {
	return &UserHandler{service: service}
}

// =============================================
// Request structs — what the client sends us.
// binding:"required" means Gin returns 400
// automatically if the field is missing.
// =============================================

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

// =============================================
// Handlers
// =============================================

func (h *UserHandler) GetAllUsers(c *gin.Context) {
	users, err := h.service.GetAllUsers(c.Request.Context())
	if err != nil {
		apperror.Respond(c, err)
		return
	}
	// Always return [] instead of null for empty lists.
	// null forces clients to do nil checks — [] is cleaner.
	if users == nil {
		users = []db.User{}
	}
	c.JSON(http.StatusOK, gin.H{"data": users})
}

func (h *UserHandler) GetActiveUsers(c *gin.Context) {
	users, err := h.service.GetActiveUsers(c.Request.Context())
	if err != nil {
		apperror.Respond(c, err)
		return
	}
	if users == nil {
		users = []db.User{}
	}
	c.JSON(http.StatusOK, gin.H{"data": users})
}

func (h *UserHandler) GetUserByID(c *gin.Context) {
	requestedID, ok := parseID(c)
	if !ok {
		return
	}

	currentUserID, ok := middleware.GetUserID(c)
	if !ok {
		apperror.Respond(c, apperror.ErrUnauthorized)
		return
	}

	role, _ := middleware.GetRole(c)

	// Customers can only view their own profile
	if role != "admin" && requestedID != currentUserID {
		apperror.Respond(c, apperror.ErrForbidden)
		return
	}

	user, err := h.service.GetUserByID(c.Request.Context(), requestedID)
	if err != nil {
		apperror.Respond(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": user})
}

func (h *UserHandler) CreateUser(c *gin.Context) {
	var req createUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// Gin binding failure — missing required field.
		// We respond with our standard error, not Gin's raw message.
		apperror.Respond(c, apperror.ErrInvalidInput)
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
		apperror.Respond(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": user})
}

func (h *UserHandler) UpdateUser(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	currentUserID, ok := middleware.GetUserID(c)
	if !ok {
		apperror.Respond(c, apperror.ErrUnauthorized)
		return
	}

	role, _ := middleware.GetRole(c)

	// Customers can only update their own profile
	if role != "admin" && id != currentUserID {
		apperror.Respond(c, apperror.ErrForbidden)
		return
	}

	var req updateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apperror.Respond(c, apperror.ErrInvalidInput)
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
		apperror.Respond(c, err)
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
		apperror.Respond(c, apperror.ErrInvalidInput)
		return
	}
	user, err := h.service.UpdateUserBonusPoints(c.Request.Context(), id, req.BonusPoints)
	if err != nil {
		apperror.Respond(c, err)
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
		apperror.Respond(c, err)
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
		apperror.Respond(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": user})
}
