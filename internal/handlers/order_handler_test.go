package handlers_test

import (
	"bar108/internal/apperror"
	"bar108/internal/db"
	"bar108/internal/handlers"
	"bar108/internal/repository"
	"bar108/internal/services/mocks"
	"errors"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"go.uber.org/mock/gomock"
)

// setupOrderHandler creates a fresh mock + handler + router for each test.
func setupOrderHandler(t *testing.T) (*gin.Engine, *mocks.MockOrderService) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	mockSvc := mocks.NewMockOrderService(ctrl)
	h := handlers.NewOrderHandler(mockSvc)

	r := gin.New()

	// Order routes — mirrors exactly what server.go registers
	orders := r.Group("/orders")
	{
		orders.POST("", h.PlaceOrder)
		orders.GET("", h.GetAllOrders)
		orders.GET("/pending", h.GetPendingOrders)
		orders.GET("/:id", h.GetOrderByID)
		orders.GET("/:id/track", h.GetOrderStatusHistory)
		orders.GET("/:id/items", h.GetOrderItems)
		orders.PATCH("/:id/status", h.UpdateOrderStatus)
		orders.PATCH("/:id/cancel", h.CancelOrder)
		orders.PATCH("/:id/courier", h.AssignCourier)
	}

	r.GET("/users/:id/orders", h.GetOrdersByUserID)

	couriers := r.Group("/couriers")
	{
		couriers.GET("", h.GetAllCouriers)
		couriers.GET("/available", h.GetAvailableCouriers)
		couriers.GET("/:id", h.GetCourierByID)
		couriers.PATCH("/:id/status", h.UpdateCourierStatus)
	}

	return r, mockSvc
}

// =============================================
// POST /orders
// =============================================

