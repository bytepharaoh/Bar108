package handlers_test

import (
	"bar108/internal/apperror"
	"bar108/internal/db"
	"bar108/internal/handlers"
	"bar108/internal/services/mocks"
	"errors"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"go.uber.org/mock/gomock"
)

func setupUserHandler(t *testing.T) (*gin.Engine, *mocks.MockUserService) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	mockSvc := mocks.NewMockUserService(ctrl)
	h := handlers.NewUserHandler(mockSvc)

	r := gin.New()
	r.Use(func(c *gin.Context) {

		c.Set("user_id", int32(1))
		c.Set("role", "admin")
		c.Next()
	})

	r.GET("/users", h.GetAllUsers)
	r.GET("/users/active", h.GetActiveUsers)
	r.GET("/users/:id", h.GetUserByID)
	r.POST("/users", h.CreateUser)
	r.PUT("/users/:id", h.UpdateUser)
	r.PATCH("/users/:id/bonus", h.UpdateUserBonusPoints)
	r.PATCH("/users/:id/activate", h.ActivateUser)
	r.PATCH("/users/:id/deactivate", h.DeactivateUser)

	return r, mockSvc
}

// =============================================
// GET /users
// =============================================

func TestUserHandler_GetAllUsers(t *testing.T) {
	tests := []struct {
		name       string
		setupMock  func(m *mocks.MockUserService)
		wantStatus int
	}{
		{
			name: "happy path — returns users",
			setupMock: func(m *mocks.MockUserService) {
				m.EXPECT().
					GetAllUsers(gomock.Any()).
					Return([]db.User{{ID: 1, Name: "Ahmed"}}, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "empty list — still 200",
			setupMock: func(m *mocks.MockUserService) {
				m.EXPECT().
					GetAllUsers(gomock.Any()).
					Return([]db.User{}, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "service error — 500",
			setupMock: func(m *mocks.MockUserService) {
				m.EXPECT().
					GetAllUsers(gomock.Any()).
					Return(nil, errors.New("db error"))
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, mockSvc := setupUserHandler(t)
			tt.setupMock(mockSvc)

			w := doRequest(r, http.MethodGet, "/users", nil)

			if w.Code != tt.wantStatus {
				t.Errorf("got status %d, want %d — body: %s",
					w.Code, tt.wantStatus, w.Body.String())
			}
		})
	}
}

// =============================================
// GET /users/:id
// =============================================

func TestUserHandler_GetUserByID(t *testing.T) {

	tests := []struct {
		name       string
		url        string
		setupMock  func(m *mocks.MockUserService)
		wantStatus int
	}{
		{
			name: "happy path",
			url:  "/users/1",
			setupMock: func(m *mocks.MockUserService) {
				m.EXPECT().
					GetUserByID(gomock.Any(), int32(1)).
					Return(db.User{
						ID:   1,
						Name: "Ahmed",
					}, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "not found — 404",
			url:  "/users/999",
			setupMock: func(m *mocks.MockUserService) {
				m.EXPECT().
					GetUserByID(gomock.Any(), int32(999)).
					Return(db.User{}, apperror.ErrUserNotFound)
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "invalid id string — 400",
			url:        "/users/abc",
			setupMock:  func(m *mocks.MockUserService) {},
			wantStatus: http.StatusBadRequest,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, mockSvc := setupUserHandler(t)
			tt.setupMock(mockSvc)
			w := doRequest(r, http.MethodGet, tt.url, nil)
			if w.Code != tt.wantStatus {
				t.Errorf(
					"got status %d, want %d — body: %s",
					w.Code,
					tt.wantStatus,
					w.Body.String(),
				)
			}
		})
	}

}

// =============================================
// POST /users
// =============================================

func TestUserHandler_CreateUser(t *testing.T) {
	tests := []struct {
		name       string
		body       interface{}
		setupMock  func(m *mocks.MockUserService)
		wantStatus int
	}{
		{
			name: "happy path — 201",
			body: map[string]interface{}{
				"name":          "Ahmed",
				"phone":         "+79001234567",
				"email":         "ahmed@bar108.com",
				"password_hash": "hashed123",
			},
			setupMock: func(m *mocks.MockUserService) {
				m.EXPECT().
					CreateUser(gomock.Any(), gomock.Any()).
					Return(db.User{ID: 1, Name: "Ahmed"}, nil)
			},
			wantStatus: http.StatusCreated,
		},
		{
			name:       "missing required fields — 400",
			body:       map[string]interface{}{"name": "Ahmed"},
			setupMock:  func(m *mocks.MockUserService) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "empty body — 400",
			body:       nil,
			setupMock:  func(m *mocks.MockUserService) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "service validation error — 400",
			body: map[string]interface{}{
				"name":          "Ahmed",
				"phone":         "+79001234567",
				"email":         "notanemail",
				"password_hash": "hashed123",
			},
			setupMock: func(m *mocks.MockUserService) {
				m.EXPECT().
					CreateUser(gomock.Any(), gomock.Any()).
					Return(db.User{}, apperror.ErrInvalidInput)
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "database error — 500",
			body: map[string]interface{}{
				"name":          "Ahmed",
				"phone":         "+79001234567",
				"email":         "ahmed@bar108.com",
				"password_hash": "hashed123",
			},
			setupMock: func(m *mocks.MockUserService) {
				m.EXPECT().
					CreateUser(gomock.Any(), gomock.Any()).
					Return(db.User{}, errors.New("duplicate key"))
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, mockSvc := setupUserHandler(t)
			tt.setupMock(mockSvc)

			w := doRequest(r, http.MethodPost, "/users", tt.body)

			if w.Code != tt.wantStatus {
				t.Errorf("got status %d, want %d — body: %s",
					w.Code, tt.wantStatus, w.Body.String())
			}
		})
	}
}

// =============================================
// PATCH /users/:id/activate
// =============================================

func TestUserHandler_ActivateUser(t *testing.T) {
	tests := []struct {
		name       string
		url        string
		setupMock  func(m *mocks.MockUserService)
		wantStatus int
	}{
		{
			name: "happy path — 200",
			url:  "/users/1/activate",
			setupMock: func(m *mocks.MockUserService) {
				m.EXPECT().
					ActivateUser(gomock.Any(), int32(1)).
					Return(db.User{ID: 1, IsActive: true}, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "already active — 409",
			url:  "/users/1/activate",
			setupMock: func(m *mocks.MockUserService) {
				m.EXPECT().
					ActivateUser(gomock.Any(), int32(1)).
					Return(db.User{}, apperror.ErrAlreadyActive)
			},
			wantStatus: http.StatusConflict,
		},
		{
			name: "not found — 404",
			url:  "/users/999/activate",
			setupMock: func(m *mocks.MockUserService) {
				m.EXPECT().
					ActivateUser(gomock.Any(), int32(999)).
					Return(db.User{}, apperror.ErrUserNotFound)
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "invalid id — 400",
			url:        "/users/abc/activate",
			setupMock:  func(m *mocks.MockUserService) {},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, mockSvc := setupUserHandler(t)
			tt.setupMock(mockSvc)

			w := doRequest(r, http.MethodPatch, tt.url, nil)

			if w.Code != tt.wantStatus {
				t.Errorf("got status %d, want %d — body: %s",
					w.Code, tt.wantStatus, w.Body.String())
			}
		})
	}
}

// =============================================
// PATCH /users/:id/deactivate
// =============================================

func TestUserHandler_DeactivateUser(t *testing.T) {
	tests := []struct {
		name       string
		url        string
		setupMock  func(m *mocks.MockUserService)
		wantStatus int
	}{
		{
			name: "happy path — 200",
			url:  "/users/1/deactivate",
			setupMock: func(m *mocks.MockUserService) {
				m.EXPECT().
					DeactivateUser(gomock.Any(), int32(1)).
					Return(db.User{ID: 1, IsActive: false}, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "already inactive — 409",
			url:  "/users/1/deactivate",
			setupMock: func(m *mocks.MockUserService) {
				m.EXPECT().
					DeactivateUser(gomock.Any(), int32(1)).
					Return(db.User{}, apperror.ErrAlreadyInactive)
			},
			wantStatus: http.StatusConflict,
		},
		{
			name: "has active orders — 409",
			url:  "/users/1/deactivate",
			setupMock: func(m *mocks.MockUserService) {
				m.EXPECT().
					DeactivateUser(gomock.Any(), int32(1)).
					Return(db.User{}, apperror.ErrHasActiveOrders)
			},
			wantStatus: http.StatusConflict,
		},
		{
			name: "not found — 404",
			url:  "/users/999/deactivate",
			setupMock: func(m *mocks.MockUserService) {
				m.EXPECT().
					DeactivateUser(gomock.Any(), int32(999)).
					Return(db.User{}, apperror.ErrUserNotFound)
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "invalid id — 400",
			url:        "/users/abc/deactivate",
			setupMock:  func(m *mocks.MockUserService) {},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, mockSvc := setupUserHandler(t)
			tt.setupMock(mockSvc)

			w := doRequest(r, http.MethodPatch, tt.url, nil)

			if w.Code != tt.wantStatus {
				t.Errorf("got status %d, want %d — body: %s",
					w.Code, tt.wantStatus, w.Body.String())
			}
		})
	}
}

// =============================================
// PATCH /users/:id/bonus
// =============================================

func TestUserHandler_UpdateUserBonusPoints(t *testing.T) {
	tests := []struct {
		name       string
		url        string
		body       interface{}
		setupMock  func(m *mocks.MockUserService)
		wantStatus int
	}{
		{
			name: "happy path — 200",
			url:  "/users/1/bonus",
			body: map[string]interface{}{"bonus_points": 100},
			setupMock: func(m *mocks.MockUserService) {
				m.EXPECT().
					UpdateUserBonusPoints(gomock.Any(), int32(1), int32(100)).
					Return(db.User{ID: 1, BonusPoints: 100}, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "negative points — service returns 400",
			url:  "/users/1/bonus",
			body: map[string]interface{}{"bonus_points": -50},
			setupMock: func(m *mocks.MockUserService) {
				m.EXPECT().
					UpdateUserBonusPoints(gomock.Any(), int32(1), int32(-50)).
					Return(db.User{}, apperror.ErrInvalidInput)
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "missing body — 400",
			url:        "/users/1/bonus",
			body:       nil,
			setupMock:  func(m *mocks.MockUserService) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid id — 400",
			url:        "/users/abc/bonus",
			body:       map[string]interface{}{"bonus_points": 100},
			setupMock:  func(m *mocks.MockUserService) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "user not found — 404",
			url:  "/users/999/bonus",
			body: map[string]interface{}{"bonus_points": 100},
			setupMock: func(m *mocks.MockUserService) {
				m.EXPECT().
					UpdateUserBonusPoints(gomock.Any(), int32(999), int32(100)).
					Return(db.User{}, apperror.ErrUserNotFound)
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, mockSvc := setupUserHandler(t)
			tt.setupMock(mockSvc)

			w := doRequest(r, http.MethodPatch, tt.url, tt.body)

			if w.Code != tt.wantStatus {
				t.Errorf("got status %d, want %d — body: %s",
					w.Code, tt.wantStatus, w.Body.String())
			}
		})
	}
}
