package handlers

import (
	"bar108/internal/db"
	"bar108/internal/repository"
	"context"
)

type menuService interface {
	GetAllMenuItems(ctx context.Context) ([]db.GetAllMenuItemsRow, error)
	GetMenuItemByID(ctx context.Context, id int32) (db.GetMenuItemByIDRow, error)
	GetAllCategories(ctx context.Context) ([]db.Category, error)
	CreateMenuItem(ctx context.Context, arg db.CreateMenuItemParams) (db.MenuItem, error)
	UpdateMenuItem(ctx context.Context, arg db.UpdateMenuItemParams) (db.MenuItem, error)
	DeleteMenuItem(ctx context.Context, id int32) error
}

type userService interface {
	CreateUser(ctx context.Context, arg db.CreateUserParams) (db.User, error)
	GetAllUsers(ctx context.Context) ([]db.User, error)
	GetActiveUsers(ctx context.Context) ([]db.User, error)
	GetUserByID(ctx context.Context, id int32) (db.User, error)
	GetUserByEmail(ctx context.Context, email string) (db.User, error)
	GetUserByPhone(ctx context.Context, phone string) (db.User, error)
	UpdateUser(ctx context.Context, arg db.UpdateUserParams) (db.User, error)
	UpdateUserBonusPoints(ctx context.Context, id, bonusPoints int32) (db.User, error)
	DeactivateUser(ctx context.Context, id int32) (db.User, error)
	ActivateUser(ctx context.Context, id int32) (db.User, error)
}

// orderService is what the order handler needs from the service layer.
// Private — only this package sees it.
// The concrete *orderService satisfies this implicitly.
type orderService interface {
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
