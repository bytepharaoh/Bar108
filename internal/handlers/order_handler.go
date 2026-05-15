package handlers

import (
	"bar108/internal/apperror"
	"bar108/internal/middleware"
	"bar108/internal/repository"
	"net/http"

	"github.com/gin-gonic/gin"
)

// OrderHandler handles all HTTP requests for orders and couriers.
// It depends on the private orderService interface —
type OrderHandler struct {
	service orderService
}

func NewOrderHandler(service orderService) *OrderHandler {
	return &OrderHandler{service: service}
}

// Request structs
// These define exactly what JSON the client must send.
// binding:"required" → Gin returns 400 automatically if missing.
type placeOrderRequest struct {
	Items []struct {
		MenuItemID int32 `json:"menu_item_id" binding:"required"`
		Quantity   int32 `json:"quantity"     binding:"required"`
	} `json:"items" binding:"required"`
	UserID          int32  `json:"user_id"          binding:"required"`
	PromoCode       string `json:"promo_code"`       // optional
	DeliveryAddress string `json:"delivery_address"` // optional
	Notes           string `json:"notes"`            // optional
}

// updateStatusRequest is the body for PATCH /orders/:id/status.
type updateStatusRequest struct {
	Status string `json:"status" binding:"required"`
}

// assignCourierRequest is the body for PATCH /orders/:id/courier.
type assignCourierRequest struct {
	CourierID int32 `json:"courier_id" binding:"required"`
}

// updateCourierStatusRequest is the body for PATCH /couriers/:id/status.
type updateCourierStatusRequest struct {
	Status string `json:"status" binding:"required"`
}

// Order Handlers
func (h *OrderHandler) PlaceOrder(c *gin.Context) {
	var req placeOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apperror.Respond(c, apperror.ErrInvalidInput)
		return
	}

	// Convert request items to repository input type.
	items := make([]repository.OrderItemInput, len(req.Items))
	for i, item := range req.Items {
		items[i] = repository.OrderItemInput{
			MenuItemID: item.MenuItemID,
			Quantity:   item.Quantity,
		}
	}

	input := repository.PlaceOrderInput{
		UserID:          req.UserID,
		Items:           items,
		PromoCode:       req.PromoCode,
		DeliveryAddress: req.DeliveryAddress,
		Notes:           req.Notes,
	}

	result, err := h.service.PlaceOrder(c.Request.Context(), input)
	if err != nil {
		apperror.Respond(c, err)
		return
	}

	// 201 Created — a new resource was created
	c.JSON(http.StatusCreated, gin.H{
		"data": gin.H{
			"order":       result.Order,
			"items":       result.Items,
			"total_price": result.TotalPrice,
			"final_price": result.FinalPrice,
		},
	})
}

// GetOrderByID handles GET /orders/:id.
// Returns the order + its items in one response.
// The customer calls this to see their full order.
func (h *OrderHandler) GetOrderByID(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	order, err := h.service.GetOrderByID(c.Request.Context(), id)
	if err != nil {
		apperror.Respond(c, err)
		return
	}

	// Also fetch the items so the customer sees what they ordered
	items, err := h.service.GetOrderItems(c.Request.Context(), id)
	if err != nil {
		apperror.Respond(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"order": order,
			"items": items,
		},
	})
}

