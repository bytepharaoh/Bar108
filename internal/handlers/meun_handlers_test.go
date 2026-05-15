package handlers_test

import (
	"bar108/internal/apperror"
	"bar108/internal/db"
	"bar108/internal/handlers"
	"bar108/internal/services/mocks"
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"go.uber.org/mock/gomock"
)

// doRequest is a shared helper that fires an HTTP request
// against the test router and returns the recorded response.
// Using httptest.NewRecorder() means no real server is needed —
// everything runs in memory.
func doRequest(r *gin.Engine, method, url string, body interface{}) *httptest.ResponseRecorder {
	var buf bytes.Buffer
	if body != nil {
		_ = json.NewEncoder(&buf).Encode(body)
	}
	req := httptest.NewRequest(method, url, &buf)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

// setupMenuHandler creates a fresh mock + handler + router for each test.
// gin.TestMode suppresses debug output during tests.
func setupMenuHandler(t *testing.T) (*gin.Engine, *mocks.MockMenuService) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	mockSvc := mocks.NewMockMenuService(ctrl)
	h := handlers.NewMenuHandler(mockSvc)

	r := gin.New()
	r.GET("/menu", h.GetAllMenuItems)
	r.GET("/menu/:id", h.GetMenuItemByID)
	r.GET("/categories", h.GetAllCategories)
	r.POST("/menu", h.CreateMenuItem)
	r.PUT("/menu/:id", h.UpdateMenuItem)
	r.DELETE("/menu/:id", h.DeleteMenuItem)

	return r, mockSvc
}

// =============================================
// GET /menu
// =============================================

