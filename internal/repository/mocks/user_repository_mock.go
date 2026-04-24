package mocks

import (
	"bar108/internal/db"
	"context"

	"github.com/stretchr/testify/mock"
)

type MockUsersRepository struct {
	mock.Mock
}

func (m *MockUsersRepository) CreateUser(ctx context.Context, arg db.CreateUserParams) (db.User, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(db.User), args.Error(1)
}
func (m *MockUsersRepository) GetAllUsers(ctx context.Context) ([]db.User, error) {
	args := m.Called(ctx)
	return args.Get(0).([]db.User), args.Error(1)
}

func (m *MockUsersRepository) GetActiveUsers(ctx context.Context) ([]db.User, error) {
	args := m.Called(ctx)
	return args.Get(0).([]db.User), args.Error(1)
}

func (m *MockUsersRepository) GetUserByID(ctx context.Context, id int32) (db.User, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(db.User), args.Error(1)
}

func (m *MockUsersRepository) GetUserByEmail(ctx context.Context, email string) (db.User, error) {
	args := m.Called(ctx, email)
	return args.Get(0).(db.User), args.Error(1)
}

func (m *MockUsersRepository) GetUserByPhone(ctx context.Context, phone string) (db.User, error) {
	args := m.Called(ctx, phone)
	return args.Get(0).(db.User), args.Error(1)
}

func (m *MockUsersRepository) UpdateUser(ctx context.Context, arg db.UpdateUserParams) (db.User, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(db.User), args.Error(1)
}

func (m *MockUsersRepository) UpdateUserBonusPoints(ctx context.Context, arg db.UpdateUserBonusPointsParams) (db.User, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(db.User), args.Error(1)
}

func (m *MockUsersRepository) ActivateUser(ctx context.Context, id int32) (db.User, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(db.User), args.Error(1)
}

func (m *MockUsersRepository) DeactivateUser(ctx context.Context, id int32) (db.User, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(db.User), args.Error(1)
}
func (m *MockUsersRepository) HasActiveOrdersByUserID(ctx context.Context, Userid int32) (bool, error) {
	args := m.Called(ctx, Userid)
	return args.Get(0).(bool), args.Error(1)
}
