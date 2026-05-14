package services_test

import (
	"bar108/internal/apperror"
	"bar108/internal/db"
	"bar108/internal/repository"
	"bar108/internal/repository/mocks"
	"bar108/internal/services"
	"context"
	"errors"
	"testing"

	"go.uber.org/mock/gomock"
)

// setupOrderService creates a fresh mock + service for each test.
func setupOrderService(t *testing.T) (services.OrderService, *mocks.MockOrderRepository) {
	t.Helper()
	ctrl := gomock.NewController(t)
	mockRepo := mocks.NewMockOrderRepository(ctrl)
	svc := services.NewOrderService(mockRepo)
	return svc, mockRepo
}

// =============================================
// PlaceOrder
// =============================================

func TestOrderService_PlaceOrder(t *testing.T) {
	validInput := repository.PlaceOrderInput{
		UserID: 1,
		Items: []repository.OrderItemInput{
			{MenuItemID: 1, Quantity: 2},
		},
	}

	tests := []struct {
		name      string
		input     repository.PlaceOrderInput
		setupMock func(m *mocks.MockOrderRepository)
		wantErr   error
	}{
		{
			name:  "happy path",
			input: validInput,
			setupMock: func(m *mocks.MockOrderRepository) {
				m.EXPECT().
					PlaceOrder(gomock.Any(), validInput).
					Return(repository.PlaceOrderResult{
						Order:      db.Order{ID: 1, Status: "pending"},
						TotalPrice: "350.00",
						FinalPrice: "350.00",
					}, nil)
			},
			wantErr: nil,
		},
		{
			// Empty items — service catches this before hitting DB
			name: "empty items list",
			input: repository.PlaceOrderInput{
				UserID: 1,
				Items:  []repository.OrderItemInput{},
			},
			setupMock: func(m *mocks.MockOrderRepository) {},
			wantErr:   apperror.ErrInvalidInput,
		},
		{
			// Invalid user ID — caught before DB
			name: "invalid user id — zero",
			input: repository.PlaceOrderInput{
				UserID: 0,
				Items:  []repository.OrderItemInput{{MenuItemID: 1, Quantity: 1}},
			},
			setupMock: func(m *mocks.MockOrderRepository) {},
			wantErr:   apperror.ErrInvalidID,
		},
		{
			// Invalid menu item ID
			name: "invalid menu item id",
			input: repository.PlaceOrderInput{
				UserID: 1,
				Items:  []repository.OrderItemInput{{MenuItemID: 0, Quantity: 1}},
			},
			setupMock: func(m *mocks.MockOrderRepository) {},
			wantErr:   apperror.ErrInvalidID,
		},
		{
			// Zero quantity — you can't order 0 burgers
			name: "zero quantity",
			input: repository.PlaceOrderInput{
				UserID: 1,
				Items:  []repository.OrderItemInput{{MenuItemID: 1, Quantity: 0}},
			},
			setupMock: func(m *mocks.MockOrderRepository) {},
			wantErr:   apperror.ErrInvalidInput,
		},
		{
			// Negative quantity
			name: "negative quantity",
			input: repository.PlaceOrderInput{
				UserID: 1,
				Items:  []repository.OrderItemInput{{MenuItemID: 1, Quantity: -1}},
			},
			setupMock: func(m *mocks.MockOrderRepository) {},
			wantErr:   apperror.ErrInvalidInput,
		},
		{
			// Item not found — caught by repository inside transaction
			name:  "menu item not found",
			input: validInput,
			setupMock: func(m *mocks.MockOrderRepository) {
				m.EXPECT().
					PlaceOrder(gomock.Any(), validInput).
					Return(repository.PlaceOrderResult{}, apperror.ErrMenuNotFound)
			},
			wantErr: apperror.ErrMenuNotFound,
		},
		{
			// Item unavailable — caught by repository
			name:  "menu item unavailable",
			input: validInput,
			setupMock: func(m *mocks.MockOrderRepository) {
				m.EXPECT().
					PlaceOrder(gomock.Any(), validInput).
					Return(repository.PlaceOrderResult{}, apperror.ErrItemUnavailable)
			},
			wantErr: apperror.ErrItemUnavailable,
		},
		{
			// Promo not found
			name: "promo code not found",
			input: repository.PlaceOrderInput{
				UserID:    1,
				Items:     []repository.OrderItemInput{{MenuItemID: 1, Quantity: 1}},
				PromoCode: "INVALID",
			},
			setupMock: func(m *mocks.MockOrderRepository) {
				m.EXPECT().
					PlaceOrder(gomock.Any(), gomock.Any()).
					Return(repository.PlaceOrderResult{}, apperror.ErrPromoNotFound)
			},
			wantErr: apperror.ErrPromoNotFound,
		},
		{
			// DB error
			name:  "database error",
			input: validInput,
			setupMock: func(m *mocks.MockOrderRepository) {
				m.EXPECT().
					PlaceOrder(gomock.Any(), validInput).
					Return(repository.PlaceOrderResult{}, errors.New("connection refused"))
			},
			wantErr: errors.New("db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, mockRepo := setupOrderService(t)
			tt.setupMock(mockRepo)

			_, err := svc.PlaceOrder(context.Background(), tt.input)

			assertError(t, err, tt.wantErr)
		})
	}
}

