package repository

import (
	db "bar108/internal/db"
	"context"
	"database/sql"
	"errors"
	"fmt"
)

// MenuRepository defines all database operations for the menu.

// MenuRepository is the concrete implementation — the real one that talks to PostgreSQL.

type MenuRepository struct {
	queries *db.Queries
}

func NewMenuRepository(conn *sql.DB) *MenuRepository {
	return &MenuRepository{
		queries: db.New(conn),
	}
}
func (r *MenuRepository) GetAllMenuItems(ctx context.Context) ([]db.GetAllMenuItemsRow, error) {
	items, err := r.queries.GetAllMenuItems(ctx)
	if err != nil {
		return nil, fmt.Errorf("GetAllMenuItems: %w", err)
	}
	return items, nil
}
func (r *MenuRepository) GetMenuItemByID(ctx context.Context, id int32) (db.GetMenuItemByIDRow, error) {
	item, err := r.queries.GetMenuItemByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return db.GetMenuItemByIDRow{}, fmt.Errorf("menu item %d not found: %w", id, sql.ErrNoRows)
		}
		return db.GetMenuItemByIDRow{}, fmt.Errorf("GetMenuItemByID: %w", err)
	}
	return item, nil
}
func (r *MenuRepository) GetAllCategories(ctx context.Context) ([]db.Category, error) {
	categories, err := r.queries.GetAllCategories(ctx)
	if err != nil {
		return nil, fmt.Errorf("GetAllCategories: %w", err)
	}
	return categories, nil
}
func (r *MenuRepository) CreateMenuItem(ctx context.Context, arg db.CreateMenuItemParams) (db.MenuItem, error) {
	item, err := r.queries.CreateMenuItem(ctx, arg)
	if err != nil {
		return db.MenuItem{}, fmt.Errorf("CreateMenuItem: %w", err)
	}
	return item, nil
}
func (r *MenuRepository) UpdateMenuItem(ctx context.Context, arg db.UpdateMenuItemParams) (db.MenuItem, error) {
	item, err := r.queries.UpdateMenuItem(ctx, arg)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return db.MenuItem{}, fmt.Errorf("menu item %d not found: %w", arg.ID, sql.ErrNoRows)

		}
		return db.MenuItem{}, fmt.Errorf("UpdateMenuItem: %w", err)

	}
	return item, nil
}
func (r *MenuRepository) DeleteMenuItem(ctx context.Context, id int32) error {
	err := r.queries.DeleteMenuItem(ctx, id)
	if err != nil {
		return fmt.Errorf("DeleteMenuItem: %w", err)
	}
	return nil
}
