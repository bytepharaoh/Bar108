package repository

import (
	db "bar108/internal/db"
	"context"
	"database/sql"
	"fmt"
)

// MenuRepository defines all database operations for the menu.
type MenuRepository interface {
	GetAllMenuItems(ctx context.Context) ([]db.GetAllMenuItemsRow, error)
	GetMenuItemByID(ctx context.Context, id int32) (db.GetMenuItemByIDRow, error)
	GetAllCategories(ctx context.Context) ([]db.Category, error)
	CreateMenuItem(ctx context.Context, arg db.CreateMenuItemParams) (db.MenuItem, error)
	UpdateMenuItem(ctx context.Context, arg db.UpdateMenuItemParams) (db.MenuItem, error)
	DeleteMenuItem(ctx context.Context, id int32) error
}

// menuRepository is the concrete implementation — the real one that talks to PostgreSQL.

type menuRepository struct {
	queries *db.Queries
}

func NewMenuRepository(conn *sql.DB) MenuRepository {
	return &menuRepository{
		queries: db.New(conn),
	}
}
func (r *menuRepository) GetAllMenuItems(ctx context.Context) ([]db.GetAllMenuItemsRow, error) {
	items, err := r.queries.GetAllMenuItems(ctx)
	if err != nil {
		return nil, fmt.Errorf("GetAllMenuItems: %w", err)
	}
	return items, nil
}
func (r *menuRepository) GetMenuItemByID(ctx context.Context, id int32) (db.GetMenuItemByIDRow, error) {
	item, err := r.queries.GetMenuItemByID(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return db.GetMenuItemByIDRow{}, fmt.Errorf("menu item %d not found: %w", id, sql.ErrNoRows)
		}
		return db.GetMenuItemByIDRow{}, fmt.Errorf("GetMenuItemByID: %w", err)
	}
	return item, nil
}
func (r *menuRepository) GetAllCategories(ctx context.Context) ([]db.Category, error) {
	categories, err := r.queries.GetAllCategories(ctx)
	if err != nil {
		return nil, fmt.Errorf("GetAllCategories: %w", err)
	}
	return categories, nil
}
func (r *menuRepository) CreateMenuItem(ctx context.Context, arg db.CreateMenuItemParams) (db.MenuItem, error) {
	item, err := r.queries.CreateMenuItem(ctx, arg)
	if err != nil {
		return db.MenuItem{}, fmt.Errorf("CreateMenuItem: %w", err)
	}
	return item, nil
}
func (r *menuRepository) UpdateMenuItem(ctx context.Context, arg db.UpdateMenuItemParams) (db.MenuItem, error) {
	item, err := r.queries.UpdateMenuItem(ctx, arg)
	if err != nil {
		if err == sql.ErrNoRows {
			return db.MenuItem{}, fmt.Errorf("menu item %d not found: %w", arg.ID, sql.ErrNoRows)

		}
		return db.MenuItem{}, fmt.Errorf("UpdateMenuItem: %w", err)

	}
	return item, nil
}
func (r *menuRepository) DeleteMenuItem(ctx context.Context, id int32) error {
	err := r.queries.DeleteMenuItem(ctx, id)
	if err != nil {
		return fmt.Errorf("DeleteMenuItem: %w", err)
	}
	return nil
}
