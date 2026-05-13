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

// setupMenuService creates a fresh mock + service for each test.
// We pass ctrl so gomock can verify expectations automatically
// via t.Cleanup() — no need for defer ctrl.Finish().
func setupMenuService(t *testing.T) (services.MenuService, *mocks.MockmenuStore) {
	t.Helper() // marks this as a helper — errors point to the caller, not here
	ctrl := gomock.NewController(t)
	mockStore := mocks.NewMockmenuStore(ctrl)
	svc := services.NewMenuService(mockStore)
	return svc, mockStore
}

// =============================================
// GetAllMenuItems
// =============================================

func TestMenuService_GetAllMenuItems(t *testing.T) {
	tests := []struct {
		name      string
		setupMock func(m *mocks.MockmenuStore)
		wantItems []db.GetAllMenuItemsRow
		wantErr   error
	}{
		{
			name: "happy path — returns all items",
			setupMock: func(m *mocks.MockmenuStore) {
				m.EXPECT().
					GetAllMenuItems(gomock.Any()).
					Return([]db.GetAllMenuItemsRow{
						{ID: 1, Name: "Classic Burger"},
						{ID: 2, Name: "Coke"},
					}, nil)
			},
			wantItems: []db.GetAllMenuItemsRow{
				{ID: 1, Name: "Classic Burger"},
				{ID: 2, Name: "Coke"},
			},
			wantErr: nil,
		},
		{
			name: "empty menu — valid, restaurant has no items yet",
			setupMock: func(m *mocks.MockmenuStore) {
				m.EXPECT().
					GetAllMenuItems(gomock.Any()).
					Return([]db.GetAllMenuItemsRow{}, nil)
			},
			wantItems: []db.GetAllMenuItemsRow{},
			wantErr:   nil,
		},
		{
			name: "database error — connection refused",
			setupMock: func(m *mocks.MockmenuStore) {
				m.EXPECT().
					GetAllMenuItems(gomock.Any()).
					Return(nil, errors.New("connection refused"))
			},
			wantItems: nil,
			wantErr:   errors.New("connection refused"), // any non-nil error
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, mockStore := setupMenuService(t)
			tt.setupMock(mockStore)

			items, err := svc.GetAllMenuItems(context.Background())

			if tt.wantErr != nil {
				if err == nil {
					t.Errorf("expected error but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if len(items) != len(tt.wantItems) {
					t.Errorf("got %d items, want %d", len(items), len(tt.wantItems))
				}
			}
		})
	}
}

// =============================================
// GetMenuItemByID
// =============================================

