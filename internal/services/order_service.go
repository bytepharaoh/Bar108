package services

import (
	"bar108/internal/apperror"
	"bar108/internal/db"
	"bar108/internal/repository"
	"context"
	"fmt"
)

// orderStore — the private interface this service
// defines for itself. Only the methods it actually
// needs from the repository.

type orderStore interface {
	PlaceOrder(ctx context.Context, input repository.PlaceOrderInput) (repository.PlaceOrderResult, error)
	GetOrderByID(ctx context.Context, id int32) (db.GetOrderByIDRow, error)
	GetOrdersByUserID(ctx context.Context, userID int32) ([]db.GetOrdersByUserIDRow, error)
	GetAllOrders(ctx context.Context) ([]db.GetAllOrdersRow, error)
	GetPendingOrders(ctx context.Context) ([]db.GetPendingOrdersRow, error)
	GetOrderItems(ctx context.Context, orderID int32) ([]db.GetOrderItemsRow, error)
	GetOrderStatusHistory(ctx context.Context, orderID int32) ([]db.OrderStatusHistory, error)
	UpdateOrderStatus(ctx context.Context, id int32, status string) (db.Order, error)
	AssignCourier(ctx context.Context, orderID int32, courierID int32) (db.Order, error)
	CancelOrder(ctx context.Context, id int32) (db.Order, error)
	GetAllCouriers(ctx context.Context) ([]db.Courier, error)
	GetAvailableCouriers(ctx context.Context) ([]db.Courier, error)
	GetCourierByID(ctx context.Context, id int32) (db.Courier, error)
	UpdateCourierStatus(ctx context.Context, id int32, status string) (db.Courier, error)
}

// OrderService — the exported interface the
// handler will depend on (via its own private
// version defined in handlers/interfaces.go)

type OrderService interface {
	PlaceOrder(ctx context.Context, input repository.PlaceOrderInput) (repository.PlaceOrderResult, error)
	GetOrderByID(ctx context.Context, id int32) (db.GetOrderByIDRow, error)
	GetOrdersByUserID(ctx context.Context, userID int32) ([]db.GetOrdersByUserIDRow, error)
	GetAllOrders(ctx context.Context) ([]db.GetAllOrdersRow, error)
	GetPendingOrders(ctx context.Context) ([]db.GetPendingOrdersRow, error)
	GetOrderItems(ctx context.Context, orderID int32) ([]db.GetOrderItemsRow, error)
	GetOrderStatusHistory(ctx context.Context, orderID int32) ([]db.OrderStatusHistory, error)
	UpdateOrderStatus(ctx context.Context, orderID int32, status string) (db.Order, error)
	AssignCourier(ctx context.Context, orderID int32, courierID int32) (db.Order, error)
	CancelOrder(ctx context.Context, orderID int32) (db.Order, error)
	GetAllCouriers(ctx context.Context) ([]db.Courier, error)
	GetAvailableCouriers(ctx context.Context) ([]db.Courier, error)
	GetCourierByID(ctx context.Context, id int32) (db.Courier, error)
	UpdateCourierStatus(ctx context.Context, id int32, status string) (db.Courier, error)
}

// =============================================
// Concrete implementation
// =============================================

type orderService struct {
	store orderStore
}

func NewOrderService(store orderStore) OrderService {
	return &orderService{store: store}
}

var validStatuses = map[string]bool{
	"pending":          true,
	"confirmed":        true,
	"preparing":        true,
	"ready":            true,
	"out_for_delivery": true,
	"delivered":        true,
	"cancelled":        true,
}

var allowedTransitions = map[string]map[string]bool{
	"pending": {
		"confirmed": true,
		"cancelled": true,
	},
	"confirmed": {
		"preparing": true,
		"cancelled": true,
	},
	"preparing": {
		"ready":     true,
		"cancelled": true,
	},
	"ready": {
		"out_for_delivery": true,
		"cancelled":        true,
	},
	"out_for_delivery": {
		"delivered": true,
		"cancelled": true,
	},
	// Terminal states — no transitions allowed FROM these
	"delivered": {},
	"cancelled": {},
}

// Validation helpers

