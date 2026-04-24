package services

import (
	"bar108/internal/db"
	"bar108/internal/repository"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

var (
	ErrMenuItemNotFound  = errors.New("menu item not found")
	ErrCategoryNotFound  = errors.New("category not found")
	ErrInvalidMenuItemID = errors.New("invalid menu item id")
	ErrInvalidCategoryID = errors.New("invalid category id")
	ErrEmptyMenuItemName = errors.New("menu item name is required")
	ErrNegativePrice     = errors.New("price cannot be negative")
	ErrZeroPrice         = errors.New("price must be greater than zero")
)

type MenuService interface {
	GetAllMenuItems(ctx context.Context) ([]db.GetAllMenuItemsRow, error)
	GetMenuItemByID(ctx context.Context, id int32) (db.GetMenuItemByIDRow, error)
	GetAllCategories(ctx context.Context) ([]db.Category, error)
	CreateMenuItem(ctx context.Context, arg db.CreateMenuItemParams) (db.MenuItem, error)
	UpdateMenuItem(ctx context.Context, arg db.UpdateMenuItemParams) (db.MenuItem, error)
	DeleteMenuItem(ctx context.Context, id int32) error
}
type menuService struct {
	repo repository.MenuRepository
}

func NewMenuService(repo repository.MenuRepository) MenuService {
	return &menuService{
		repo: repo,
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
	items, err := s.repo.GetAllMenuItems(ctx)
	if err != nil {
		return nil, fmt.Errorf("GetAllMenuItems service: %w", err)
	}
	return items, nil
}
func (s *menuService) GetMenuItemByID(ctx context.Context, id int32) (db.GetMenuItemByIDRow, error) {
	if err := validateMenuItemId(id); err != nil {
		return db.GetMenuItemByIDRow{}, err
	}
	item, err := s.repo.GetMenuItemByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return db.GetMenuItemByIDRow{}, ErrMenuItemNotFound
		}
		return db.GetMenuItemByIDRow{}, fmt.Errorf("GetMenuItemByID service: %w", err)
	}
	return item, nil

}
func (s *menuService) GetAllCategories(ctx context.Context) ([]db.Category, error) {
	categories, err := s.repo.GetAllCategories(ctx)
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
	item, err := s.repo.CreateMenuItem(ctx, arg)
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
	item, err := s.repo.UpdateMenuItem(ctx, arg)
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

	_, err := s.repo.GetMenuItemByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrMenuItemNotFound
		}
		return fmt.Errorf("DeleteMenuItem service get item: %w", err)
	}

	if err := s.repo.DeleteMenuItem(ctx, id); err != nil {
		return fmt.Errorf("DeleteMenuItem service: %w", err)
	}
	return nil
}
