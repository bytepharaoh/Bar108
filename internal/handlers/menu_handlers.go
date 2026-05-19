package handlers

import (
	"bar108/internal/apperror"
	"bar108/internal/cache"
	"bar108/internal/db"
	"bar108/internal/services"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type MenuHandler struct {
	service services.MenuService
	cache   *cache.Client
}
type createMenuItemsRequest struct {
	CategoryID  int32  `json:"category_id" binding:"required"`
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	Price       string `json:"price" binding:"required"`
}

type updateMenuItemsRequest struct {
	CategoryID  int32  `json:"category_id" binding:"required"`
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	Price       string `json:"price" binding:"required"`
	Available   bool   `json:"available"`
}

// menuItemResponse is what we send to the client.
// We never expose raw db structs directly — this gives us
// full control over the JSON shape.
type menuItemResponse struct {
	ID           int32   `json:"id"`
	CategoryID   int32   `json:"category_id"`
	CategoryName string  `json:"category_name,omitempty"`
	Name         string  `json:"name"`
	Description  *string `json:"description"` // pointer: null if empty, string if set
	Price        string  `json:"price"`
	Available    bool    `json:"available"`
	CreatedAt    string  `json:"created_at"`
}

func toMenuItemResponse(item db.GetAllMenuItemsRow) menuItemResponse {
	var desc *string
	if item.Description.Valid {
		desc = &item.Description.String
	}
	return menuItemResponse{
		ID:           item.ID,
		CategoryID:   item.CategoryID,
		CategoryName: item.CategoryName,
		Name:         item.Name,
		Description:  desc,
		Price:        item.Price,
		Available:    item.Available,
		CreatedAt:    item.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

func toMenuItemByIDResponse(item db.GetMenuItemByIDRow) menuItemResponse {
	var desc *string
	if item.Description.Valid {
		desc = &item.Description.String
	}
	return menuItemResponse{
		ID:           item.ID,
		CategoryID:   item.CategoryID,
		CategoryName: item.CategoryName,
		Name:         item.Name,
		Description:  desc,
		Price:        item.Price,
		Available:    item.Available,
		CreatedAt:    item.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

func toMenuItemFromCreate(item db.MenuItem) menuItemResponse {
	var desc *string
	if item.Description.Valid {
		desc = &item.Description.String
	}
	return menuItemResponse{
		ID:          item.ID,
		CategoryID:  item.CategoryID,
		Name:        item.Name,
		Description: desc,
		Price:       item.Price,
		Available:   item.Available,
		CreatedAt:   item.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

func NewMenuHandler(service menuService) *MenuHandler {
	return &MenuHandler{
		service: service,
	}
}
func (h *MenuHandler) GetAllMenuItems(c *gin.Context) {
	// Try cache first
	if h.cache != nil {
		cached, err := h.cache.GetMenu(c.Request.Context())
		if err == nil && cached != "" {
			// Cache hit — return immediately, no DB query
			c.Header("X-Cache", "HIT")
			c.Data(http.StatusOK, "application/json", []byte(cached))
			return
		}
	}

	// Cache miss — query the database
	items, err := h.service.GetAllMenuItems(c.Request.Context())
	if err != nil {
		apperror.Respond(c, err)
		return
	}

	resp := make([]menuItemResponse, len(items))
	for i, item := range items {
		resp[i] = toMenuItemResponse(item)
	}

	result := gin.H{"data": resp}

	// Store in cache for 60 seconds
	if h.cache != nil {
		if data, err := json.Marshal(result); err == nil {
			_ = h.cache.SetMenu(c.Request.Context(), string(data), 60*time.Second)
		}
	}

	c.Header("X-Cache", "MISS")
	c.JSON(http.StatusOK, result)
}
func (h *MenuHandler) GetMenuItemByID(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	item, err := h.service.GetMenuItemByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, services.ErrInvalidMenuItemID) {
			apperror.Respond(c, err)
			return
		}
		if errors.Is(err, services.ErrMenuItemNotFound) {
			apperror.Respond(c, err)
			return
		}
		apperror.Respond(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": toMenuItemByIDResponse(item)})
}

func (h *MenuHandler) GetAllCategories(c *gin.Context) {
	categories, err := h.service.GetAllCategories(c.Request.Context())
	if err != nil {
		apperror.Respond(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"data": categories,
	})

}
func (h *MenuHandler) CreateMenuItem(c *gin.Context) {
	var req createMenuItemsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apperror.Respond(c, apperror.ErrInvalidInput)
		return
	}
	args := db.CreateMenuItemParams{
		CategoryID: req.CategoryID,
		Name:       req.Name,
		Description: sql.NullString{
			String: req.Description,
			Valid:  req.Description != "",
		},
		Price: req.Price,
	}
	item, err := h.service.CreateMenuItem(c.Request.Context(), args)
	if err != nil {
		if errors.Is(err, services.ErrInvalidCategoryID) ||
			errors.Is(err, services.ErrEmptyMenuItemName) ||
			errors.Is(err, services.ErrZeroPrice) ||
			errors.Is(err, services.ErrNegativePrice) {
			apperror.Respond(c, err)
			return
		}
		apperror.Respond(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": toMenuItemFromCreate(item)})
	if h.cache != nil {
		_ = h.cache.InvalidateMenu(c.Request.Context())
	}

}
func (h *MenuHandler) UpdateMenuItem(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var req updateMenuItemsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apperror.Respond(c, apperror.ErrInvalidInput)
		return
	}
	args := db.UpdateMenuItemParams{
		ID:         int32(id),
		CategoryID: req.CategoryID,
		Name:       req.Name,
		Description: sql.NullString{
			String: req.Description,
			Valid:  req.Description != "",
		},
		Price:     req.Price,
		Available: req.Available,
	}
	item, err := h.service.UpdateMenuItem(c.Request.Context(), args)
	if err != nil {
		if errors.Is(err, services.ErrInvalidCategoryID) ||
			errors.Is(err, services.ErrEmptyMenuItemName) ||
			errors.Is(err, services.ErrZeroPrice) ||
			errors.Is(err, services.ErrNegativePrice) {
			apperror.Respond(c, err)
			return
		}
		if errors.Is(err, services.ErrMenuItemNotFound) {
			apperror.Respond(c, err)
			return
		}
		apperror.Respond(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": toMenuItemFromCreate(item)})
	if h.cache != nil {
		_ = h.cache.InvalidateMenu(c.Request.Context())
	}

}
func (h *MenuHandler) DeleteMenuItem(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	err := h.service.DeleteMenuItem(c.Request.Context(), int32(id))
	if err != nil {
		if errors.Is(err, services.ErrInvalidMenuItemID) {
			apperror.Respond(c, err)
			return
		}
		if errors.Is(err, services.ErrMenuItemNotFound) {
			apperror.Respond(c, err)
			return
		}
		apperror.Respond(c, err)
		return
	}
	c.Status(http.StatusNoContent)
	if h.cache != nil {
		_ = h.cache.InvalidateMenu(c.Request.Context())
	}

}