func validateOrderID(id int32) error {
	if id <= 0 {
		return apperror.ErrInvalidID
	}
	return nil
}

func validateStatusTransition(currentStatus, newStatus string) error {
	allowed, exists := allowedTransitions[currentStatus]
	if !exists {
		return apperror.New(400, fmt.Sprintf("unknown current status: %s", currentStatus))
	}
	if !allowed[newStatus] {
		return apperror.New(400, fmt.Sprintf(
			"cannot transition from '%s' to '%s'", currentStatus, newStatus,
		))
	}
	return nil
}

// PlaceOrder — validates input then delegates
// to the repository for the actual transaction.

func (s *orderService) PlaceOrder(ctx context.Context, input repository.PlaceOrderInput) (repository.PlaceOrderResult, error) {
	// Validate user ID
	if input.UserID <= 0 {
		return repository.PlaceOrderResult{}, apperror.ErrInvalidID
	}

	// Validate items list — an order must have at least one item
	if len(input.Items) == 0 {
		return repository.PlaceOrderResult{}, apperror.New(400, "order must contain at least one item")
	}

	// Validate each item
	for _, item := range input.Items {
		if item.MenuItemID <= 0 {
			return repository.PlaceOrderResult{}, apperror.ErrInvalidID
		}
		// Quantity must be positive — you can't order 0 or negative burgers
		if item.Quantity <= 0 {
			return repository.PlaceOrderResult{}, apperror.New(400, "item quantity must be greater than zero")
		}
	}

	// Delegate to repository — it handles the transaction
	result, err := s.store.PlaceOrder(ctx, input)
	if err != nil {
		return repository.PlaceOrderResult{}, err
	}
	return result, nil
}

func (s *orderService) GetOrderByID(ctx context.Context, id int32) (db.GetOrderByIDRow, error) {
	if err := validateOrderID(id); err != nil {
		return db.GetOrderByIDRow{}, err
	}
	order, err := s.store.GetOrderByID(ctx, id)
	if err != nil {
		return db.GetOrderByIDRow{}, err
	}
	return order, nil
}

func (s *orderService) GetOrdersByUserID(ctx context.Context, userID int32) ([]db.GetOrdersByUserIDRow, error) {
	if err := validateOrderID(userID); err != nil {
		return nil, err
	}
	orders, err := s.store.GetOrdersByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("GetOrdersByUserID service: %w", err)
	}
	return orders, nil
}

func (s *orderService) GetAllOrders(ctx context.Context) ([]db.GetAllOrdersRow, error) {
	orders, err := s.store.GetAllOrders(ctx)
	if err != nil {
		return nil, fmt.Errorf("GetAllOrders service: %w", err)
	}
	return orders, nil
}

func (s *orderService) GetPendingOrders(ctx context.Context) ([]db.GetPendingOrdersRow, error) {
	orders, err := s.store.GetPendingOrders(ctx)
	if err != nil {
		return nil, fmt.Errorf("GetPendingOrders service: %w", err)
	}
	return orders, nil
}

func (s *orderService) GetOrderItems(ctx context.Context, orderID int32) ([]db.GetOrderItemsRow, error) {
	if err := validateOrderID(orderID); err != nil {
		return nil, err
	}
	items, err := s.store.GetOrderItems(ctx, orderID)
	if err != nil {
		return nil, fmt.Errorf("GetOrderItems service: %w", err)
	}
	return items, nil
}

func (s *orderService) GetOrderStatusHistory(ctx context.Context, orderID int32) ([]db.OrderStatusHistory, error) {
	if err := validateOrderID(orderID); err != nil {
		return nil, err
	}
	history, err := s.store.GetOrderStatusHistory(ctx, orderID)
	if err != nil {
		return nil, fmt.Errorf("GetOrderStatusHistory service: %w", err)
	}
	return history, nil
}

