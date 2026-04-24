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
	ErrInvalidUserID       = errors.New("invalid user id")
	ErrUserNotFound        = errors.New("user not found")
	ErrEmptyUserName       = errors.New("username is required")
	ErrEmptyUserPhone      = errors.New("phone is required")
	ErrEmptyUserEmail      = errors.New("email is required")
	ErrInvalidUserEmail    = errors.New("invalid user email")
	ErrEmptyPasswordHash   = errors.New("password hash is required")
	ErrNegativeBonusPoints = errors.New("bonus points cannot be negative")
	ErrUserAlreadyInactive = errors.New("user is already inactive")
	ErrUserAlreadyActive   = errors.New("user is already active")
	ErrUserHasActiveOrders = errors.New("user has active orders")
)

type UserService interface {
	CreateUser(ctx context.Context, arg db.CreateUserParams) (db.User, error)
	GetAllUsers(ctx context.Context) ([]db.User, error)
	GetActiveUsers(ctx context.Context) ([]db.User, error)
	GetUserByID(ctx context.Context, id int32) (db.User, error)
	GetUserByEmail(ctx context.Context, email string) (db.User, error)
	GetUserByPhone(ctx context.Context, phone string) (db.User, error)
	UpdateUser(ctx context.Context, arg db.UpdateUserParams) (db.User, error)
	UpdateUserBonusPoints(ctx context.Context, id int32, bonusPoints int32) (db.User, error)
	DeactivateUser(ctx context.Context, id int32) (db.User, error)
	ActivateUser(ctx context.Context, id int32) (db.User, error)
}
type userService struct {
	repo repository.UsersRepository
}

