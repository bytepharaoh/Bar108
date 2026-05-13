package services

import (
	"bar108/internal/apperror"
	"bar108/internal/db"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

type menuStore interface {
	GetAllMenuItems(ctx context.Context) ([]db.GetAllMenuItemsRow, error)
	GetMenuItemByID(ctx context.Context, id int32) (db.GetMenuItemByIDRow, error)
	GetAllCategories(ctx context.Context) ([]db.Category, error)
	CreateMenuItem(ctx context.Context, arg db.CreateMenuItemParams) (db.MenuItem, error)
	UpdateMenuItem(ctx context.Context, arg db.UpdateMenuItemParams) (db.MenuItem, error)
	DeleteMenuItem(ctx context.Context, id int32) error
}

var (
	ErrMenuItemNotFound  = apperror.ErrMenuNotFound
	ErrCategoryNotFound  = apperror.ErrNotFound
	ErrInvalidMenuItemID = apperror.ErrInvalidID
	ErrInvalidCategoryID = apperror.ErrInvalidID
	ErrEmptyMenuItemName = apperror.ErrInvalidInput
	ErrNegativePrice     = apperror.ErrNegativePrice
	ErrZeroPrice         = apperror.ErrZeroPrice
)

//go:generate mockgen -source=menu_service.go -destination=mocks/menu_service_mock.go -package=mocks

type MenuService interface {
	GetAllMenuItems(ctx context.Context) ([]db.GetAllMenuItemsRow, error)
	GetMenuItemByID(ctx context.Context, id int32) (db.GetMenuItemByIDRow, error)
	GetAllCategories(ctx context.Context) ([]db.Category, error)
	CreateMenuItem(ctx context.Context, arg db.CreateMenuItemParams) (db.MenuItem, error)
	UpdateMenuItem(ctx context.Context, arg db.UpdateMenuItemParams) (db.MenuItem, error)
	DeleteMenuItem(ctx context.Context, id int32) error
}
type menuService struct {
	store menuStore
}

func NewMenuService(store menuStore) MenuService {
	return &menuService{
		store: store,
	}

}
func validateMenuItemId(id int32) error {
	if id <= 0 {
		return ErrInvalidMenuItemID
	}
	return nil
}
func validateCategoryItemId(id int32) error {
	if id <= 0 {
		return ErrInvalidCategoryID
	}
	return nil
}
func validateMenuItemName(name string) error {
	if strings.TrimSpace(name) == "" {
		return ErrEmptyMenuItemName
	}
	return nil
}
func validatePrice(price string) error {
	trimmed := strings.TrimSpace(price)
	if trimmed == "" || trimmed == "0" || trimmed == "0.00" {
		return ErrZeroPrice
	}
	if strings.HasPrefix(trimmed, "-") {
		return ErrNegativePrice
	}
	return nil

}
func (s *menuService) GetAllMenuItems(ctx context.Context) ([]db.GetAllMenuItemsRow, error) {
	items, err := s.store.GetAllMenuItems(ctx)
	if err != nil {
		return nil, fmt.Errorf("GetAllMenuItems service: %w", err)
	}
	return items, nil
}
func (s *menuService) GetMenuItemByID(ctx context.Context, id int32) (db.GetMenuItemByIDRow, error) {
	if err := validateMenuItemId(id); err != nil {
		return db.GetMenuItemByIDRow{}, err
	}
	item, err := s.store.GetMenuItemByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return db.GetMenuItemByIDRow{}, ErrMenuItemNotFound
		}
		return db.GetMenuItemByIDRow{}, fmt.Errorf("GetMenuItemByID service: %w", err)
	}
	return item, nil

}
func (s *menuService) GetAllCategories(ctx context.Context) ([]db.Category, error) {
	categories, err := s.store.GetAllCategories(ctx)
	if err != nil {
		return nil, fmt.Errorf("GetAllCategories service: %w", err)
	}
	return categories, nil
}
func (s *menuService) CreateMenuItem(ctx context.Context, arg db.CreateMenuItemParams) (db.MenuItem, error) {
	if err := validateCategoryItemId(arg.CategoryID); err != nil {
		return db.MenuItem{}, err
	}
	if err := validateMenuItemName(arg.Name); err != nil {
		return db.MenuItem{}, err
	}
	if err := validatePrice(arg.Price); err != nil {
		return db.MenuItem{}, err
	}
	arg.Available = true
	item, err := s.store.CreateMenuItem(ctx, arg)
	if err != nil {
		return db.MenuItem{}, fmt.Errorf("CreateMenuItem service: %w", err)
	}
	return item, nil
}

func (s *menuService) UpdateMenuItem(ctx context.Context, arg db.UpdateMenuItemParams) (db.MenuItem, error) {
	if err := validateMenuItemId(arg.ID); err != nil {
		return db.MenuItem{}, err
	}
	if err := validateCategoryItemId(arg.CategoryID); err != nil {
		return db.MenuItem{}, err
	}
	if err := validateMenuItemName(arg.Name); err != nil {
		return db.MenuItem{}, err
	}
	if err := validatePrice(arg.Price); err != nil {
		return db.MenuItem{}, err
	}
	item, err := s.store.UpdateMenuItem(ctx, arg)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return db.MenuItem{}, ErrMenuItemNotFound

		}
		return db.MenuItem{}, fmt.Errorf("UpdateMenuItem service: %w", err)

	}
	return item, nil
}
func (s *menuService) DeleteMenuItem(ctx context.Context, id int32) error {
	if err := validateMenuItemId(id); err != nil {
		return err
	}

	_, err := s.store.GetMenuItemByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrMenuItemNotFound
		}
		return fmt.Errorf("DeleteMenuItem service get item: %w", err)
	}

	if err := s.store.DeleteMenuItem(ctx, id); err != nil {
		return fmt.Errorf("DeleteMenuItem service: %w", err)
	}
	return nil
}