// =============================================
// GetOrderByID
// =============================================

func TestOrderService_GetOrderByID(t *testing.T) {
	tests := []struct {
		name      string
		id        int32
		setupMock func(m *mocks.MockOrderRepository)
		wantErr   error
	}{
		{
			name: "happy path",
			id:   1,
			setupMock: func(m *mocks.MockOrderRepository) {
				m.EXPECT().
					GetOrderByID(gomock.Any(), int32(1)).
					Return(db.GetOrderByIDRow{ID: 1, Status: "pending"}, nil)
			},
			wantErr: nil,
		},
		{
			name: "not found",
			id:   999,
			setupMock: func(m *mocks.MockOrderRepository) {
				m.EXPECT().
					GetOrderByID(gomock.Any(), int32(999)).
					Return(db.GetOrderByIDRow{}, apperror.ErrOrderNotFound)
			},
			wantErr: apperror.ErrOrderNotFound,
		},
		{
			// Invalid ID — repo never called
			name:      "invalid id — zero",
			id:        0,
			setupMock: func(m *mocks.MockOrderRepository) {},
			wantErr:   apperror.ErrInvalidID,
		},
		{
			name:      "invalid id — negative",
			id:        -1,
			setupMock: func(m *mocks.MockOrderRepository) {},
			wantErr:   apperror.ErrInvalidID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, mockRepo := setupOrderService(t)
			tt.setupMock(mockRepo)

			_, err := svc.GetOrderByID(context.Background(), tt.id)

			assertError(t, err, tt.wantErr)
		})
	}
}

// =============================================
// UpdateOrderStatus
// =============================================

