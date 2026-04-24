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

func setupUserService(t *testing.T) (services.UserService, *mocks.MockUsersRepository) {
	mockRepo := new(mocks.MockUsersRepository)
	svc := services.NewUserService(mockRepo)
	return svc, mockRepo
}

// CreateUser

func TestCreateUser_HappyPath(t *testing.T) {
	svc, mockRepo := setupUserService(t)

	arg := db.CreateUserParams{
		Name:         "Ahmed",
		Phone:        "+79001234567",
		Email:        "ahmed@bar108.com",
		PasswordHash: "hashedpassword123",
		BonusPoints:  0,
	}

	expected := db.User{
		ID:    1,
		Name:  "Ahmed",
		Email: "ahmed@bar108.com",
	}

	mockRepo.On("CreateUser", context.Background(), arg).
		Return(expected, nil)

	result, err := svc.CreateUser(context.Background(), arg)

	require.NoError(t, err)
	assert.Equal(t, expected, result)
	mockRepo.AssertExpectations(t)
}

func TestCreateUser_BonusPointsAlwaysZero(t *testing.T) {
	svc, mockRepo := setupUserService(t)

	// Even if caller passes 500 bonus points,
	// the service must reset it to 0
	arg := db.CreateUserParams{
		Name:         "Ahmed",
		Phone:        "+79001234567",
		Email:        "ahmed@bar108.com",
		PasswordHash: "hashedpassword123",
		BonusPoints:  500, // caller tries to cheat
	}

	// The service should normalize BonusPoints to 0
	// before calling the repo — so we expect 0 here
	expectedArg := db.CreateUserParams{
		Name:         "Ahmed",
		Phone:        "+79001234567",
		Email:        "ahmed@bar108.com",
		PasswordHash: "hashedpassword123",
		BonusPoints:  0, // must be 0 regardless
	}

	expected := db.User{ID: 1, Name: "Ahmed"}

	mockRepo.On("CreateUser", context.Background(), expectedArg).
		Return(expected, nil)

	result, err := svc.CreateUser(context.Background(), arg)

	require.NoError(t, err)
	assert.Equal(t, expected, result)
	mockRepo.AssertExpectations(t)
}

func TestCreateUser_EmptyName(t *testing.T) {
	svc, mockRepo := setupUserService(t)

	arg := db.CreateUserParams{
		Name:         "",
		Phone:        "+79001234567",
		Email:        "ahmed@bar108.com",
		PasswordHash: "hashedpassword123",
	}

	_, err := svc.CreateUser(context.Background(), arg)

	require.Error(t, err)
	assert.ErrorIs(t, err, services.ErrEmptyUserName)
	mockRepo.AssertNotCalled(t, "CreateUser")
}

func TestCreateUser_WhitespaceName(t *testing.T) {
	svc, mockRepo := setupUserService(t)

	arg := db.CreateUserParams{
		Name:         "   ",
		Phone:        "+79001234567",
		Email:        "ahmed@bar108.com",
		PasswordHash: "hashedpassword123",
	}

	_, err := svc.CreateUser(context.Background(), arg)

	require.Error(t, err)
	assert.ErrorIs(t, err, services.ErrEmptyUserName)
	mockRepo.AssertNotCalled(t, "CreateUser")
}

func TestCreateUser_EmptyPhone(t *testing.T) {
	svc, mockRepo := setupUserService(t)

	arg := db.CreateUserParams{
		Name:         "Ahmed",
		Phone:        "",
		Email:        "ahmed@bar108.com",
		PasswordHash: "hashedpassword123",
	}

	_, err := svc.CreateUser(context.Background(), arg)

	require.Error(t, err)
	assert.ErrorIs(t, err, services.ErrEmptyUserPhone)
	mockRepo.AssertNotCalled(t, "CreateUser")
}

func TestCreateUser_EmptyEmail(t *testing.T) {
	svc, mockRepo := setupUserService(t)

	arg := db.CreateUserParams{
		Name:         "Ahmed",
		Phone:        "+79001234567",
		Email:        "",
		PasswordHash: "hashedpassword123",
	}

	_, err := svc.CreateUser(context.Background(), arg)

	require.Error(t, err)
	assert.ErrorIs(t, err, services.ErrEmptyUserEmail)
	mockRepo.AssertNotCalled(t, "CreateUser")
}