func TestMenuHandler_GetAllMenuItems(t *testing.T) {
	tests := []struct {
		name       string
		setupMock  func(m *mocks.MockMenuService)
		wantStatus int
	}{
		{
			name: "happy path — returns items",
			setupMock: func(m *mocks.MockMenuService) {
				m.EXPECT().
					GetAllMenuItems(gomock.Any()).
					Return([]db.GetAllMenuItemsRow{
						{ID: 1, Name: "Classic Burger"},
					}, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "empty menu — still 200",
			setupMock: func(m *mocks.MockMenuService) {
				m.EXPECT().
					GetAllMenuItems(gomock.Any()).
					Return([]db.GetAllMenuItemsRow{}, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "service error — 500",
			setupMock: func(m *mocks.MockMenuService) {
				m.EXPECT().
					GetAllMenuItems(gomock.Any()).
					Return(nil, errors.New("db down"))
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, mockSvc := setupMenuHandler(t)
			tt.setupMock(mockSvc)

			w := doRequest(r, http.MethodGet, "/menu", nil)

			if w.Code != tt.wantStatus {
				t.Errorf("got status %d, want %d — body: %s",
					w.Code, tt.wantStatus, w.Body.String())
			}
		})
	}
}

// =============================================
// GET /menu/:id
// =============================================

func TestMenuHandler_GetMenuItemByID(t *testing.T) {
	tests := []struct {
		name       string
		url        string
		setupMock  func(m *mocks.MockMenuService)
		wantStatus int
	}{
		{
			name: "happy path",
			url:  "/menu/1",
			setupMock: func(m *mocks.MockMenuService) {
				m.EXPECT().
					GetMenuItemByID(gomock.Any(), int32(1)).
					Return(db.GetMenuItemByIDRow{ID: 1, Name: "Classic Burger"}, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "not found — 404",
			url:  "/menu/999",
			setupMock: func(m *mocks.MockMenuService) {
				m.EXPECT().
					GetMenuItemByID(gomock.Any(), int32(999)).
					Return(db.GetMenuItemByIDRow{}, apperror.ErrMenuNotFound)
			},
			wantStatus: http.StatusNotFound,
		},
		{
			// String ID can't be parsed — parseID returns 400.
			// Service must never be called.
			// gomock enforces this: no EXPECT = fail if called.
			name:       "invalid id — string",
			url:        "/menu/abc",
			setupMock:  func(m *mocks.MockMenuService) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "invalid id — zero from service",
			url:  "/menu/0",
			setupMock: func(m *mocks.MockMenuService) {
				m.EXPECT().
					GetMenuItemByID(gomock.Any(), int32(0)).
					Return(db.GetMenuItemByIDRow{}, apperror.ErrInvalidID)
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "database error — 500",
			url:  "/menu/1",
			setupMock: func(m *mocks.MockMenuService) {
				m.EXPECT().
					GetMenuItemByID(gomock.Any(), int32(1)).
					Return(db.GetMenuItemByIDRow{}, errors.New("timeout"))
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, mockSvc := setupMenuHandler(t)
			tt.setupMock(mockSvc)

			w := doRequest(r, http.MethodGet, tt.url, nil)

			if w.Code != tt.wantStatus {
				t.Errorf("got status %d, want %d — body: %s",
					w.Code, tt.wantStatus, w.Body.String())
			}
		})
	}
}

// =============================================
// POST /menu
// =============================================

func TestMenuHandler_CreateMenuItem(t *testing.T) {
	tests := []struct {
		name       string
		body       interface{}
		setupMock  func(m *mocks.MockMenuService)
		wantStatus int
	}{
		{
			name: "happy path — 201 created",
			body: map[string]interface{}{
				"category_id": 1,
				"name":        "Classic Burger",
				"price":       "350.00",
			},
			setupMock: func(m *mocks.MockMenuService) {
				m.EXPECT().
					CreateMenuItem(gomock.Any(), gomock.Any()).
					Return(db.MenuItem{ID: 1, Name: "Classic Burger"}, nil)
			},
			wantStatus: http.StatusCreated,
		},
		{
			// Missing required fields — Gin binding returns 400.
			// Service never called.
			name:       "missing required fields — 400",
			body:       map[string]interface{}{"name": "Burger"},
			setupMock:  func(m *mocks.MockMenuService) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "empty body — 400",
			body:       nil,
			setupMock:  func(m *mocks.MockMenuService) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "zero price — service returns 400",
			body: map[string]interface{}{
				"category_id": 1,
				"name":        "Burger",
				"price":       "0.00",
			},
			setupMock: func(m *mocks.MockMenuService) {
				m.EXPECT().
					CreateMenuItem(gomock.Any(), gomock.Any()).
					Return(db.MenuItem{}, apperror.ErrZeroPrice)
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "database error — 500",
			body: map[string]interface{}{
				"category_id": 1,
				"name":        "Burger",
				"price":       "350.00",
			},
			setupMock: func(m *mocks.MockMenuService) {
				m.EXPECT().
					CreateMenuItem(gomock.Any(), gomock.Any()).
					Return(db.MenuItem{}, errors.New("constraint violated"))
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, mockSvc := setupMenuHandler(t)
			tt.setupMock(mockSvc)

			w := doRequest(r, http.MethodPost, "/menu", tt.body)

			if w.Code != tt.wantStatus {
				t.Errorf("got status %d, want %d — body: %s",
					w.Code, tt.wantStatus, w.Body.String())
			}
		})
	}
}

// =============================================
// DELETE /menu/:id
// =============================================

func TestMenuHandler_DeleteMenuItem(t *testing.T) {
	tests := []struct {
		name       string
		url        string
		setupMock  func(m *mocks.MockMenuService)
		wantStatus int
	}{
		{
			name: "happy path — 204 no content",
			url:  "/menu/1",
			setupMock: func(m *mocks.MockMenuService) {
				m.EXPECT().
					DeleteMenuItem(gomock.Any(), int32(1)).
					Return(nil)
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name: "not found — 404",
			url:  "/menu/999",
			setupMock: func(m *mocks.MockMenuService) {
				m.EXPECT().
					DeleteMenuItem(gomock.Any(), int32(999)).
					Return(apperror.ErrMenuNotFound)
			},
			wantStatus: http.StatusNotFound,
		},
		{
			// parseID rejects "abc" before service is ever called.
			name:       "invalid id string — 400",
			url:        "/menu/abc",
			setupMock:  func(m *mocks.MockMenuService) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "database error — 500",
			url:  "/menu/1",
			setupMock: func(m *mocks.MockMenuService) {
				m.EXPECT().
					DeleteMenuItem(gomock.Any(), int32(1)).
					Return(errors.New("db error"))
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, mockSvc := setupMenuHandler(t)
			tt.setupMock(mockSvc)

			w := doRequest(r, http.MethodDelete, tt.url, nil)

			if w.Code != tt.wantStatus {
				t.Errorf("got status %d, want %d — body: %s",
					w.Code, tt.wantStatus, w.Body.String())
			}
		})
	}
}

// =============================================
// PUT /menu/:id
// =============================================

func TestMenuHandler_UpdateMenuItem(t *testing.T) {
	tests := []struct {
		name       string
		url        string
		body       interface{}
		setupMock  func(m *mocks.MockMenuService)
		wantStatus int
	}{
		{
			name: "happy path — 200",
			url:  "/menu/1",
			body: map[string]interface{}{
				"category_id": 1,
				"name":        "Updated Burger",
				"price":       "400.00",
				"available":   true,
			},
			setupMock: func(m *mocks.MockMenuService) {
				m.EXPECT().
					UpdateMenuItem(gomock.Any(), gomock.Any()).
					Return(db.MenuItem{ID: 1, Name: "Updated Burger"}, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "invalid id string — 400",
			url:        "/menu/abc",
			body:       map[string]interface{}{"category_id": 1, "name": "Burger", "price": "350.00"},
			setupMock:  func(m *mocks.MockMenuService) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "missing body — 400",
			url:        "/menu/1",
			body:       nil,
			setupMock:  func(m *mocks.MockMenuService) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "not found — 404",
			url:  "/menu/999",
			body: map[string]interface{}{
				"category_id": 1,
				"name":        "Ghost Burger",
				"price":       "350.00",
			},
			setupMock: func(m *mocks.MockMenuService) {
				m.EXPECT().
					UpdateMenuItem(gomock.Any(), gomock.Any()).
					Return(db.MenuItem{}, apperror.ErrMenuNotFound)
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, mockSvc := setupMenuHandler(t)
			tt.setupMock(mockSvc)

			w := doRequest(r, http.MethodPut, tt.url, tt.body)

			if w.Code != tt.wantStatus {
				t.Errorf("got status %d, want %d — body: %s",
					w.Code, tt.wantStatus, w.Body.String())
			}
		})
	}
}
