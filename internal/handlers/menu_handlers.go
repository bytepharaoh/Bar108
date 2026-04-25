package handlers

import (
	"bar108/internal/db"
	"bar108/internal/services"
	"database/sql"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type MenuHandler struct {
	service services.MenuService
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

func NewMenuHandler(service services.MenuService) *MenuHandler {
	return &MenuHandler{
		service: service,
	}
}
func (h *MenuHandler) GetAllMenuItems(c *gin.Context) {
	items, err := h.service.GetAllMenuItems(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to get menu items",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"data": items,
	})
}
func (h *MenuHandler) GetMenuItemByID(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	item, err := h.service.GetMenuItemByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, services.ErrInvalidMenuItemID) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid menu item id"})
			return
		}
		if errors.Is(err, services.ErrMenuItemNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "menu item not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get menu item"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": item})
}

func (h *MenuHandler) GetAllCategories(c *gin.Context) {
	categories, err := h.service.GetAllCategories(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to get categories",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"data": categories,
	})

}
func (h *MenuHandler) CreateMenuItem(c *gin.Context) {
	var req createMenuItemsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body",
		})
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
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to create menu item",
		})
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"data": item,
	})
}
func (h *MenuHandler) UpdateMenuItem(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var req updateMenuItemsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body",
		})
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
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
		if errors.Is(err, services.ErrMenuItemNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "menu item not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to update menu item",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"data": item,
	})
}
func (h *MenuHandler) DeleteMenuItem(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	err := h.service.DeleteMenuItem(c.Request.Context(), int32(id))
	if err != nil {
		if errors.Is(err, services.ErrInvalidMenuItemID) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid menu item id",
			})
			return
		}
		if errors.Is(err, services.ErrMenuItemNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "menu item not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to delete menu item",
		})
		return
	}
	c.Status(http.StatusNoContent)
}