func TestCreateUser_InvalidEmail_NoAt(t *testing.T) {
	svc, mockRepo := setupUserService(t)

	arg := db.CreateUserParams{
		Name:         "Ahmed",
		Phone:        "+79001234567",
		Email:        "ahmedbar108.com", // missing @
		PasswordHash: "hashedpassword123",
	}

	_, err := svc.CreateUser(context.Background(), arg)

	require.Error(t, err)
	assert.ErrorIs(t, err, services.ErrInvalidUserEmail)
	mockRepo.AssertNotCalled(t, "CreateUser")
}

func TestCreateUser_InvalidEmail_OnlyAt(t *testing.T) {
	svc, mockRepo := setupUserService(t)

	arg := db.CreateUserParams{
		Name:         "Ahmed",
		Phone:        "+79001234567",
		Email:        "@",
		PasswordHash: "hashedpassword123",
	}

	_, err := svc.CreateUser(context.Background(), arg)

	require.Error(t, err)
	assert.ErrorIs(t, err, services.ErrInvalidUserEmail)
	mockRepo.AssertNotCalled(t, "CreateUser")
}

func TestCreateUser_InvalidEmail_NoDomain(t *testing.T) {
	svc, mockRepo := setupUserService(t)

	arg := db.CreateUserParams{
		Name:         "Ahmed",
		Phone:        "+79001234567",
		Email:        "ahmed@",
		PasswordHash: "hashedpassword123",
	}

	_, err := svc.CreateUser(context.Background(), arg)

	require.Error(t, err)
	assert.ErrorIs(t, err, services.ErrInvalidUserEmail)
	mockRepo.AssertNotCalled(t, "CreateUser")
}

func TestCreateUser_EmptyPasswordHash(t *testing.T) {
	svc, mockRepo := setupUserService(t)

	arg := db.CreateUserParams{
		Name:         "Ahmed",
		Phone:        "+79001234567",
		Email:        "ahmed@bar108.com",
		PasswordHash: "",
	}

	_, err := svc.CreateUser(context.Background(), arg)

	require.Error(t, err)
	assert.ErrorIs(t, err, services.ErrEmptyPasswordHash)
	mockRepo.AssertNotCalled(t, "CreateUser")
}

func TestCreateUser_DatabaseError(t *testing.T) {
	svc, mockRepo := setupUserService(t)

	arg := db.CreateUserParams{
		Name:         "Ahmed",
		Phone:        "+79001234567",
		Email:        "ahmed@bar108.com",
		PasswordHash: "hashedpassword123",
		BonusPoints:  0,
	}

	dbErr := errors.New("duplicate key value violates unique constraint")
	mockRepo.On("CreateUser", context.Background(), arg).
		Return(db.User{}, dbErr)

	_, err := svc.CreateUser(context.Background(), arg)

	require.Error(t, err)
	assert.ErrorContains(t, err, "CreateUser service")
	mockRepo.AssertExpectations(t)
}

// =============================================
// GetUserByID
// =============================================

func TestGetUserByID_HappyPath(t *testing.T) {
	svc, mockRepo := setupUserService(t)

	expected := db.User{ID: 1, Name: "Ahmed"}

	mockRepo.On("GetUserByID", context.Background(), int32(1)).
		Return(expected, nil)

	result, err := svc.GetUserByID(context.Background(), 1)

	require.NoError(t, err)
	assert.Equal(t, expected, result)
	mockRepo.AssertExpectations(t)
}

func TestGetUserByID_NotFound(t *testing.T) {
	svc, mockRepo := setupUserService(t)

	mockRepo.On("GetUserByID", context.Background(), int32(999)).
		Return(db.User{}, sql.ErrNoRows)

	_, err := svc.GetUserByID(context.Background(), 999)

	require.Error(t, err)
	assert.ErrorIs(t, err, services.ErrUserNotFound)
	mockRepo.AssertExpectations(t)
}

func TestGetUserByID_InvalidID_Zero(t *testing.T) {
	svc, mockRepo := setupUserService(t)

	_, err := svc.GetUserByID(context.Background(), 0)

	require.Error(t, err)
	assert.ErrorIs(t, err, services.ErrInvalidUserID)
	mockRepo.AssertNotCalled(t, "GetUserByID")
}

func TestGetUserByID_InvalidID_Negative(t *testing.T) {
	svc, mockRepo := setupUserService(t)

	_, err := svc.GetUserByID(context.Background(), -10)

	require.Error(t, err)
	assert.ErrorIs(t, err, services.ErrInvalidUserID)
	mockRepo.AssertNotCalled(t, "GetUserByID")
}