func TestOrderService_UpdateOrderStatus(t *testing.T) {
	tests := []struct {
		name      string
		orderID   int32
		newStatus string
		setupMock func(m *mocks.MockOrderRepository)
		wantErr   error
	}{
		{
			// pending → confirmed is a valid transition
			name:      "happy path — pending to confirmed",
			orderID:   1,
			newStatus: "confirmed",
			setupMock: func(m *mocks.MockOrderRepository) {
				m.EXPECT().
					GetOrderByID(gomock.Any(), int32(1)).
					Return(db.GetOrderByIDRow{ID: 1, Status: "pending"}, nil)
				m.EXPECT().
					UpdateOrderStatus(gomock.Any(), int32(1), "confirmed").
					Return(db.Order{ID: 1, Status: "confirmed"}, nil)
			},
			wantErr: nil,
		},
		{
			// confirmed → preparing is valid
			name:      "happy path — confirmed to preparing",
			orderID:   1,
			newStatus: "preparing",
			setupMock: func(m *mocks.MockOrderRepository) {
				m.EXPECT().
					GetOrderByID(gomock.Any(), int32(1)).
					Return(db.GetOrderByIDRow{ID: 1, Status: "confirmed"}, nil)
				m.EXPECT().
					UpdateOrderStatus(gomock.Any(), int32(1), "preparing").
					Return(db.Order{ID: 1, Status: "preparing"}, nil)
			},
			wantErr: nil,
		},
		{
			// delivered → pending is NOT valid — terminal state
			// UpdateOrderStatus on repo must NEVER be called
			name:      "invalid transition — delivered to pending",
			orderID:   1,
			newStatus: "pending",
			setupMock: func(m *mocks.MockOrderRepository) {
				m.EXPECT().
					GetOrderByID(gomock.Any(), int32(1)).
					Return(db.GetOrderByIDRow{ID: 1, Status: "delivered"}, nil)
				// No EXPECT for UpdateOrderStatus — gomock fails if it's called
			},
			wantErr: apperror.ErrInvalidInput,
		},
		{
			// cancelled → confirmed is NOT valid
			name:      "invalid transition — cancelled to confirmed",
			orderID:   1,
			newStatus: "confirmed",
			setupMock: func(m *mocks.MockOrderRepository) {
				m.EXPECT().
					GetOrderByID(gomock.Any(), int32(1)).
					Return(db.GetOrderByIDRow{ID: 1, Status: "cancelled"}, nil)
			},
			wantErr: apperror.ErrInvalidInput,
		},
		{
			// pending → delivered is NOT valid — skips steps
			name:      "invalid transition — pending to delivered",
			orderID:   1,
			newStatus: "delivered",
			setupMock: func(m *mocks.MockOrderRepository) {
				m.EXPECT().
					GetOrderByID(gomock.Any(), int32(1)).
					Return(db.GetOrderByIDRow{ID: 1, Status: "pending"}, nil)
			},
			wantErr: apperror.ErrInvalidInput,
		},
		{
			// Completely unknown status
			name:      "unknown status",
			orderID:   1,
			newStatus: "flying",
			setupMock: func(m *mocks.MockOrderRepository) {},
			wantErr:   apperror.ErrInvalidInput,
		},
		{
			name:      "invalid order id",
			orderID:   0,
			newStatus: "confirmed",
			setupMock: func(m *mocks.MockOrderRepository) {},
			wantErr:   apperror.ErrInvalidID,
		},
		{
			name:      "order not found",
			orderID:   999,
			newStatus: "confirmed",
			setupMock: func(m *mocks.MockOrderRepository) {
				m.EXPECT().
					GetOrderByID(gomock.Any(), int32(999)).
					Return(db.GetOrderByIDRow{}, apperror.ErrOrderNotFound)
			},
			wantErr: apperror.ErrOrderNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, mockRepo := setupOrderService(t)
			tt.setupMock(mockRepo)

			_, err := svc.UpdateOrderStatus(context.Background(), tt.orderID, tt.newStatus)

			assertError(t, err, tt.wantErr)
		})
	}
}

// =============================================
// CancelOrder
// =============================================

