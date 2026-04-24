package services_test

import (
	"bar108/internal/db"
	"bar108/internal/repository/mocks"
	"bar108/internal/services"
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// helper — builds a fresh service + mock for each test
// We create a new mock per test so tests don't interfere
// with each other

func setupMenuService(t *testing.T) (services.MenuService, *mocks.MockMenuRepository) {
	mockRepo := new(mocks.MockMenuRepository)
	svc := services.NewMenuService(mockRepo)
	return svc, mockRepo
}
func TestGetAllMenuItems_HappyPath(t *testing.T) {
	svc, mockRepo := setupMenuService(t)
	// What we expect the repo to return

	expected := []db.GetAllMenuItemsRow{
		{ID: 1, Name: "Classic Burger"},
		{ID: 2, Name: "Coke"},
	}
	// Tell the mock: when GetAllMenuItems is called, return this
	mockRepo.On("GetAllMenuItems", context.Background()).Return(expected, nil)
	result, err := svc.GetAllMenuItems(context.Background())
	require.NoError(t, err)
	assert.Equal(t, expected, result)
	mockRepo.AssertExpectations(t)
}
func TestGetMenuItems_EmptyList(t *testing.T) {
	svc, mockRepo := setupMenuService(t)
	//empty slice - valid case -- resturant has no items yet
	mockRepo.On("GetAllMenuItems", context.Background()).Return([]db.GetAllMenuItemsRow{}, nil)
	result, err := svc.GetAllMenuItems(context.Background())
	require.NoError(t, err)
	assert.Empty(t, result)
	mockRepo.AssertExpectations(t)
}
func TestGetMenuItems_DB_Error(t *testing.T) {
	svc, mockRepo := setupMenuService(t)
	dbErr := errors.New("Connections refused")
	mockRepo.On("GetAllMenuItems", context.Background()).Return([]db.GetAllMenuItemsRow{}, dbErr)
	result, err := svc.GetAllMenuItems(context.Background())
	require.Error(t, err)
	assert.Empty(t, result)
	assert.ErrorContains(t, err, "GetAllMenuItems service")
	mockRepo.AssertExpectations(t)
}
func TestGetMeniItemByID_HappyPath(t *testing.T) {
	svc, mockRepo := setupMenuService(t)
	expected := db.GetMenuItemByIDRow{
		ID: 1, Name: "Classic Burger",
	}
	mockRepo.On("GetMenuItemByID", context.Background(), int32(1)).Return(expected, nil)
	result, err := svc.GetMenuItemByID(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, expected, result)
	mockRepo.AssertExpectations(t)
}
func TestGetMeniItemByID_NotFound(t *testing.T) {
	svc, mockRepo := setupMenuService(t)
	mockRepo.On("GetMenuItemByID", context.Background(), int32(999)).Return(db.GetMenuItemByIDRow{}, sql.ErrNoRows)

	_, err := svc.GetMenuItemByID(context.Background(), 999)
	require.Error(t, err)
	assert.ErrorIs(t, err, services.ErrMenuItemNotFound)
	mockRepo.AssertExpectations(t)
}
func TestGetMenuItemByID_InvalidID_Zero(t *testing.T) {
	svc, mockRepo := setupMenuService(t)
	// ID = 0 is invalid — repo should never be called
	_, err := svc.GetMenuItemByID(context.Background(), 0)

	require.Error(t, err)
	assert.ErrorIs(t, err, services.ErrInvalidMenuItemID)
	// AssertNotCalled proves the repo was never touched
	mockRepo.AssertNotCalled(t, "GetMenuItemByID")
}
func TestGetMenuItemByID_InvalidID_Negative(t *testing.T) {
	svc, mockRepo := setupMenuService(t)

	_, err := svc.GetMenuItemByID(context.Background(), -5)

	require.Error(t, err)
	assert.ErrorIs(t, err, services.ErrInvalidMenuItemID)
	mockRepo.AssertNotCalled(t, "GetMenuItemByID")
}
func TestGetMenuItemByID_DatabaseError(t *testing.T) {
	svc, mockRepo := setupMenuService(t)
	dbErr := errors.New("timeout")
	mockRepo.On("GetMenuItemByID", context.Background(), int32(1)).Return(db.GetMenuItemByIDRow{}, dbErr)

	_, err := svc.GetMenuItemByID(context.Background(), 1)
	require.Error(t, err)
	assert.False(t, errors.Is(err, services.ErrMenuItemNotFound))
	mockRepo.AssertExpectations(t)
}

// CreateMenuItems

func TestCreateMenuItem_HappyPath(t *testing.T) {
	svc, mockRepo := setupMenuService(t)
	arg := db.CreateMenuItemParams{
		CategoryID: 1,
		Name:       "Classic Burger",
		Price:      "9.99",
		Available:  true,
	}
	expected := db.MenuItem{ID: 1, Name: "Classic Burger"}
	mockRepo.On("CreateMenuItem", context.Background(), arg).Return(expected, nil)
	result, err := svc.CreateMenuItem(context.Background(), arg)
	require.NoError(t, err)
	assert.Equal(t, expected, result)
	mockRepo.AssertExpectations(t)
}
func TestCreateMenuItem_EmptyName(t *testing.T) {
	svc, mockRepo := setupMenuService(t)

	arg := db.CreateMenuItemParams{
		CategoryID: 1,
		Name:       "", // empty name
	}

	_, err := svc.CreateMenuItem(context.Background(), arg)

	require.Error(t, err)
	assert.ErrorIs(t, err, services.ErrEmptyMenuItemName)
	mockRepo.AssertNotCalled(t, "CreateMenuItem")
}
func TestCreateMenuItem_WhitespaceName(t *testing.T) {
	svc, mockRepo := setupMenuService(t)

	arg := db.CreateMenuItemParams{
		CategoryID: 1,
		Name:       "   ", // spaces only — should fail
	}

	_, err := svc.CreateMenuItem(context.Background(), arg)

	require.Error(t, err)
	assert.ErrorIs(t, err, services.ErrEmptyMenuItemName)
	mockRepo.AssertNotCalled(t, "CreateMenuItem")
}
func TestCreateMenuItem_InvalidCategoryID(t *testing.T) {
	svc, mockRepo := setupMenuService(t)

	arg := db.CreateMenuItemParams{
		CategoryID: 0, // invalid
		Name:       "Classic Burger",
	}

	_, err := svc.CreateMenuItem(context.Background(), arg)

	require.Error(t, err)
	assert.ErrorIs(t, err, services.ErrInvalidCategoryID)
	mockRepo.AssertNotCalled(t, "CreateMenuItem")
}
func TestCreateMenuItem_DatabaseError(t *testing.T) {
	svc, mockRepo := setupMenuService(t)

	arg := db.CreateMenuItemParams{
		CategoryID: 1,
		Name:       "Classic Burger",
		Price:      "9,99",
		Available:  true,
	}

	dbErr := errors.New("unique constraint violation")
	mockRepo.On("CreateMenuItem", context.Background(), arg).
		Return(db.MenuItem{}, dbErr)

	_, err := svc.CreateMenuItem(context.Background(), arg)

	require.Error(t, err)
	assert.ErrorContains(t, err, "CreateMenuItem service")
	mockRepo.AssertExpectations(t)
}

// DeleteMenuItem
func TestDeleteMenuItem_HappyPath(t *testing.T) {
	svc, mockRepo := setupMenuService(t)

	// First GetMenuItemByID is called to check existence
	mockRepo.On("GetMenuItemByID", context.Background(), int32(1)).
		Return(db.GetMenuItemByIDRow{ID: 1}, nil)

	// Then DeleteMenuItem is called
	mockRepo.On("DeleteMenuItem", context.Background(), int32(1)).
		Return(nil)

	err := svc.DeleteMenuItem(context.Background(), 1)

	require.NoError(t, err)
	mockRepo.AssertExpectations(t)
}
func TestDeleteMenuItem_NotFound(t *testing.T) {
	svc, mockRepo := setupMenuService(t)

	mockRepo.On("GetMenuItemByID", context.Background(), int32(999)).
		Return(db.GetMenuItemByIDRow{}, sql.ErrNoRows)

	err := svc.DeleteMenuItem(context.Background(), 999)

	require.Error(t, err)
	assert.ErrorIs(t, err, services.ErrMenuItemNotFound)
	// Delete should never be called if item doesn't exist
	mockRepo.AssertNotCalled(t, "DeleteMenuItem")
}
func TestDeleteMenuItem_InvalidID(t *testing.T) {
	svc, mockRepo := setupMenuService(t)

	err := svc.DeleteMenuItem(context.Background(), 0)

	require.Error(t, err)
	assert.ErrorIs(t, err, services.ErrInvalidMenuItemID)
	mockRepo.AssertNotCalled(t, "GetMenuItemByID")
	mockRepo.AssertNotCalled(t, "DeleteMenuItem")
}

// UpdateMenuItem
func TestUpdateMenuItem_HappyPath(t *testing.T) {
	svc, mockRepo := setupMenuService(t)

	arg := db.UpdateMenuItemParams{
		ID:         1,
		CategoryID: 1,
		Name:       "Updated Burger",
		Price:      "9.99",
	}
	expected := db.MenuItem{ID: 1, Name: "Updated Burger"}

	mockRepo.On("UpdateMenuItem", context.Background(), arg).
		Return(expected, nil)

	result, err := svc.UpdateMenuItem(context.Background(), arg)

	require.NoError(t, err)
	assert.Equal(t, expected, result)
	mockRepo.AssertExpectations(t)
}
func TestUpdateMenuItem_NotFound(t *testing.T) {
	svc, mockRepo := setupMenuService(t)

	arg := db.UpdateMenuItemParams{
		ID:         999,
		CategoryID: 1,
		Name:       "Ghost Burger",
		Price:      "9.99",
	}

	mockRepo.On("UpdateMenuItem", context.Background(), arg).
		Return(db.MenuItem{}, sql.ErrNoRows)

	_, err := svc.UpdateMenuItem(context.Background(), arg)

	require.Error(t, err)
	assert.ErrorIs(t, err, services.ErrMenuItemNotFound)
	mockRepo.AssertExpectations(t)
}
func TestUpdateMenuItem_InvalidID(t *testing.T) {
	svc, mockRepo := setupMenuService(t)

	arg := db.UpdateMenuItemParams{
		ID:         0,
		CategoryID: 1,
		Name:       "Some Burger",
	}

	_, err := svc.UpdateMenuItem(context.Background(), arg)

	require.Error(t, err)
	assert.ErrorIs(t, err, services.ErrInvalidMenuItemID)
	mockRepo.AssertNotCalled(t, "UpdateMenuItem")
}

func TestUpdateMenuItem_EmptyName(t *testing.T) {
	svc, mockRepo := setupMenuService(t)

	arg := db.UpdateMenuItemParams{
		ID:         1,
		CategoryID: 1,
		Name:       "",
	}

	_, err := svc.UpdateMenuItem(context.Background(), arg)

	require.Error(t, err)
	assert.ErrorIs(t, err, services.ErrEmptyMenuItemName)
	mockRepo.AssertNotCalled(t, "UpdateMenuItem")
}