// GetUserByEmail

func TestGetUserByEmail_HappyPath(t *testing.T) {
	svc, mockRepo := setupUserService(t)

	expected := db.User{ID: 1, Email: "ahmed@bar108.com"}

	mockRepo.On("GetUserByEmail", context.Background(), "ahmed@bar108.com").
		Return(expected, nil)

	result, err := svc.GetUserByEmail(context.Background(), "ahmed@bar108.com")

	require.NoError(t, err)
	assert.Equal(t, expected, result)
	mockRepo.AssertExpectations(t)
}

func TestGetUserByEmail_NotFound(t *testing.T) {
	svc, mockRepo := setupUserService(t)

	mockRepo.On("GetUserByEmail", context.Background(), "ghost@bar108.com").
		Return(db.User{}, sql.ErrNoRows)

	_, err := svc.GetUserByEmail(context.Background(), "ghost@bar108.com")

	require.Error(t, err)
	assert.ErrorIs(t, err, services.ErrUserNotFound)
	mockRepo.AssertExpectations(t)
}

func TestGetUserByEmail_EmptyEmail(t *testing.T) {
	svc, mockRepo := setupUserService(t)

	_, err := svc.GetUserByEmail(context.Background(), "")

	require.Error(t, err)
	assert.ErrorIs(t, err, services.ErrEmptyUserEmail)
	mockRepo.AssertNotCalled(t, "GetUserByEmail")
}

func TestGetUserByEmail_InvalidEmail(t *testing.T) {
	svc, mockRepo := setupUserService(t)

	_, err := svc.GetUserByEmail(context.Background(), "notanemail")

	require.Error(t, err)
	assert.ErrorIs(t, err, services.ErrInvalidUserEmail)
	mockRepo.AssertNotCalled(t, "GetUserByEmail")
}

// GetUserByPhone

func TestGetUserByPhone_HappyPath(t *testing.T) {
	svc, mockRepo := setupUserService(t)

	expected := db.User{ID: 1, Phone: "+79001234567"}

	mockRepo.On("GetUserByPhone", context.Background(), "+79001234567").
		Return(expected, nil)

	result, err := svc.GetUserByPhone(context.Background(), "+79001234567")

	require.NoError(t, err)
	assert.Equal(t, expected, result)
	mockRepo.AssertExpectations(t)
}

func TestGetUserByPhone_NotFound(t *testing.T) {
	svc, mockRepo := setupUserService(t)

	mockRepo.On("GetUserByPhone", context.Background(), "+70000000000").
		Return(db.User{}, sql.ErrNoRows)

	_, err := svc.GetUserByPhone(context.Background(), "+70000000000")

	require.Error(t, err)
	assert.ErrorIs(t, err, services.ErrUserNotFound)
	mockRepo.AssertExpectations(t)
}

func TestGetUserByPhone_EmptyPhone(t *testing.T) {
	svc, mockRepo := setupUserService(t)

	_, err := svc.GetUserByPhone(context.Background(), "")

	require.Error(t, err)
	assert.ErrorIs(t, err, services.ErrEmptyUserPhone)
	mockRepo.AssertNotCalled(t, "GetUserByPhone")
}

// UpdateUserBonusPoints

func TestUpdateUserBonusPoints_HappyPath(t *testing.T) {
	svc, mockRepo := setupUserService(t)

	arg := db.UpdateUserBonusPointsParams{
		ID:          1,
		BonusPoints: 100,
	}
	expected := db.User{ID: 1, BonusPoints: 100}

	mockRepo.On("UpdateUserBonusPoints", context.Background(), arg).
		Return(expected, nil)

	result, err := svc.UpdateUserBonusPoints(context.Background(), 1, 100)

	require.NoError(t, err)
	assert.Equal(t, expected, result)
	mockRepo.AssertExpectations(t)
}

func TestUpdateUserBonusPoints_NegativePoints(t *testing.T) {
	svc, mockRepo := setupUserService(t)

	_, err := svc.UpdateUserBonusPoints(context.Background(), 1, -50)

	require.Error(t, err)
	assert.ErrorIs(t, err, services.ErrNegativeBonusPoints)
	mockRepo.AssertNotCalled(t, "UpdateUserBonusPoints")
}