func TestOrderHandler_PlaceOrder(t *testing.T) {
	tests := []struct {
		name       string
		body       interface{}
		setupMock  func(m *mocks.MockOrderService)
		wantStatus int
	}{
		{
			name: "happy path — 201 created",
			body: map[string]interface{}{
				"user_id": 1,
				"items": []map[string]interface{}{
					{"menu_item_id": 1, "quantity": 2},
				},
			},
			setupMock: func(m *mocks.MockOrderService) {
				m.EXPECT().
					PlaceOrder(gomock.Any(), gomock.Any()).
					Return(repository.PlaceOrderResult{
						Order:      db.Order{ID: 1, Status: "pending"},
						TotalPrice: "350.00",
						FinalPrice: "350.00",
					}, nil)
			},
			wantStatus: http.StatusCreated,
		},
		{
			// Missing items — Gin binding returns 400
			// Service never called
			name:       "missing items — 400",
			body:       map[string]interface{}{"user_id": 1},
			setupMock:  func(m *mocks.MockOrderService) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "empty body — 400",
			body:       nil,
			setupMock:  func(m *mocks.MockOrderService) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			// Item unavailable — service/repo returns apperror
			name: "item unavailable — 400",
			body: map[string]interface{}{
				"user_id": 1,
				"items":   []map[string]interface{}{{"menu_item_id": 1, "quantity": 1}},
			},
			setupMock: func(m *mocks.MockOrderService) {
				m.EXPECT().
					PlaceOrder(gomock.Any(), gomock.Any()).
					Return(repository.PlaceOrderResult{}, apperror.ErrItemUnavailable)
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			// Promo not found
			name: "promo code not found — 400",
			body: map[string]interface{}{
				"user_id":    1,
				"items":      []map[string]interface{}{{"menu_item_id": 1, "quantity": 1}},
				"promo_code": "INVALID",
			},
			setupMock: func(m *mocks.MockOrderService) {
				m.EXPECT().
					PlaceOrder(gomock.Any(), gomock.Any()).
					Return(repository.PlaceOrderResult{}, apperror.ErrPromoNotFound)
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "database error — 500",
			body: map[string]interface{}{
				"user_id": 1,
				"items":   []map[string]interface{}{{"menu_item_id": 1, "quantity": 1}},
			},
			setupMock: func(m *mocks.MockOrderService) {
				m.EXPECT().
					PlaceOrder(gomock.Any(), gomock.Any()).
					Return(repository.PlaceOrderResult{}, errors.New("tx failed"))
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, mockSvc := setupOrderHandler(t)
			tt.setupMock(mockSvc)

			w := doRequest(r, http.MethodPost, "/orders", tt.body)

			if w.Code != tt.wantStatus {
				t.Errorf("got status %d, want %d — body: %s",
					w.Code, tt.wantStatus, w.Body.String())
			}
		})
	}
}

// =============================================
// GET /orders/:id
// =============================================

func TestOrderHandler_GetOrderByID(t *testing.T) {
	tests := []struct {
		name       string
		url        string
		setupMock  func(m *mocks.MockOrderService)
		wantStatus int
	}{
		{
			name: "happy path — 200",
			url:  "/orders/1",
			setupMock: func(m *mocks.MockOrderService) {
				m.EXPECT().
					GetOrderByID(gomock.Any(), int32(1)).
					Return(db.GetOrderByIDRow{ID: 1, Status: "pending"}, nil)
				m.EXPECT().
					GetOrderItems(gomock.Any(), int32(1)).
					Return([]db.GetOrderItemsRow{}, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "not found — 404",
			url:  "/orders/999",
			setupMock: func(m *mocks.MockOrderService) {
				m.EXPECT().
					GetOrderByID(gomock.Any(), int32(999)).
					Return(db.GetOrderByIDRow{}, apperror.ErrOrderNotFound)
			},
			wantStatus: http.StatusNotFound,
		},
		{
			// String ID — parseID rejects before service is called
			name:       "invalid id string — 400",
			url:        "/orders/abc",
			setupMock:  func(m *mocks.MockOrderService) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "database error — 500",
			url:  "/orders/1",
			setupMock: func(m *mocks.MockOrderService) {
				m.EXPECT().
					GetOrderByID(gomock.Any(), int32(1)).
					Return(db.GetOrderByIDRow{}, errors.New("db error"))
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, mockSvc := setupOrderHandler(t)
			tt.setupMock(mockSvc)

			w := doRequest(r, http.MethodGet, tt.url, nil)

			if w.Code != tt.wantStatus {
				t.Errorf("got status %d, want %d — body: %s",
					w.Code, tt.wantStatus, w.Body.String())
			}
		})
	}
}

// =============================================
// GET /orders/:id/track
// =============================================

func TestOrderHandler_GetOrderStatusHistory(t *testing.T) {
	tests := []struct {
		name       string
		url        string
		setupMock  func(m *mocks.MockOrderService)
		wantStatus int
	}{
		{
			name: "happy path — returns full timeline",
			url:  "/orders/1/track",
			setupMock: func(m *mocks.MockOrderService) {
				m.EXPECT().
					GetOrderStatusHistory(gomock.Any(), int32(1)).
					Return([]db.OrderStatusHistory{
						{OrderID: 1, Status: "pending"},
						{OrderID: 1, Status: "confirmed"},
					}, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "not found — 404",
			url:  "/orders/999/track",
			setupMock: func(m *mocks.MockOrderService) {
				m.EXPECT().
					GetOrderStatusHistory(gomock.Any(), int32(999)).
					Return(nil, apperror.ErrOrderNotFound)
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "invalid id — 400",
			url:        "/orders/abc/track",
			setupMock:  func(m *mocks.MockOrderService) {},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, mockSvc := setupOrderHandler(t)
			tt.setupMock(mockSvc)

			w := doRequest(r, http.MethodGet, tt.url, nil)

			if w.Code != tt.wantStatus {
				t.Errorf("got status %d, want %d — body: %s",
					w.Code, tt.wantStatus, w.Body.String())
			}
		})
	}
}

// =============================================
// PATCH /orders/:id/status
// =============================================

func TestOrderHandler_UpdateOrderStatus(t *testing.T) {
	tests := []struct {
		name       string
		url        string
		body       interface{}
		setupMock  func(m *mocks.MockOrderService)
		wantStatus int
	}{
		{
			name: "happy path — 200",
			url:  "/orders/1/status",
			body: map[string]interface{}{"status": "confirmed"},
			setupMock: func(m *mocks.MockOrderService) {
				m.EXPECT().
					UpdateOrderStatus(gomock.Any(), int32(1), "confirmed").
					Return(db.Order{ID: 1, Status: "confirmed"}, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			// Invalid transition — service returns 400
			name: "invalid transition — 400",
			url:  "/orders/1/status",
			body: map[string]interface{}{"status": "pending"},
			setupMock: func(m *mocks.MockOrderService) {
				m.EXPECT().
					UpdateOrderStatus(gomock.Any(), int32(1), "pending").
					Return(db.Order{}, apperror.New(400, "cannot transition"))
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "missing body — 400",
			url:        "/orders/1/status",
			body:       nil,
			setupMock:  func(m *mocks.MockOrderService) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid id — 400",
			url:        "/orders/abc/status",
			body:       map[string]interface{}{"status": "confirmed"},
			setupMock:  func(m *mocks.MockOrderService) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "order not found — 404",
			url:  "/orders/999/status",
			body: map[string]interface{}{"status": "confirmed"},
			setupMock: func(m *mocks.MockOrderService) {
				m.EXPECT().
					UpdateOrderStatus(gomock.Any(), int32(999), "confirmed").
					Return(db.Order{}, apperror.ErrOrderNotFound)
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, mockSvc := setupOrderHandler(t)
			tt.setupMock(mockSvc)

			w := doRequest(r, http.MethodPatch, tt.url, tt.body)

			if w.Code != tt.wantStatus {
				t.Errorf("got status %d, want %d — body: %s",
					w.Code, tt.wantStatus, w.Body.String())
			}
		})
	}
}

// =============================================
// PATCH /orders/:id/cancel
// =============================================

func TestOrderHandler_CancelOrder(t *testing.T) {
	tests := []struct {
		name       string
		url        string
		setupMock  func(m *mocks.MockOrderService)
		wantStatus int
	}{
		{
			name: "happy path — 200",
			url:  "/orders/1/cancel",
			setupMock: func(m *mocks.MockOrderService) {
				m.EXPECT().
					CancelOrder(gomock.Any(), int32(1)).
					Return(db.Order{ID: 1, Status: "cancelled"}, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			// Delivered order — service returns 400
			name: "cannot cancel delivered — 400",
			url:  "/orders/1/cancel",
			setupMock: func(m *mocks.MockOrderService) {
				m.EXPECT().
					CancelOrder(gomock.Any(), int32(1)).
					Return(db.Order{}, apperror.New(400, "cannot cancel a delivered order"))
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "order not found — 404",
			url:  "/orders/999/cancel",
			setupMock: func(m *mocks.MockOrderService) {
				m.EXPECT().
					CancelOrder(gomock.Any(), int32(999)).
					Return(db.Order{}, apperror.ErrOrderNotFound)
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "invalid id — 400",
			url:        "/orders/abc/cancel",
			setupMock:  func(m *mocks.MockOrderService) {},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, mockSvc := setupOrderHandler(t)
			tt.setupMock(mockSvc)

			w := doRequest(r, http.MethodPatch, tt.url, nil)

			if w.Code != tt.wantStatus {
				t.Errorf("got status %d, want %d — body: %s",
					w.Code, tt.wantStatus, w.Body.String())
			}
		})
	}
}

// =============================================
// PATCH /orders/:id/courier
// =============================================

func TestOrderHandler_AssignCourier(t *testing.T) {
	tests := []struct {
		name       string
		url        string
		body       interface{}
		setupMock  func(m *mocks.MockOrderService)
		wantStatus int
	}{
		{
			name: "happy path — 200",
			url:  "/orders/1/courier",
			body: map[string]interface{}{"courier_id": 1},
			setupMock: func(m *mocks.MockOrderService) {
				m.EXPECT().
					AssignCourier(gomock.Any(), int32(1), int32(1)).
					Return(db.Order{ID: 1, Status: "out_for_delivery"}, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			// Order not ready — service returns 400
			name: "order not ready — 400",
			url:  "/orders/1/courier",
			body: map[string]interface{}{"courier_id": 1},
			setupMock: func(m *mocks.MockOrderService) {
				m.EXPECT().
					AssignCourier(gomock.Any(), int32(1), int32(1)).
					Return(db.Order{}, apperror.New(400, "order not ready"))
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "missing body — 400",
			url:        "/orders/1/courier",
			body:       nil,
			setupMock:  func(m *mocks.MockOrderService) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid order id — 400",
			url:        "/orders/abc/courier",
			body:       map[string]interface{}{"courier_id": 1},
			setupMock:  func(m *mocks.MockOrderService) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "order not found — 404",
			url:  "/orders/999/courier",
			body: map[string]interface{}{"courier_id": 1},
			setupMock: func(m *mocks.MockOrderService) {
				m.EXPECT().
					AssignCourier(gomock.Any(), int32(999), int32(1)).
					Return(db.Order{}, apperror.ErrOrderNotFound)
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, mockSvc := setupOrderHandler(t)
			tt.setupMock(mockSvc)

			w := doRequest(r, http.MethodPatch, tt.url, tt.body)

			if w.Code != tt.wantStatus {
				t.Errorf("got status %d, want %d — body: %s",
					w.Code, tt.wantStatus, w.Body.String())
			}
		})
	}
}

// =============================================
// GET /couriers
// =============================================

func TestOrderHandler_GetAllCouriers(t *testing.T) {
	tests := []struct {
		name       string
		setupMock  func(m *mocks.MockOrderService)
		wantStatus int
	}{
		{
			name: "happy path — 200",
			setupMock: func(m *mocks.MockOrderService) {
				m.EXPECT().
					GetAllCouriers(gomock.Any()).
					Return([]db.Courier{{ID: 1, Name: "Ali"}}, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "empty list — still 200",
			setupMock: func(m *mocks.MockOrderService) {
				m.EXPECT().
					GetAllCouriers(gomock.Any()).
					Return([]db.Courier{}, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "database error — 500",
			setupMock: func(m *mocks.MockOrderService) {
				m.EXPECT().
					GetAllCouriers(gomock.Any()).
					Return(nil, errors.New("db error"))
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, mockSvc := setupOrderHandler(t)
			tt.setupMock(mockSvc)

			w := doRequest(r, http.MethodGet, "/couriers", nil)

			if w.Code != tt.wantStatus {
				t.Errorf("got status %d, want %d — body: %s",
					w.Code, tt.wantStatus, w.Body.String())
			}
		})
	}
}

// =============================================
// PATCH /couriers/:id/status
// =============================================

func TestOrderHandler_UpdateCourierStatus(t *testing.T) {
	tests := []struct {
		name       string
		url        string
		body       interface{}
		setupMock  func(m *mocks.MockOrderService)
		wantStatus int
	}{
		{
			name: "happy path — available",
			url:  "/couriers/1/status",
			body: map[string]interface{}{"status": "available"},
			setupMock: func(m *mocks.MockOrderService) {
				m.EXPECT().
					UpdateCourierStatus(gomock.Any(), int32(1), "available").
					Return(db.Courier{ID: 1, Status: "available"}, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			// Invalid status — service returns 400
			name: "invalid status — 400",
			url:  "/couriers/1/status",
			body: map[string]interface{}{"status": "flying"},
			setupMock: func(m *mocks.MockOrderService) {
				m.EXPECT().
					UpdateCourierStatus(gomock.Any(), int32(1), "flying").
					Return(db.Courier{}, apperror.New(400, "invalid courier status"))
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "missing body — 400",
			url:        "/couriers/1/status",
			body:       nil,
			setupMock:  func(m *mocks.MockOrderService) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid id — 400",
			url:        "/couriers/abc/status",
			body:       map[string]interface{}{"status": "available"},
			setupMock:  func(m *mocks.MockOrderService) {},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, mockSvc := setupOrderHandler(t)
			tt.setupMock(mockSvc)

			w := doRequest(r, http.MethodPatch, tt.url, tt.body)

			if w.Code != tt.wantStatus {
				t.Errorf("got status %d, want %d — body: %s",
					w.Code, tt.wantStatus, w.Body.String())
			}
		})
	}
}