func NewUserService(repo repository.UsersRepository) UserService {
	return &userService{
		repo: repo,
	}
}
func validateUserID(id int32) error {
	if id <= 0 {
		return ErrInvalidUserID
	}
	return nil
}
func validateRequiredUserFields(name, phone, email, passHash string) error {
	if strings.TrimSpace(name) == "" {
		return ErrEmptyUserName
	}
	if strings.TrimSpace(phone) == "" {
		return ErrEmptyUserPhone
	}

	if strings.TrimSpace(email) == "" {
		return ErrEmptyUserEmail
	}

	parts := strings.Split(email, "@")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" || !strings.Contains(parts[1], ".") {
		return ErrInvalidUserEmail
	}

	if strings.TrimSpace(passHash) == "" {
		return ErrEmptyPasswordHash
	}
	return nil
}
func (s *userService) GetUserByID(ctx context.Context, id int32) (db.User, error) {
	if err := validateUserID(id); err != nil {
		return db.User{}, err
	}
	user, err := s.repo.GetUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return db.User{}, ErrUserNotFound
		}
		return db.User{}, fmt.Errorf("GetUserByID service: %w", err)
	}
	return user, nil
}
func (s *userService) CreateUser(ctx context.Context, arg db.CreateUserParams) (db.User, error) {

	if err := validateRequiredUserFields(arg.Name, arg.Phone, arg.Email, arg.PasswordHash); err != nil {
		return db.User{}, err
	}
	// Business rule: new users always start with 0 bonus points

	arg.BonusPoints = 0
	user, err := s.repo.CreateUser(ctx, arg)
	if err != nil {
		return db.User{}, fmt.Errorf("CreateUser service: %w", err)
	}

	return user, nil

}
func (s *userService) GetAllUsers(ctx context.Context) ([]db.User, error) {
	users, err := s.repo.GetAllUsers(ctx)
	if err != nil {
		return nil, fmt.Errorf("GetAllUsers service: %w", err)
	}
	return users, nil
}
func (s *userService) GetActiveUsers(ctx context.Context) ([]db.User, error) {
	users, err := s.repo.GetActiveUsers(ctx)
	if err != nil {
		return nil, fmt.Errorf("GetActiveUsers service: %w", err)
	}
	return users, nil
}
func (s *userService) GetUserByEmail(ctx context.Context, email string) (db.User, error) {
	email = strings.TrimSpace(email)
	if email == "" {
		return db.User{}, ErrEmptyUserEmail
	}
	if !strings.Contains(email, "@") {
		return db.User{}, ErrInvalidUserEmail
	}
	users, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return db.User{}, ErrUserNotFound
		}
		return db.User{}, fmt.Errorf("GetUserByEmail: %w", err)
	}
	return users, nil
}
func (s *userService) GetUserByPhone(ctx context.Context, phone string) (db.User, error) {
	phone = strings.TrimSpace(phone)
	if phone == "" {
		return db.User{}, ErrEmptyUserPhone
	}
	user, err := s.repo.GetUserByPhone(ctx, phone)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return db.User{}, ErrUserNotFound
		}
		return db.User{}, fmt.Errorf("GetUserByPhone Service: %w", err)
	}
	return user, nil
}
func (s *userService) UpdateUser(ctx context.Context, arg db.UpdateUserParams) (db.User, error) {
	if err := validateUserID(arg.ID); err != nil {
		return db.User{}, err
	}
	if err := validateRequiredUserFields(arg.Name, arg.Phone, arg.Email, arg.PasswordHash); err != nil {
		return db.User{}, err
	}
	if arg.BonusPoints < 0 {
		return db.User{}, ErrNegativeBonusPoints
	}
	user, err := s.repo.UpdateUser(ctx, arg)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return db.User{}, ErrUserNotFound
		}
		return db.User{}, fmt.Errorf("UpdateUser Service: %w", err)
	}
	return user, nil

}
func (s *userService) UpdateUserBonusPoints(ctx context.Context, id int32, bonusPoints int32) (db.User, error) {
	if err := validateUserID(id); err != nil {
		return db.User{}, err
	}
	if bonusPoints < 0 {
		return db.User{}, ErrNegativeBonusPoints
	}
	arg := db.UpdateUserBonusPointsParams{
		ID:          id,
		BonusPoints: bonusPoints,
	}
	user, err := s.repo.UpdateUserBonusPoints(ctx, arg)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return db.User{}, ErrUserNotFound
		}
		return db.User{}, fmt.Errorf("UpdateUserBonusPoints Service: %w", err)
	}
	return user, nil
}
func (s *userService) ActivateUser(ctx context.Context, id int32) (db.User, error) {
	if err := validateUserID(id); err != nil {
		return db.User{}, err
	}
	user, err := s.repo.GetUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return db.User{}, ErrUserNotFound
		}
		return db.User{}, fmt.Errorf("ActivateUser service get user: %w", err)
	}
	if user.IsActive {
		return db.User{}, ErrUserAlreadyActive
	}
	activatedUser, err := s.repo.ActivateUser(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return db.User{}, ErrUserNotFound
		}
		return db.User{}, fmt.Errorf("ActivateUser service: %w", err)
	}
	return activatedUser, nil

}
func (s *userService) DeactivateUser(ctx context.Context, id int32) (db.User, error) {
	if err := validateUserID(id); err != nil {
		return db.User{}, err
	}
	user, err := s.repo.GetUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return db.User{}, ErrUserNotFound
		}
		return db.User{}, fmt.Errorf("DeactivateUser service get user: %w", err)
	}
	if !user.IsActive {
		return db.User{}, ErrUserAlreadyInactive
	}
	hasActiveOrders, err := s.repo.HasActiveOrdersByUserID(ctx, id)
	if err != nil {
		return db.User{}, fmt.Errorf("DeactivateUser service check active orders: %w", err)
	}
	if hasActiveOrders {
		return db.User{}, ErrUserHasActiveOrders
	}
	deactivatedUser, err := s.repo.DeactivateUser(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return db.User{}, ErrUserNotFound
		}
		return db.User{}, fmt.Errorf("DeactivateUser service: %w", err)
	}
	return deactivatedUser, nil
}