func TestUpdateUserBonusPoints_InvalidID(t *testing.T) {
	svc, mockRepo := setupUserService(t)

	_, err := svc.UpdateUserBonusPoints(context.Background(), 0, 100)

	require.Error(t, err)
	assert.ErrorIs(t, err, services.ErrInvalidUserID)
	mockRepo.AssertNotCalled(t, "UpdateUserBonusPoints")
}

func TestUpdateUserBonusPoints_UserNotFound(t *testing.T) {
	svc, mockRepo := setupUserService(t)

	arg := db.UpdateUserBonusPointsParams{
		ID:          999,
		BonusPoints: 100,
	}

	mockRepo.On("UpdateUserBonusPoints", context.Background(), arg).
		Return(db.User{}, sql.ErrNoRows)

	_, err := svc.UpdateUserBonusPoints(context.Background(), 999, 100)

	require.Error(t, err)
	assert.ErrorIs(t, err, services.ErrUserNotFound)
	mockRepo.AssertExpectations(t)
}

// ActivateUser

func TestActivateUser_HappyPath(t *testing.T) {
	svc, mockRepo := setupUserService(t)

	// User is currently inactive
	inactiveUser := db.User{ID: 1, IsActive: false}
	activatedUser := db.User{ID: 1, IsActive: true}

	mockRepo.On("GetUserByID", context.Background(), int32(1)).
		Return(inactiveUser, nil)
	mockRepo.On("ActivateUser", context.Background(), int32(1)).
		Return(activatedUser, nil)

	result, err := svc.ActivateUser(context.Background(), 1)

	require.NoError(t, err)
	assert.True(t, result.IsActive)
	mockRepo.AssertExpectations(t)
}

func TestActivateUser_AlreadyActive(t *testing.T) {
	svc, mockRepo := setupUserService(t)

	// User is already active — should return error
	activeUser := db.User{ID: 1, IsActive: true}

	mockRepo.On("GetUserByID", context.Background(), int32(1)).
		Return(activeUser, nil)

	_, err := svc.ActivateUser(context.Background(), 1)

	require.Error(t, err)
	assert.ErrorIs(t, err, services.ErrUserAlreadyActive)
	// ActivateUser on repo should never be called
	mockRepo.AssertNotCalled(t, "ActivateUser")
}

func TestActivateUser_UserNotFound(t *testing.T) {
	svc, mockRepo := setupUserService(t)

	mockRepo.On("GetUserByID", context.Background(), int32(999)).
		Return(db.User{}, sql.ErrNoRows)

	_, err := svc.ActivateUser(context.Background(), 999)

	require.Error(t, err)
	assert.ErrorIs(t, err, services.ErrUserNotFound)
	mockRepo.AssertNotCalled(t, "ActivateUser")
}

func TestActivateUser_InvalidID(t *testing.T) {
	svc, mockRepo := setupUserService(t)

	_, err := svc.ActivateUser(context.Background(), 0)

	require.Error(t, err)
	assert.ErrorIs(t, err, services.ErrInvalidUserID)
	mockRepo.AssertNotCalled(t, "GetUserByID")
	mockRepo.AssertNotCalled(t, "ActivateUser")
}

// DeactivateUser

func TestDeactivateUser_HappyPath(t *testing.T) {
	svc, mockRepo := setupUserService(t)

	activeUser := db.User{ID: 1, IsActive: true}
	deactivatedUser := db.User{ID: 1, IsActive: false}

	mockRepo.On("GetUserByID", context.Background(), int32(1)).
		Return(activeUser, nil)
	// No active orders
	mockRepo.On("HasActiveOrdersByUserID", context.Background(), int32(1)).
		Return(false, nil)
	mockRepo.On("DeactivateUser", context.Background(), int32(1)).
		Return(deactivatedUser, nil)

	result, err := svc.DeactivateUser(context.Background(), 1)

	require.NoError(t, err)
	assert.False(t, result.IsActive)
	mockRepo.AssertExpectations(t)
}

func TestDeactivateUser_AlreadyInactive(t *testing.T) {
	svc, mockRepo := setupUserService(t)

	inactiveUser := db.User{ID: 1, IsActive: false}

	mockRepo.On("GetUserByID", context.Background(), int32(1)).
		Return(inactiveUser, nil)

	_, err := svc.DeactivateUser(context.Background(), 1)

	require.Error(t, err)
	assert.ErrorIs(t, err, services.ErrUserAlreadyInactive)
	mockRepo.AssertNotCalled(t, "HasActiveOrdersByUserID")
	mockRepo.AssertNotCalled(t, "DeactivateUser")
}