func (s *orderService) UpdateOrderStatus(ctx context.Context, orderID int32, newStatus string) (db.Order, error) {
	if err := validateOrderID(orderID); err != nil {
		return db.Order{}, err
	}

	// Validate the new status is a known value
	if !validStatuses[newStatus] {
		return db.Order{}, apperror.New(400, fmt.Sprintf("invalid status: %s", newStatus))
	}

	// Fetch current order to check its current status
	// We need to know where it IS before deciding if it can move
	order, err := s.store.GetOrderByID(ctx, orderID)
	if err != nil {
		return db.Order{}, err
	}

	// Validate the transition is allowed by business rules
	if err := validateStatusTransition(order.Status, newStatus); err != nil {
		return db.Order{}, err
	}

	// All good — update in the database
	updated, err := s.store.UpdateOrderStatus(ctx, orderID, newStatus)
	if err != nil {
		return db.Order{}, err
	}
	return updated, nil
}

func (s *orderService) AssignCourier(ctx context.Context, orderID int32, courierID int32) (db.Order, error) {
	if err := validateOrderID(orderID); err != nil {
		return db.Order{}, err
	}
	if courierID <= 0 {
		return db.Order{}, apperror.ErrInvalidID
	}

	// Business rule: you can only assign a courier when the order is ready
	// Fetching the order first lets us enforce this
	order, err := s.store.GetOrderByID(ctx, orderID)
	if err != nil {
		return db.Order{}, err
	}
	if order.Status != "ready" {
		return db.Order{}, apperror.New(400,
			fmt.Sprintf("courier can only be assigned when order is 'ready', current status: '%s'", order.Status),
		)
	}

	// Verify the courier actually exists
	_, err = s.store.GetCourierByID(ctx, courierID)
	if err != nil {
		return db.Order{}, err
	}

	// Assign — this also sets status to out_for_delivery in the DB query
	updated, err := s.store.AssignCourier(ctx, orderID, courierID)
	if err != nil {
		return db.Order{}, err
	}
	return updated, nil
}

func (s *orderService) CancelOrder(ctx context.Context, orderID int32) (db.Order, error) {
	if err := validateOrderID(orderID); err != nil {
		return db.Order{}, err
	}

	// Fetch current order to check if cancellation is allowed
	order, err := s.store.GetOrderByID(ctx, orderID)
	if err != nil {
		return db.Order{}, err
	}

	// Business rule: terminal states cannot be cancelled
	if order.Status == "delivered" {
		return db.Order{}, apperror.New(400, "cannot cancel a delivered order")
	}
	if order.Status == "cancelled" {
		return db.Order{}, apperror.New(400, "order is already cancelled")
	}

	cancelled, err := s.store.CancelOrder(ctx, orderID)
	if err != nil {
		return db.Order{}, err
	}
	return cancelled, nil
}

func (s *orderService) GetAllCouriers(ctx context.Context) ([]db.Courier, error) {
	couriers, err := s.store.GetAllCouriers(ctx)
	if err != nil {
		return nil, fmt.Errorf("GetAllCouriers service: %w", err)
	}
	return couriers, nil
}

func (s *orderService) GetAvailableCouriers(ctx context.Context) ([]db.Courier, error) {
	couriers, err := s.store.GetAvailableCouriers(ctx)
	if err != nil {
		return nil, fmt.Errorf("GetAvailableCouriers service: %w", err)
	}
	return couriers, nil
}

func (s *orderService) GetCourierByID(ctx context.Context, id int32) (db.Courier, error) {
	if id <= 0 {
		return db.Courier{}, apperror.ErrInvalidID
	}
	courier, err := s.store.GetCourierByID(ctx, id)
	if err != nil {
		return db.Courier{}, err
	}
	return courier, nil
}

func (s *orderService) UpdateCourierStatus(ctx context.Context, id int32, status string) (db.Courier, error) {
	if id <= 0 {
		return db.Courier{}, apperror.ErrInvalidID
	}

	// Validate courier status values
	validCourierStatuses := map[string]bool{
		"available": true,
		"busy":      true,
		"offline":   true,
	}
	if !validCourierStatuses[status] {
		return db.Courier{}, apperror.New(400, fmt.Sprintf("invalid courier status: %s", status))
	}

	courier, err := s.store.UpdateCourierStatus(ctx, id, status)
	if err != nil {
		return db.Courier{}, err
	}
	return courier, nil
}