func TestOrderService_CancelOrder(t *testing.T) {
	tests := []struct {
		name      string
		orderID   int32
		setupMock func(m *mocks.MockOrderRepository)
		wantErr   error
	}{
		{
			name:    "happy path — pending order cancelled",
			orderID: 1,
			setupMock: func(m *mocks.MockOrderRepository) {
				m.EXPECT().
					GetOrderByID(gomock.Any(), int32(1)).
					Return(db.GetOrderByIDRow{ID: 1, Status: "pending"}, nil)
				m.EXPECT().
					CancelOrder(gomock.Any(), int32(1)).
					Return(db.Order{ID: 1, Status: "cancelled"}, nil)
			},
			wantErr: nil,
		},
		{
			name:    "happy path — confirmed order cancelled",
			orderID: 1,
			setupMock: func(m *mocks.MockOrderRepository) {
				m.EXPECT().
					GetOrderByID(gomock.Any(), int32(1)).
					Return(db.GetOrderByIDRow{ID: 1, Status: "confirmed"}, nil)
				m.EXPECT().
					CancelOrder(gomock.Any(), int32(1)).
					Return(db.Order{ID: 1, Status: "cancelled"}, nil)
			},
			wantErr: nil,
		},
		{
			// Delivered order can NEVER be cancelled
			// CancelOrder on repo must never be called
			name:    "cannot cancel delivered order",
			orderID: 1,
			setupMock: func(m *mocks.MockOrderRepository) {
				m.EXPECT().
					GetOrderByID(gomock.Any(), int32(1)).
					Return(db.GetOrderByIDRow{ID: 1, Status: "delivered"}, nil)
			},
			wantErr: apperror.ErrInvalidInput,
		},
		{
			// Already cancelled
			name:    "cannot cancel already cancelled order",
			orderID: 1,
			setupMock: func(m *mocks.MockOrderRepository) {
				m.EXPECT().
					GetOrderByID(gomock.Any(), int32(1)).
					Return(db.GetOrderByIDRow{ID: 1, Status: "cancelled"}, nil)
			},
			wantErr: apperror.ErrInvalidInput,
		},
		{
			name:      "invalid id",
			orderID:   0,
			setupMock: func(m *mocks.MockOrderRepository) {},
			wantErr:   apperror.ErrInvalidID,
		},
		{
			name:    "order not found",
			orderID: 999,
			setupMock: func(m *mocks.MockOrderRepository) {
				m.EXPECT().
					GetOrderByID(gomock.Any(), int32(999)).
					Return(db.GetOrderByIDRow{}, apperror.ErrOrderNotFound)
			},
			wantErr: apperror.ErrOrderNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, mockRepo := setupOrderService(t)
			tt.setupMock(mockRepo)

			_, err := svc.CancelOrder(context.Background(), tt.orderID)

			assertError(t, err, tt.wantErr)
		})
	}
}

// =============================================
// AssignCourier
// =============================================

func TestOrderService_AssignCourier(t *testing.T) {
	tests := []struct {
		name      string
		orderID   int32
		courierID int32
		setupMock func(m *mocks.MockOrderRepository)
		wantErr   error
	}{
		{
			// Order is ready → courier can be assigned
			name:      "happy path",
			orderID:   1,
			courierID: 1,
			setupMock: func(m *mocks.MockOrderRepository) {
				m.EXPECT().
					GetOrderByID(gomock.Any(), int32(1)).
					Return(db.GetOrderByIDRow{ID: 1, Status: "ready"}, nil)
				m.EXPECT().
					GetCourierByID(gomock.Any(), int32(1)).
					Return(db.Courier{ID: 1, Name: "Ali"}, nil)
				m.EXPECT().
					AssignCourier(gomock.Any(), int32(1), int32(1)).
					Return(db.Order{ID: 1, Status: "out_for_delivery"}, nil)
			},
			wantErr: nil,
		},
		{
			// Order is still pending — can't assign courier yet
			// AssignCourier on repo must never be called
			name:      "order not ready — pending",
			orderID:   1,
			courierID: 1,
			setupMock: func(m *mocks.MockOrderRepository) {
				m.EXPECT().
					GetOrderByID(gomock.Any(), int32(1)).
					Return(db.GetOrderByIDRow{ID: 1, Status: "pending"}, nil)
			},
			wantErr: apperror.ErrInvalidInput,
		},
		{
			// Order is preparing — still can't assign courier
			name:      "order not ready — preparing",
			orderID:   1,
			courierID: 1,
			setupMock: func(m *mocks.MockOrderRepository) {
				m.EXPECT().
					GetOrderByID(gomock.Any(), int32(1)).
					Return(db.GetOrderByIDRow{ID: 1, Status: "preparing"}, nil)
			},
			wantErr: apperror.ErrInvalidInput,
		},
		{
			// Courier doesn't exist
			name:      "courier not found",
			orderID:   1,
			courierID: 999,
			setupMock: func(m *mocks.MockOrderRepository) {
				m.EXPECT().
					GetOrderByID(gomock.Any(), int32(1)).
					Return(db.GetOrderByIDRow{ID: 1, Status: "ready"}, nil)
				m.EXPECT().
					GetCourierByID(gomock.Any(), int32(999)).
					Return(db.Courier{}, apperror.ErrNotFound)
			},
			wantErr: apperror.ErrNotFound,
		},
		{
			name:      "invalid order id",
			orderID:   0,
			courierID: 1,
			setupMock: func(m *mocks.MockOrderRepository) {},
			wantErr:   apperror.ErrInvalidID,
		},
		{
			name:      "invalid courier id",
			orderID:   1,
			courierID: 0,
			setupMock: func(m *mocks.MockOrderRepository) {},
			wantErr:   apperror.ErrInvalidID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, mockRepo := setupOrderService(t)
			tt.setupMock(mockRepo)

			_, err := svc.AssignCourier(context.Background(), tt.orderID, tt.courierID)

			assertError(t, err, tt.wantErr)
		})
	}
}