func TestMenuService_GetMenuItemByID(t *testing.T) {
	tests := []struct {
		name      string
		id        int32
		setupMock func(m *mocks.MockmenuStore)
		wantItem  db.GetMenuItemByIDRow
		wantErr   error
	}{
		{
			name: "happy path — item found",
			id:   1,
			setupMock: func(m *mocks.MockmenuStore) {
				m.EXPECT().
					GetMenuItemByID(gomock.Any(), int32(1)).
					Return(db.GetMenuItemByIDRow{ID: 1, Name: "Classic Burger"}, nil)
			},
			wantItem: db.GetMenuItemByIDRow{ID: 1, Name: "Classic Burger"},
			wantErr:  nil,
		},
		{
			name: "not found — item doesn't exist",
			id:   999,
			setupMock: func(m *mocks.MockmenuStore) {
				m.EXPECT().
					GetMenuItemByID(gomock.Any(), int32(999)).
					Return(db.GetMenuItemByIDRow{}, apperror.ErrMenuNotFound)
			},
			wantItem: db.GetMenuItemByIDRow{},
			wantErr:  apperror.ErrMenuNotFound,
		},
		{
			// Invalid ID — mock should NEVER be called.
			// The service validates input BEFORE hitting the store.
			// gomock enforces this — if GetMenuItemByID is called,
			// the test fails automatically because we set no EXPECT.
			name:      "invalid id — zero",
			id:        0,
			setupMock: func(m *mocks.MockmenuStore) {}, // no expectations
			wantErr:   apperror.ErrInvalidID,
		},
		{
			name:      "invalid id — negative",
			id:        -5,
			setupMock: func(m *mocks.MockmenuStore) {},
			wantErr:   apperror.ErrInvalidID,
		},
		{
			name: "database error — unexpected failure",
			id:   1,
			setupMock: func(m *mocks.MockmenuStore) {
				m.EXPECT().
					GetMenuItemByID(gomock.Any(), int32(1)).
					Return(db.GetMenuItemByIDRow{}, errors.New("timeout"))
			},
			wantErr: errors.New("timeout"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, mockStore := setupMenuService(t)
			tt.setupMock(mockStore)

			item, err := svc.GetMenuItemByID(context.Background(), tt.id)

			if tt.wantErr != nil {
				if err == nil {
					t.Fatalf("expected error %v but got nil", tt.wantErr)
				}
				// Check it's the right AppError when expected
				var appErr *apperror.AppError
				var wantAppErr *apperror.AppError
				if errors.As(tt.wantErr, &wantAppErr) {
					if !errors.As(err, &appErr) {
						t.Errorf("expected AppError but got %T: %v", err, err)
					} else if appErr.Code != wantAppErr.Code {
						t.Errorf("got HTTP code %d, want %d", appErr.Code, wantAppErr.Code)
					}
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if item.ID != tt.wantItem.ID {
					t.Errorf("got item ID %d, want %d", item.ID, tt.wantItem.ID)
				}
			}
		})
	}
}

// =============================================
// CreateMenuItem
// =============================================

func TestMenuService_CreateMenuItem(t *testing.T) {
	tests := []struct {
		name      string
		arg       db.CreateMenuItemParams
		setupMock func(m *mocks.MockmenuStore)
		wantErr   error
	}{
		{
			name: "happy path",
			arg: db.CreateMenuItemParams{
				CategoryID: 1,
				Name:       "Classic Burger",
				Price:      "350.00",
			},
			setupMock: func(m *mocks.MockmenuStore) {
				m.EXPECT().
					CreateMenuItem(gomock.Any(), gomock.Any()).
					Return(db.MenuItem{ID: 1, Name: "Classic Burger"}, nil)
			},
			wantErr: nil,
		},
		{
			// Empty name — store must never be called
			name: "empty name",
			arg: db.CreateMenuItemParams{
				CategoryID: 1,
				Name:       "",
				Price:      "350.00",
			},
			setupMock: func(m *mocks.MockmenuStore) {},
			wantErr:   apperror.ErrInvalidInput,
		},
		{
			// Whitespace name — trimmed to empty, same result
			name: "whitespace name",
			arg: db.CreateMenuItemParams{
				CategoryID: 1,
				Name:       "   ",
				Price:      "350.00",
			},
			setupMock: func(m *mocks.MockmenuStore) {},
			wantErr:   apperror.ErrInvalidInput,
		},
		{
			name: "invalid category id — zero",
			arg: db.CreateMenuItemParams{
				CategoryID: 0,
				Name:       "Classic Burger",
				Price:      "350.00",
			},
			setupMock: func(m *mocks.MockmenuStore) {},
			wantErr:   apperror.ErrInvalidID,
		},
		{
			name: "zero price",
			arg: db.CreateMenuItemParams{
				CategoryID: 1,
				Name:       "Classic Burger",
				Price:      "0.00",
			},
			setupMock: func(m *mocks.MockmenuStore) {},
			wantErr:   apperror.ErrZeroPrice,
		},
		{
			name: "negative price",
			arg: db.CreateMenuItemParams{
				CategoryID: 1,
				Name:       "Classic Burger",
				Price:      "-10.00",
			},
			setupMock: func(m *mocks.MockmenuStore) {},
			wantErr:   apperror.ErrNegativePrice,
		},
		{
			name: "database error",
			arg: db.CreateMenuItemParams{
				CategoryID: 1,
				Name:       "Classic Burger",
				Price:      "350.00",
			},
			setupMock: func(m *mocks.MockmenuStore) {
				m.EXPECT().
					CreateMenuItem(gomock.Any(), gomock.Any()).
					Return(db.MenuItem{}, errors.New("unique constraint violated"))
			},
			wantErr: errors.New("db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, mockStore := setupMenuService(t)
			tt.setupMock(mockStore)

			_, err := svc.CreateMenuItem(context.Background(), tt.arg)

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
// DeleteMenuItem
// =============================================

func TestMenuService_DeleteMenuItem(t *testing.T) {
	tests := []struct {
		name      string
		id        int32
		setupMock func(m *mocks.MockmenuStore)
		wantErr   error
	}{
		{
			name: "happy path",
			id:   1,
			setupMock: func(m *mocks.MockmenuStore) {
				// Delete first checks existence, then deletes
				m.EXPECT().
					GetMenuItemByID(gomock.Any(), int32(1)).
					Return(db.GetMenuItemByIDRow{ID: 1}, nil)
				m.EXPECT().
					DeleteMenuItem(gomock.Any(), int32(1)).
					Return(nil)
			},
			wantErr: nil,
		},
		{
			// Item doesn't exist — Delete should never be called.
			// gomock enforces this automatically.
			name: "item not found",
			id:   999,
			setupMock: func(m *mocks.MockmenuStore) {
				m.EXPECT().
					GetMenuItemByID(gomock.Any(), int32(999)).
					Return(db.GetMenuItemByIDRow{}, apperror.ErrMenuNotFound)
				// No EXPECT for DeleteMenuItem — gomock fails if it's called
			},
			wantErr: apperror.ErrMenuNotFound,
		},
		{
			name:      "invalid id",
			id:        0,
			setupMock: func(m *mocks.MockmenuStore) {},
			wantErr:   apperror.ErrInvalidID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, mockStore := setupMenuService(t)
			tt.setupMock(mockStore)

			err := svc.DeleteMenuItem(context.Background(), tt.id)

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