// GetOrderStatusHistory handles GET /orders/:id/track.
// This is the order tracking endpoint — returns the full
// timeline of status changes so the customer knows exactly
// what happened: "Confirmed at 13:02, Preparing at 13:15..."
func (h *OrderHandler) GetOrderStatusHistory(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	history, err := h.service.GetOrderStatusHistory(c.Request.Context(), id)
	if err != nil {
		apperror.Respond(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": history})
}

// GetAllOrders handles GET /orders — admin only.
func (h *OrderHandler) GetAllOrders(c *gin.Context) {
	orders, err := h.service.GetAllOrders(c.Request.Context())
	if err != nil {
		apperror.Respond(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": orders})
}

// GetPendingOrders handles GET /orders/pending — admin only.
// Shows only orders that haven't been confirmed yet.
// This is the main admin dashboard view.
func (h *OrderHandler) GetPendingOrders(c *gin.Context) {
	orders, err := h.service.GetPendingOrders(c.Request.Context())
	if err != nil {
		apperror.Respond(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": orders})
}

// GetOrdersByUserID handles GET /users/:id/orders.
// Returns all orders for a specific user — their order history.
// func (h *OrderHandler) GetOrdersByUserID(c *gin.Context) {
// 	userID, ok := parseID(c)
// 	if !ok {
// 		return
// 	}

// 	orders, err := h.service.GetOrdersByUserID(c.Request.Context(), userID)
// 	if err != nil {
// 		apperror.Respond(c, err)
// 		return
// 	}

//		c.JSON(http.StatusOK, gin.H{"data": orders})
//	}
func (h *OrderHandler) GetOrdersByUserID(c *gin.Context) {
	requestedUserID, ok := parseID(c)
	if !ok {
		return
	}

	currentUserID, ok := middleware.GetUserID(c)
	if !ok {
		apperror.Respond(c, apperror.ErrUnauthorized)
		return
	}

	role, ok := middleware.GetRole(c)
	if !ok {
		apperror.Respond(c, apperror.ErrUnauthorized)
		return
	}

	if role != "admin" && requestedUserID != currentUserID {
		apperror.Respond(c, apperror.ErrForbidden)
		return
	}

	orders, err := h.service.GetOrdersByUserID(c.Request.Context(), requestedUserID)
	if err != nil {
		apperror.Respond(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": orders})
}

// GetOrderItems handles GET /orders/:id/items.
func (h *OrderHandler) GetOrderItems(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	items, err := h.service.GetOrderItems(c.Request.Context(), id)
	if err != nil {
		apperror.Respond(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": items})
}

// UpdateOrderStatus handles PATCH /orders/:id/status — admin only.
// The service validates that the transition is legal.
func (h *OrderHandler) UpdateOrderStatus(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	var req updateStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apperror.Respond(c, apperror.ErrInvalidInput)
		return
	}

	order, err := h.service.UpdateOrderStatus(c.Request.Context(), id, req.Status)
	if err != nil {
		apperror.Respond(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": order})
}

// CancelOrder handles PATCH /orders/:id/cancel.
// Service enforces that delivered/already-cancelled orders
// cannot be cancelled.
func (h *OrderHandler) CancelOrder(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	order, err := h.service.CancelOrder(c.Request.Context(), id)
	if err != nil {
		apperror.Respond(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": order})
}

// AssignCourier handles PATCH /orders/:id/courier — admin only.
// Service enforces the order must be in "ready" status.
func (h *OrderHandler) AssignCourier(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	var req assignCourierRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apperror.Respond(c, apperror.ErrInvalidInput)
		return
	}

	order, err := h.service.AssignCourier(c.Request.Context(), id, req.CourierID)
	if err != nil {
		apperror.Respond(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": order})
}

// =============================================
// Courier Handlers
// =============================================

func (h *OrderHandler) GetAllCouriers(c *gin.Context) {
	couriers, err := h.service.GetAllCouriers(c.Request.Context())
	if err != nil {
		apperror.Respond(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": couriers})
}

func (h *OrderHandler) GetAvailableCouriers(c *gin.Context) {
	couriers, err := h.service.GetAvailableCouriers(c.Request.Context())
	if err != nil {
		apperror.Respond(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": couriers})
}

func (h *OrderHandler) GetCourierByID(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	courier, err := h.service.GetCourierByID(c.Request.Context(), id)
	if err != nil {
		apperror.Respond(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": courier})
}

// UpdateCourierStatus handles PATCH /couriers/:id/status.
// Sets a courier as available, busy, or offline.
func (h *OrderHandler) UpdateCourierStatus(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	var req updateCourierStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apperror.Respond(c, apperror.ErrInvalidInput)
		return
	}

	courier, err := h.service.UpdateCourierStatus(c.Request.Context(), id, req.Status)
	if err != nil {
		apperror.Respond(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": courier})
}