// =============================================
// UpdateCourierStatus
// =============================================

func TestOrderService_UpdateCourierStatus(t *testing.T) {
	tests := []struct {
		name      string
		id        int32
		status    string
		setupMock func(m *mocks.MockOrderRepository)
		wantErr   error
	}{
		{
			name:   "happy path — available",
			id:     1,
			status: "available",
			setupMock: func(m *mocks.MockOrderRepository) {
				m.EXPECT().
					UpdateCourierStatus(gomock.Any(), int32(1), "available").
					Return(db.Courier{ID: 1, Status: "available"}, nil)
			},
			wantErr: nil,
		},
		{
			name:   "happy path — busy",
			id:     1,
			status: "busy",
			setupMock: func(m *mocks.MockOrderRepository) {
				m.EXPECT().
					UpdateCourierStatus(gomock.Any(), int32(1), "busy").
					Return(db.Courier{ID: 1, Status: "busy"}, nil)
			},
			wantErr: nil,
		},
		{
			name:   "happy path — offline",
			id:     1,
			status: "offline",
			setupMock: func(m *mocks.MockOrderRepository) {
				m.EXPECT().
					UpdateCourierStatus(gomock.Any(), int32(1), "offline").
					Return(db.Courier{ID: 1, Status: "offline"}, nil)
			},
			wantErr: nil,
		},
		{
			// Unknown status — service rejects before hitting DB
			name:      "invalid status",
			id:        1,
			status:    "flying",
			setupMock: func(m *mocks.MockOrderRepository) {},
			wantErr:   apperror.ErrInvalidInput,
		},
		{
			name:      "invalid id",
			id:        0,
			status:    "available",
			setupMock: func(m *mocks.MockOrderRepository) {},
			wantErr:   apperror.ErrInvalidID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, mockRepo := setupOrderService(t)
			tt.setupMock(mockRepo)

			_, err := svc.UpdateCourierStatus(context.Background(), tt.id, tt.status)

			assertError(t, err, tt.wantErr)
		})
	}
}

// =============================================
// assertError — shared helper
//
// Why extract this? Every test does the same check:
// "if wantErr is an AppError, verify HTTP code matches.
// if wantErr is any other error, just verify error is non-nil."
// Extracting it removes 10+ lines of duplication per test.
// =============================================

func assertError(t *testing.T, got error, want error) {
	t.Helper()

	if want == nil {
		if got != nil {
			t.Errorf("expected no error but got: %v", got)
		}
		return
	}

	// We expected an error
	if got == nil {
		t.Fatalf("expected error but got nil")
	}

	// If expected error is an AppError, verify HTTP code matches
	var wantAppErr *apperror.AppError
	if errors.As(want, &wantAppErr) {
		var gotAppErr *apperror.AppError
		if !errors.As(got, &gotAppErr) {
			t.Errorf("expected *apperror.AppError but got %T: %v", got, got)
			return
		}
		if gotAppErr.Code != wantAppErr.Code {
			t.Errorf("got HTTP code %d, want %d (error: %v)", gotAppErr.Code, wantAppErr.Code, got)
		}
		return
	}

	// For non-AppError cases, just verify an error was returned
	// We don't check the exact message — it may come from deep in the stack
}