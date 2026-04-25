package mocks

import (
	"bar108/internal/db"
	"context"

	"github.com/stretchr/testify/mock"
)

type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) GetAllUsers(ctx context.Context) ([]db.User, error) {
	args := m.Called(ctx)
	return args.Get(0).([]db.User), args.Error(1)
}

func (m *MockUserService) GetActiveUsers(ctx context.Context) ([]db.User, error) {
	args := m.Called(ctx)
	return args.Get(0).([]db.User), args.Error(1)
}

func (m *MockUserService) GetUserByID(ctx context.Context, id int32) (db.User, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(db.User), args.Error(1)
}

func (m *MockUserService) GetUserByEmail(ctx context.Context, email string) (db.User, error) {
	args := m.Called(ctx, email)
	return args.Get(0).(db.User), args.Error(1)
}

func (m *MockUserService) GetUserByPhone(ctx context.Context, phone string) (db.User, error) {
	args := m.Called(ctx, phone)
	return args.Get(0).(db.User), args.Error(1)
}

func (m *MockUserService) CreateUser(ctx context.Context, arg db.CreateUserParams) (db.User, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(db.User), args.Error(1)
}

func (m *MockUserService) UpdateUser(ctx context.Context, arg db.UpdateUserParams) (db.User, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(db.User), args.Error(1)
}

func (m *MockUserService) UpdateUserBonusPoints(ctx context.Context, id int32, points int32) (db.User, error) {
	args := m.Called(ctx, id, points)
	return args.Get(0).(db.User), args.Error(1)
}

func (m *MockUserService) ActivateUser(ctx context.Context, id int32) (db.User, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(db.User), args.Error(1)
}

func (m *MockUserService) DeactivateUser(ctx context.Context, id int32) (db.User, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(db.User), args.Error(1)
}
