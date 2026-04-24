package mocks

import (
	"bar108/internal/db"
	"context"

	"github.com/stretchr/testify/mock"
)

type MockMenuRepository struct {
	mock.Mock
}

func (m *MockMenuRepository) GetAllMenuItems(ctx context.Context) ([]db.GetAllMenuItemsRow, error) {
	args := m.Called(ctx)
	return args.Get(0).([]db.GetAllMenuItemsRow), args.Error(1)

}
func (m *MockMenuRepository) GetMenuItemsByID(ctx context.Context, id int32) (db.GetMenuItemByIDRow, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(db.GetMenuItemByIDRow), args.Error(1)
}
func (m *MockMenuRepository) GetAllCategories(ctx context.Context) ([]db.Category, error) {
	args := m.Called(ctx)
	return args.Get(0).([]db.Category), args.Error(1)
}

func (m *MockMenuRepository) CreateMenuItem(ctx context.Context, arg db.CreateMenuItemParams) (db.MenuItem, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(db.MenuItem), args.Error(1)
}
func (m *MockMenuRepository) UpdateMenuItem(ctx context.Context, arg db.UpdateMenuItemParams) (db.MenuItem, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(db.MenuItem), args.Error(1)
}
func (m *MockMenuRepository) DeleteMenuItem(ctx context.Context, id int32) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
