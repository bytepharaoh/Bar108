package services_test

import (
	"bar108/internal/apperror"
	"bar108/internal/db"
	"bar108/internal/services"
	"bar108/internal/services/mocks"
	"context"
	"errors"
	"testing"

	"go.uber.org/mock/gomock"
)

func setupUserService(t *testing.T) (services.UserService, *mocks.MockuserStore) {
	t.Helper()
	ctrl := gomock.NewController(t)
	mockStore := mocks.NewMockuserStore(ctrl)
	svc := services.NewUserService(mockStore)
	return svc, mockStore
}

// =============================================
// CreateUser
// =============================================

func TestUserService_CreateUser(t *testing.T) {
	tests := []struct {
		name      string
		arg       db.CreateUserParams
		setupMock func(m *mocks.MockuserStore)
		wantErr   error
	}{
		{
			name: "happy path",
			arg: db.CreateUserParams{
				Name:         "Ahmed",
				Phone:        "+79001234567",
				Email:        "ahmed@bar108.com",
				PasswordHash: "hashed123",
			},
			setupMock: func(m *mocks.MockuserStore) {
				// BonusPoints must be 0 regardless of what caller sends
				// gomock.Any() here because service modifies the arg
				m.EXPECT().
					CreateUser(gomock.Any(), gomock.Any()).
					Return(db.User{ID: 1, Name: "Ahmed"}, nil)
			},
			wantErr: nil,
		},
		{
			// Business rule: bonus points always reset to 0 on creation.
			// Even if caller sends 500, service must pass 0 to the store.
			// We verify this with a custom matcher.
			name: "bonus points always forced to zero",
			arg: db.CreateUserParams{
				Name:         "Ahmed",
				Phone:        "+79001234567",
				Email:        "ahmed@bar108.com",
				PasswordHash: "hashed123",
				BonusPoints:  500, // caller tries to set this
			},
			setupMock: func(m *mocks.MockuserStore) {
				m.EXPECT().
					CreateUser(gomock.Any(), gomock.Cond(func(x any) bool {
						// Verify the service reset BonusPoints to 0
						params, ok := x.(db.CreateUserParams)
						return ok && params.BonusPoints == 0
					})).
					Return(db.User{ID: 1}, nil)
			},
			wantErr: nil,
		},
		{
			name: "empty name",
			arg: db.CreateUserParams{
				Name:         "",
				Phone:        "+79001234567",
				Email:        "ahmed@bar108.com",
				PasswordHash: "hashed123",
			},
			setupMock: func(m *mocks.MockuserStore) {},
			wantErr:   apperror.ErrInvalidInput,
		},
		{
			name: "whitespace name",
			arg: db.CreateUserParams{
				Name:         "   ",
				Phone:        "+79001234567",
				Email:        "ahmed@bar108.com",
				PasswordHash: "hashed123",
			},
			setupMock: func(m *mocks.MockuserStore) {},
			wantErr:   apperror.ErrInvalidInput,
		},
		{
			name: "empty phone",
			arg: db.CreateUserParams{
				Name:         "Ahmed",
				Phone:        "",
				Email:        "ahmed@bar108.com",
				PasswordHash: "hashed123",
			},
			setupMock: func(m *mocks.MockuserStore) {},
			wantErr:   apperror.ErrInvalidInput,
		},
		{
			name: "empty email",
			arg: db.CreateUserParams{
				Name:         "Ahmed",
				Phone:        "+79001234567",
				Email:        "",
				PasswordHash: "hashed123",
			},
			setupMock: func(m *mocks.MockuserStore) {},
			wantErr:   apperror.ErrInvalidInput,
		},
		{
			name: "invalid email — no at sign",
			arg: db.CreateUserParams{
				Name:         "Ahmed",
				Phone:        "+79001234567",
				Email:        "ahmedbar108.com",
				PasswordHash: "hashed123",
			},
			setupMock: func(m *mocks.MockuserStore) {},
			wantErr:   apperror.ErrInvalidInput,
		},
		{
			name: "invalid email — only at sign",
			arg: db.CreateUserParams{
				Name:         "Ahmed",
				Phone:        "+79001234567",
				Email:        "@",
				PasswordHash: "hashed123",
			},
			setupMock: func(m *mocks.MockuserStore) {},
			wantErr:   apperror.ErrInvalidInput,
		},
		{
			name: "invalid email — no domain",
			arg: db.CreateUserParams{
				Name:         "Ahmed",
				Phone:        "+79001234567",
				Email:        "ahmed@",
				PasswordHash: "hashed123",
			},
			setupMock: func(m *mocks.MockuserStore) {},
			wantErr:   apperror.ErrInvalidInput,
		},
		{
			name: "empty password hash",
			arg: db.CreateUserParams{
				Name:         "Ahmed",
				Phone:        "+79001234567",
				Email:        "ahmed@bar108.com",
				PasswordHash: "",
			},
			setupMock: func(m *mocks.MockuserStore) {},
			wantErr:   apperror.ErrInvalidInput,
		},
		{
			name: "database error — duplicate email",
			arg: db.CreateUserParams{
				Name:         "Ahmed",
				Phone:        "+79001234567",
				Email:        "ahmed@bar108.com",
				PasswordHash: "hashed123",
			},
			setupMock: func(m *mocks.MockuserStore) {
				m.EXPECT().
					CreateUser(gomock.Any(), gomock.Any()).
					Return(db.User{}, errors.New("duplicate key"))
			},
			wantErr: errors.New("db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, mockStore := setupUserService(t)
			tt.setupMock(mockStore)

			_, err := svc.CreateUser(context.Background(), tt.arg)

			if tt.wantErr != nil {
				if err == nil {
					t.Fatalf("expected error but got nil")
				}
				var appErr *apperror.AppError
				var wantAppErr *apperror.AppError
				if errors.As(tt.wantErr, &wantAppErr) && errors.As(err, &appErr) {
					if appErr.Code != wantAppErr.Code {
						t.Errorf("got HTTP code %d, want %d", appErr.Code, wantAppErr.Code)
					}
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

// =============================================
// GetUserByID
// =============================================

func TestUserService_GetUserByID(t *testing.T) {
	tests := []struct {
		name      string
		id        int32
		setupMock func(m *mocks.MockuserStore)
		wantUser  db.User
		wantErr   error
	}{
		{
			name: "happy path",
			id:   1,
			setupMock: func(m *mocks.MockuserStore) {
				m.EXPECT().
					GetUserByID(gomock.Any(), int32(1)).
					Return(db.User{ID: 1, Name: "Ahmed"}, nil)
			},
			wantUser: db.User{ID: 1, Name: "Ahmed"},
			wantErr:  nil,
		},
		{
			name: "not found",
			id:   999,
			setupMock: func(m *mocks.MockuserStore) {
				m.EXPECT().
					GetUserByID(gomock.Any(), int32(999)).
					Return(db.User{}, apperror.ErrUserNotFound)
			},
			wantErr: apperror.ErrUserNotFound,
		},
		{
			name:      "invalid id — zero",
			id:        0,
			setupMock: func(m *mocks.MockuserStore) {},
			wantErr:   apperror.ErrInvalidID,
		},
		{
			name:      "invalid id — negative",
			id:        -1,
			setupMock: func(m *mocks.MockuserStore) {},
			wantErr:   apperror.ErrInvalidID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, mockStore := setupUserService(t)
			tt.setupMock(mockStore)

			user, err := svc.GetUserByID(context.Background(), tt.id)

			if tt.wantErr != nil {
				if err == nil {
					t.Fatalf("expected error but got nil")
				}
				var appErr *apperror.AppError
				var wantAppErr *apperror.AppError
				if errors.As(tt.wantErr, &wantAppErr) && errors.As(err, &appErr) {
					if appErr.Code != wantAppErr.Code {
						t.Errorf("got HTTP code %d, want %d", appErr.Code, wantAppErr.Code)
					}
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if user.ID != tt.wantUser.ID {
					t.Errorf("got user ID %d, want %d", user.ID, tt.wantUser.ID)
				}
			}
		})
	}
}

// =============================================
// ActivateUser
// =============================================

func TestUserService_ActivateUser(t *testing.T) {
	tests := []struct {
		name      string
		id        int32
		setupMock func(m *mocks.MockuserStore)
		wantErr   error
	}{
		{
			name: "happy path — inactive user gets activated",
			id:   1,
			setupMock: func(m *mocks.MockuserStore) {
				// Service first fetches user to check current state
				m.EXPECT().
					GetUserByID(gomock.Any(), int32(1)).
					Return(db.User{ID: 1, IsActive: false}, nil)
				m.EXPECT().
					ActivateUser(gomock.Any(), int32(1)).
					Return(db.User{ID: 1, IsActive: true}, nil)
			},
			wantErr: nil,
		},
		{
			// Already active — ActivateUser on store must never be called.
			// gomock enforces this because we set no EXPECT for it.
			name: "already active — conflict",
			id:   1,
			setupMock: func(m *mocks.MockuserStore) {
				m.EXPECT().
					GetUserByID(gomock.Any(), int32(1)).
					Return(db.User{ID: 1, IsActive: true}, nil)
			},
			wantErr: apperror.ErrAlreadyActive,
		},
		{
			name: "user not found",
			id:   999,
			setupMock: func(m *mocks.MockuserStore) {
				m.EXPECT().
					GetUserByID(gomock.Any(), int32(999)).
					Return(db.User{}, apperror.ErrUserNotFound)
			},
			wantErr: apperror.ErrUserNotFound,
		},
		{
			name:      "invalid id",
			id:        0,
			setupMock: func(m *mocks.MockuserStore) {},
			wantErr:   apperror.ErrInvalidID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, mockStore := setupUserService(t)
			tt.setupMock(mockStore)

			_, err := svc.ActivateUser(context.Background(), tt.id)

			if tt.wantErr != nil {
				if err == nil {
					t.Fatalf("expected error but got nil")
				}
				var appErr *apperror.AppError
				var wantAppErr *apperror.AppError
				if errors.As(tt.wantErr, &wantAppErr) && errors.As(err, &appErr) {
					if appErr.Code != wantAppErr.Code {
						t.Errorf("got HTTP code %d, want %d", appErr.Code, wantAppErr.Code)
					}
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

// =============================================
// DeactivateUser
// =============================================

func TestUserService_DeactivateUser(t *testing.T) {
	tests := []struct {
		name      string
		id        int32
		setupMock func(m *mocks.MockuserStore)
		wantErr   error
	}{
		{
			name: "happy path — active user with no active orders",
			id:   1,
			setupMock: func(m *mocks.MockuserStore) {
				m.EXPECT().
					GetUserByID(gomock.Any(), int32(1)).
					Return(db.User{ID: 1, IsActive: true}, nil)
				m.EXPECT().
					HasActiveOrdersByUserID(gomock.Any(), int32(1)).
					Return(false, nil)
				m.EXPECT().
					DeactivateUser(gomock.Any(), int32(1)).
					Return(db.User{ID: 1, IsActive: false}, nil)
			},
			wantErr: nil,
		},
		{
			// Has active orders — must not be deactivated.
			// DeactivateUser on store must never be called.
			name: "has active orders — conflict",
			id:   1,
			setupMock: func(m *mocks.MockuserStore) {
				m.EXPECT().
					GetUserByID(gomock.Any(), int32(1)).
					Return(db.User{ID: 1, IsActive: true}, nil)
				m.EXPECT().
					HasActiveOrdersByUserID(gomock.Any(), int32(1)).
					Return(true, nil)
				// No EXPECT for DeactivateUser — gomock fails if called
			},
			wantErr: apperror.ErrHasActiveOrders,
		},
		{
			name: "already inactive — conflict",
			id:   1,
			setupMock: func(m *mocks.MockuserStore) {
				m.EXPECT().
					GetUserByID(gomock.Any(), int32(1)).
					Return(db.User{ID: 1, IsActive: false}, nil)
				// No EXPECT for HasActiveOrdersByUserID or DeactivateUser
			},
			wantErr: apperror.ErrAlreadyInactive,
		},
		{
			name: "user not found",
			id:   999,
			setupMock: func(m *mocks.MockuserStore) {
				m.EXPECT().
					GetUserByID(gomock.Any(), int32(999)).
					Return(db.User{}, apperror.ErrUserNotFound)
			},
			wantErr: apperror.ErrUserNotFound,
		},
		{
			name:      "invalid id — zero",
			id:        0,
			setupMock: func(m *mocks.MockuserStore) {},
			wantErr:   apperror.ErrInvalidID,
		},
		{
			name:      "invalid id — negative",
			id:        -3,
			setupMock: func(m *mocks.MockuserStore) {},
			wantErr:   apperror.ErrInvalidID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, mockStore := setupUserService(t)
			tt.setupMock(mockStore)

			_, err := svc.DeactivateUser(context.Background(), tt.id)

			if tt.wantErr != nil {
				if err == nil {
					t.Fatalf("expected error but got nil")
				}
				var appErr *apperror.AppError
				var wantAppErr *apperror.AppError
				if errors.As(tt.wantErr, &wantAppErr) && errors.As(err, &appErr) {
					if appErr.Code != wantAppErr.Code {
						t.Errorf("got HTTP code %d, want %d", appErr.Code, wantAppErr.Code)
					}
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

// =============================================
// UpdateUserBonusPoints
// =============================================

func TestUserService_UpdateUserBonusPoints(t *testing.T) {
	tests := []struct {
		name        string
		id          int32
		bonusPoints int32
		setupMock   func(m *mocks.MockuserStore)
		wantErr     error
	}{
		{
			name:        "happy path",
			id:          1,
			bonusPoints: 100,
			setupMock: func(m *mocks.MockuserStore) {
				m.EXPECT().
					UpdateUserBonusPoints(gomock.Any(), gomock.Any()).
					Return(db.User{ID: 1, BonusPoints: 100}, nil)
			},
			wantErr: nil,
		},
		{
			// Negative bonus points — business rule violation.
			// Store must never be called.
			name:        "negative bonus points",
			id:          1,
			bonusPoints: -50,
			setupMock:   func(m *mocks.MockuserStore) {},
			wantErr:     apperror.ErrInvalidInput,
		},
		{
			name:        "invalid id",
			id:          0,
			bonusPoints: 100,
			setupMock:   func(m *mocks.MockuserStore) {},
			wantErr:     apperror.ErrInvalidID,
		},
		{
			name:        "user not found",
			id:          999,
			bonusPoints: 100,
			setupMock: func(m *mocks.MockuserStore) {
				m.EXPECT().
					UpdateUserBonusPoints(gomock.Any(), gomock.Any()).
					Return(db.User{}, apperror.ErrUserNotFound)
			},
			wantErr: apperror.ErrUserNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, mockStore := setupUserService(t)
			tt.setupMock(mockStore)

			_, err := svc.UpdateUserBonusPoints(context.Background(), tt.id, tt.bonusPoints)

			if tt.wantErr != nil {
				if err == nil {
					t.Fatalf("expected error but got nil")
				}
				var appErr *apperror.AppError
				var wantAppErr *apperror.AppError
				if errors.As(tt.wantErr, &wantAppErr) && errors.As(err, &appErr) {
					if appErr.Code != wantAppErr.Code {
						t.Errorf("got HTTP code %d, want %d", appErr.Code, wantAppErr.Code)
					}
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}
