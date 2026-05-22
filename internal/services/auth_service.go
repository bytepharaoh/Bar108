package services

import (
	"bar108/internal/apperror"
	"bar108/internal/db"
	"bar108/internal/jwt"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

// authStore — private interface, only what auth needs.
type authStore interface {
	CreateUser(ctx context.Context, arg db.CreateUserParams) (db.User, error)
	GetUserByEmailForAuth(ctx context.Context, email string) (db.GetUserByEmailForAuthRow, error)
}

// RegisterInput — what the client sends to register.
type RegisterInput struct {
	Name     string
	Phone    string
	Email    string
	Password string // plain text — we hash it here
}

// LoginInput — what the client sends to login.
type LoginInput struct {
	Email    string
	Password string
}

// AuthResult — what we return after successful register/login.
type AuthResult struct {
	Token string
	User  db.User
}

// AuthService — exported interface for the handler.
type AuthService interface {
	Register(ctx context.Context, input RegisterInput) (AuthResult, error)
	Login(ctx context.Context, input LoginInput) (AuthResult, error)
}

type authService struct {
	store      authStore
	jwtManager *jwt.Manager
}

func NewAuthService(store authStore, jwtManager *jwt.Manager) AuthService {
	return &authService{
		store:      store,
		jwtManager: jwtManager,
	}
}
func (s *authService) Register(ctx context.Context, input RegisterInput) (AuthResult, error) {
	// Validate inputs
	if strings.TrimSpace(input.Name) == "" {
		return AuthResult{}, apperror.ErrInvalidInput
	}
	if strings.TrimSpace(input.Email) == "" {
		return AuthResult{}, apperror.ErrInvalidInput
	}
	if strings.TrimSpace(input.Phone) == "" {
		return AuthResult{}, apperror.ErrInvalidInput
	}
	if len(input.Password) < 6 {
		return AuthResult{}, apperror.New(400, "password must be at least 6 characters")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), 12)
	if err != nil {
		return AuthResult{}, fmt.Errorf("Register hash password: %w", err)
	}

	user, err := s.store.CreateUser(ctx, db.CreateUserParams{
		Name:         input.Name,
		Phone:        input.Phone,
		Email:        input.Email,
		PasswordHash: string(hash),
	})
	if err != nil {
		// Map only UNIQUE violations to 409, keep other DB failures as 500.
		var pgErr *pq.Error
		if errors.As(err, &pgErr) && string(pgErr.Code) == "23505" {
			return AuthResult{}, apperror.ErrAlreadyExists
		}
		return AuthResult{}, fmt.Errorf("Register create user: %w", err)
	}

	token, err := s.jwtManager.Generate(user.ID, user.Role)
	if err != nil {
		return AuthResult{}, fmt.Errorf("Register generate token: %w", err)
	}

	return AuthResult{Token: token, User: user}, nil
}

func (s *authService) Login(ctx context.Context, input LoginInput) (AuthResult, error) {
	if strings.TrimSpace(input.Email) == "" {
		return AuthResult{}, apperror.ErrInvalidInput
	}
	if strings.TrimSpace(input.Password) == "" {
		return AuthResult{}, apperror.ErrInvalidInput
	}
	// Fetch user — we use a special query that also returns password_hash
	authUser, err := s.store.GetUserByEmailForAuth(ctx, input.Email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// IMPORTANT: return the same error for wrong email AND wrong password.
			// If we returned "email not found" vs "wrong password" separately,
			// attackers could enumerate valid emails in our system.
			return AuthResult{}, apperror.New(401, "invalid credentials")
		}
		return AuthResult{}, fmt.Errorf("Login get user: %w", err)
	}
	// Check if account is active
	if !authUser.IsActive {
		return AuthResult{}, apperror.New(401, "account is deactivated")
	}
	if err := bcrypt.CompareHashAndPassword(
		[]byte(authUser.PasswordHash),
		[]byte(input.Password),
	); err != nil {
		// Same vague error — never tell the client which field was wrong
		return AuthResult{}, apperror.New(401, "invalid credentials")
	}
	// Password correct — generate token
	token, err := s.jwtManager.Generate(authUser.ID, authUser.Role)
	if err != nil {
		return AuthResult{}, fmt.Errorf("Login generate token: %w", err)
	}
	// Fetch full user for response
	// GetUserByEmailForAuth only returns a subset of fields
	user := db.User{
		ID:    authUser.ID,
		Email: authUser.Email,
		Role:  authUser.Role,
	}

	return AuthResult{Token: token, User: user}, nil
}
