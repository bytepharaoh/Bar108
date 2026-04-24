package repository

import (
	"bar108/internal/db"
	"context"
	"database/sql"
	"errors"
	"fmt"
)

type UsersRepository interface {
	ActivateUser(ctx context.Context, id int32) (db.User, error)
	CreateUser(ctx context.Context, arg db.CreateUserParams) (db.User, error)
	DeactivateUser(ctx context.Context, id int32) (db.User, error)
	GetActiveUsers(ctx context.Context) ([]db.User, error)
	GetAllUsers(ctx context.Context) ([]db.User, error)
	GetUserByEmail(ctx context.Context, email string) (db.User, error)
	GetUserByID(ctx context.Context, id int32) (db.User, error)
	GetUserByPhone(ctx context.Context, phone string) (db.User, error)
	HasActiveOrdersByUserID(ctx context.Context, userID int32) (bool, error)
	UpdateUser(ctx context.Context, arg db.UpdateUserParams) (db.User, error)
	UpdateUserBonusPoints(ctx context.Context, arg db.UpdateUserBonusPointsParams) (db.User, error)
}
type usersRepository struct {
	queries *db.Queries
}

func NewUserRepository(conn *sql.DB) UsersRepository {
	return &usersRepository{
		queries: db.New(conn),
	}
}
func (r *usersRepository) CreateUser(ctx context.Context, arg db.CreateUserParams) (db.User, error) {
	user, err := r.queries.CreateUser(ctx, arg)
	if err != nil {
		return db.User{}, fmt.Errorf("CreateUser: %w", err)
	}
	return user, nil
}

func (r *usersRepository) GetAllUsers(ctx context.Context) ([]db.User, error) {
	users, err := r.queries.GetAllUsers(ctx)
	if err != nil {
		return nil, fmt.Errorf("GetAllUsers: %w", err)
	}
	return users, nil
}
func (r *usersRepository) GetUserByID(ctx context.Context, id int32) (db.User, error) {
	user, err := r.queries.GetUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return db.User{}, fmt.Errorf("user with id %d not found: %w", id, sql.ErrNoRows)
		}
		return db.User{}, fmt.Errorf("GetUserByID: %w", err)

	}
	return user, nil
}
func (r *usersRepository) GetUserByEmail(ctx context.Context, email string) (db.User, error) {
	user, err := r.queries.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return db.User{}, fmt.Errorf("user with email %s not found: %w", email, sql.ErrNoRows)
		}
		return db.User{}, fmt.Errorf("GetUserByEmail: %w", err)

	}
	return user, nil
}
func (r *usersRepository) GetUserByPhone(ctx context.Context, phone string) (db.User, error) {
	user, err := r.queries.GetUserByPhone(ctx, phone)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return db.User{}, fmt.Errorf("user with phone %s not found: %w", phone, sql.ErrNoRows)
		}
		return db.User{}, fmt.Errorf("GetUserByPhone: %w", err)

	}
	return user, nil
}
func (r *usersRepository) UpdateUser(ctx context.Context, arg db.UpdateUserParams) (db.User, error) {
	user, err := r.queries.UpdateUser(ctx, arg)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return db.User{}, fmt.Errorf("user with id %d not found: %w", arg.ID, sql.ErrNoRows)

		}
		return db.User{}, fmt.Errorf("UpdateUser: %w", err)

	}
	return user, nil
}
func (r *usersRepository) ActivateUser(ctx context.Context, id int32) (db.User, error) {
	user, err := r.queries.ActivateUser(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return db.User{}, fmt.Errorf("user with id %d not found: %w", id, sql.ErrNoRows)

		}
		return db.User{}, fmt.Errorf("ActivateUser: %w", err)
	}
	return user, nil
}
func (r *usersRepository) GetActiveUsers(ctx context.Context) ([]db.User, error) {
	users, err := r.queries.GetActiveUsers(ctx)
	if err != nil {
		return nil, fmt.Errorf("GetActiveUsers: %w ", err)
	}
	return users, nil
}
func (r *usersRepository) DeactivateUser(ctx context.Context, id int32) (db.User, error) {
	user, err := r.queries.DeactivateUser(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return db.User{}, fmt.Errorf("user with id %d not found: %w", id, sql.ErrNoRows)

		}
		return db.User{}, fmt.Errorf("DeactivateUser: %w", err)
	}
	return user, nil
}
func (r *usersRepository) HasActiveOrdersByUserID(ctx context.Context, userID int32) (bool, error) {
	element, err := r.queries.HasActiveOrdersByUserID(ctx, userID)
	if err != nil {
		return false, fmt.Errorf("HasActiveOrdersByUserID: %w", err)
	}
	return element, nil
}
func (r *usersRepository) UpdateUserBonusPoints(ctx context.Context, arg db.UpdateUserBonusPointsParams) (db.User, error) {
	element, err := r.queries.UpdateUserBonusPoints(ctx, arg)
	if err != nil {
		return db.User{}, fmt.Errorf("UpdateUserBonusPoints: %w", err)
	}
	return element, nil
}