func TestDeactivateUser_HasActiveOrders(t *testing.T) {
	svc, mockRepo := setupUserService(t)

	activeUser := db.User{ID: 1, IsActive: true}

	mockRepo.On("GetUserByID", context.Background(), int32(1)).
		Return(activeUser, nil)
	// User has active orders — cannot deactivate
	mockRepo.On("HasActiveOrdersByUserID", context.Background(), int32(1)).
		Return(true, nil)

	_, err := svc.DeactivateUser(context.Background(), 1)

	require.Error(t, err)
	assert.ErrorIs(t, err, services.ErrUserHasActiveOrders)
	// DeactivateUser on repo should never be called
	mockRepo.AssertNotCalled(t, "DeactivateUser")
}

func TestDeactivateUser_UserNotFound(t *testing.T) {
	svc, mockRepo := setupUserService(t)

	mockRepo.On("GetUserByID", context.Background(), int32(999)).
		Return(db.User{}, sql.ErrNoRows)

	_, err := svc.DeactivateUser(context.Background(), 999)

	require.Error(t, err)
	assert.ErrorIs(t, err, services.ErrUserNotFound)
	mockRepo.AssertNotCalled(t, "HasActiveOrdersByUserID")
	mockRepo.AssertNotCalled(t, "DeactivateUser")
}

func TestDeactivateUser_InvalidID(t *testing.T) {
	svc, mockRepo := setupUserService(t)

	_, err := svc.DeactivateUser(context.Background(), -1)

	require.Error(t, err)
	assert.ErrorIs(t, err, services.ErrInvalidUserID)
	mockRepo.AssertNotCalled(t, "GetUserByID")
	mockRepo.AssertNotCalled(t, "HasActiveOrdersByUserID")
	mockRepo.AssertNotCalled(t, "DeactivateUser")
}

// GetAllUsers

func TestGetAllUsers_HappyPath(t *testing.T) {
	svc, mockRepo := setupUserService(t)

	expected := []db.User{
		{ID: 1, Name: "Ahmed"},
		{ID: 2, Name: "Ali"},
	}

	mockRepo.On("GetAllUsers", context.Background()).
		Return(expected, nil)

	result, err := svc.GetAllUsers(context.Background())

	require.NoError(t, err)
	assert.Equal(t, expected, result)
	mockRepo.AssertExpectations(t)
}

func TestGetAllUsers_EmptyList(t *testing.T) {
	svc, mockRepo := setupUserService(t)

	mockRepo.On("GetAllUsers", context.Background()).
		Return([]db.User{}, nil)

	result, err := svc.GetAllUsers(context.Background())

	require.NoError(t, err)
	assert.Empty(t, result)
	mockRepo.AssertExpectations(t)
}

func TestGetAllUsers_DatabaseError(t *testing.T) {
	svc, mockRepo := setupUserService(t)

	mockRepo.On("GetAllUsers", context.Background()).
		Return([]db.User{}, errors.New("db error"))

	result, err := svc.GetAllUsers(context.Background())

	require.Error(t, err)
	assert.Nil(t, result)
	mockRepo.AssertExpectations(t)
}

// GetActiveUsers

func TestGetActiveUsers_HappyPath(t *testing.T) {
	svc, mockRepo := setupUserService(t)

	expected := []db.User{
		{ID: 1, Name: "Ahmed", IsActive: true},
	}

	mockRepo.On("GetActiveUsers", context.Background()).
		Return(expected, nil)

	result, err := svc.GetActiveUsers(context.Background())

	require.NoError(t, err)
	assert.Equal(t, expected, result)
	// Extra safety — all returned users must be active
	for _, u := range result {
		assert.True(t, u.IsActive)
	}
	mockRepo.AssertExpectations(t)
}

func TestGetActiveUsers_EmptyList(t *testing.T) {
	svc, mockRepo := setupUserService(t)

	mockRepo.On("GetActiveUsers", context.Background()).
		Return([]db.User{}, nil)

	result, err := svc.GetActiveUsers(context.Background())

	require.NoError(t, err)
	assert.Empty(t, result)
	mockRepo.AssertExpectations(t)
}
