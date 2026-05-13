package handlers

import (
	"bar108/internal/db"
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

// userService is the private interface the user handler needs.
type userService interface {
	CreateUser(ctx context.Context, arg db.CreateUserParams) (db.User, error)
	GetAllUsers(ctx context.Context) ([]db.User, error)
	GetActiveUsers(ctx context.Context) ([]db.User, error)
	GetUserByID(ctx context.Context, id int32) (db.User, error)
	GetUserByEmail(ctx context.Context, email string) (db.User, error)
	GetUserByPhone(ctx context.Context, phone string) (db.User, error)
	UpdateUser(ctx context.Context, arg db.UpdateUserParams) (db.User, error)
	UpdateUserBonusPoints(ctx context.Context, id int32, bonusPoints int32) (db.User, error)
	DeactivateUser(ctx context.Context, id int32) (db.User, error)
	ActivateUser(ctx context.Context, id int32) (db.User, error)
}
