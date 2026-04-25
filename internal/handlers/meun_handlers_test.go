package handlers_test

import (
	"bar108/internal/db"
	"bar108/internal/handlers"
	"bar108/internal/services"
	servicemocks "bar108/internal/services/mock"
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func setupMenuHandler(t *testing.T) (*gin.Engine, *servicemocks.MockMenuService) {
	gin.SetMode(gin.TestMode)
	mockSvc := new(servicemocks.MockMenuService)
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

func doRequest(r *gin.Engine, method, path string, body interface{}) *httptest.ResponseRecorder {
	var buf *bytes.Buffer
	if body != nil {
		b, _ := json.Marshal(body)
		buf = bytes.NewBuffer(b)
	} else {
		buf = bytes.NewBuffer(nil)
	}
	req := httptest.NewRequest(method, path, buf)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

// GET /menu

func TestMenuHandler_GetAllMenuItems_HappyPath(t *testing.T) {
	r, mockSvc := setupMenuHandler(t)

	expected := []db.GetAllMenuItemsRow{
		{ID: 1, Name: "Classic Burger"},
		{ID: 2, Name: "Coke"},
	}
	mockSvc.On("GetAllMenuItems", mock.Anything).Return(expected, nil)

	w := doRequest(r, http.MethodGet, "/menu", nil)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.NotNil(t, resp["data"])
	mockSvc.AssertExpectations(t)
}

func TestMenuHandler_GetAllMenuItems_EmptyList(t *testing.T) {
	r, mockSvc := setupMenuHandler(t)

	mockSvc.On("GetAllMenuItems", mock.Anything).
		Return([]db.GetAllMenuItemsRow{}, nil)

	w := doRequest(r, http.MethodGet, "/menu", nil)

	assert.Equal(t, http.StatusOK, w.Code)
	// Empty list should still return 200, not 404
	mockSvc.AssertExpectations(t)
}

func TestMenuHandler_GetAllMenuItems_ServiceError(t *testing.T) {
	r, mockSvc := setupMenuHandler(t)

	mockSvc.On("GetAllMenuItems", mock.Anything).
		Return([]db.GetAllMenuItemsRow{}, errors.New("db down"))

	w := doRequest(r, http.MethodGet, "/menu", nil)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockSvc.AssertExpectations(t)
}

// GET /menu/:id

func TestMenuHandler_GetMenuItemByID_HappyPath(t *testing.T) {
	r, mockSvc := setupMenuHandler(t)

	expected := db.GetMenuItemByIDRow{ID: 1, Name: "Classic Burger"}
	mockSvc.On("GetMenuItemByID", mock.Anything, int32(1)).Return(expected, nil)

	w := doRequest(r, http.MethodGet, "/menu/1", nil)

	assert.Equal(t, http.StatusOK, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestMenuHandler_GetMenuItemByID_NotFound(t *testing.T) {
	r, mockSvc := setupMenuHandler(t)

	mockSvc.On("GetMenuItemByID", mock.Anything, int32(999)).
		Return(db.GetMenuItemByIDRow{}, services.ErrMenuItemNotFound)

	w := doRequest(r, http.MethodGet, "/menu/999", nil)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestMenuHandler_GetMenuItemByID_InvalidID_String(t *testing.T) {
	r, mockSvc := setupMenuHandler(t)

	// "abc" can't be parsed as int — handler should catch this
	// before even calling the service
	w := doRequest(r, http.MethodGet, "/menu/abc", nil)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockSvc.AssertNotCalled(t, "GetMenuItemByID")
}

func TestMenuHandler_GetMenuItemByID_InvalidID_Zero(t *testing.T) {
	r, mockSvc := setupMenuHandler(t)

	mockSvc.On("GetMenuItemByID", mock.Anything, int32(0)).
		Return(db.GetMenuItemByIDRow{}, services.ErrInvalidMenuItemID)

	w := doRequest(r, http.MethodGet, "/menu/0", nil)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestMenuHandler_GetMenuItemByID_ServiceError(t *testing.T) {
	r, mockSvc := setupMenuHandler(t)

	mockSvc.On("GetMenuItemByID", mock.Anything, int32(1)).
		Return(db.GetMenuItemByIDRow{}, errors.New("connection lost"))

	w := doRequest(r, http.MethodGet, "/menu/1", nil)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockSvc.AssertExpectations(t)
}

// POST /menu

func TestMenuHandler_CreateMenuItem_HappyPath(t *testing.T) {
	r, mockSvc := setupMenuHandler(t)

	reqBody := map[string]interface{}{
		"category_id": 1,
		"name":        "Classic Burger",
		"price":       "350.00",
		"description": "Juicy beef patty",
	}
	expected := db.MenuItem{ID: 1, Name: "Classic Burger"}

	mockSvc.On("CreateMenuItem", mock.Anything, mock.MatchedBy(func(arg db.CreateMenuItemParams) bool {
		// We use MatchedBy because Description is a sql.NullString
		// which makes exact struct matching awkward.
		// We just verify the important fields.
		return arg.CategoryID == 1 && arg.Name == "Classic Burger" && arg.Price == "350.00"
	})).Return(expected, nil)

	w := doRequest(r, http.MethodPost, "/menu", reqBody)

	assert.Equal(t, http.StatusCreated, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestMenuHandler_CreateMenuItem_MissingName(t *testing.T) {
	r, mockSvc := setupMenuHandler(t)

	// "name" is binding:"required" — Gin catches this before service is called
	reqBody := map[string]interface{}{
		"category_id": 1,
		"price":       "350.00",
		// name is missing
	}

	w := doRequest(r, http.MethodPost, "/menu", reqBody)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockSvc.AssertNotCalled(t, "CreateMenuItem")
}

func TestMenuHandler_CreateMenuItem_MissingPrice(t *testing.T) {
	r, mockSvc := setupMenuHandler(t)

	reqBody := map[string]interface{}{
		"category_id": 1,
		"name":        "Classic Burger",
		// price is missing
	}

	w := doRequest(r, http.MethodPost, "/menu", reqBody)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockSvc.AssertNotCalled(t, "CreateMenuItem")
}

func TestMenuHandler_CreateMenuItem_EmptyBody(t *testing.T) {
	r, mockSvc := setupMenuHandler(t)

	w := doRequest(r, http.MethodPost, "/menu", nil)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockSvc.AssertNotCalled(t, "CreateMenuItem")
}

func TestMenuHandler_CreateMenuItem_ValidationError_FromService(t *testing.T) {
	r, mockSvc := setupMenuHandler(t)

	reqBody := map[string]interface{}{
		"category_id": 1,
		"name":        "Classic Burger",
		"price":       "350.00",
	}

	// Service rejects it — e.g. invalid category
	mockSvc.On("CreateMenuItem", mock.Anything, mock.Anything).
		Return(db.MenuItem{}, services.ErrInvalidCategoryID)

	w := doRequest(r, http.MethodPost, "/menu", reqBody)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestMenuHandler_CreateMenuItem_ServiceError(t *testing.T) {
	r, mockSvc := setupMenuHandler(t)

	reqBody := map[string]interface{}{
		"category_id": 1,
		"name":        "Classic Burger",
		"price":       "350.00",
	}

	mockSvc.On("CreateMenuItem", mock.Anything, mock.Anything).
		Return(db.MenuItem{}, errors.New("db error"))

	w := doRequest(r, http.MethodPost, "/menu", reqBody)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockSvc.AssertExpectations(t)
}

// PUT /menu/:id

func TestMenuHandler_UpdateMenuItem_HappyPath(t *testing.T) {
	r, mockSvc := setupMenuHandler(t)

	reqBody := map[string]interface{}{
		"category_id": 1,
		"name":        "Updated Burger",
		"price":       "400.00",
		"available":   true,
	}
	expected := db.MenuItem{ID: 1, Name: "Updated Burger"}

	mockSvc.On("UpdateMenuItem", mock.Anything, mock.MatchedBy(func(arg db.UpdateMenuItemParams) bool {
		return arg.ID == 1 && arg.Name == "Updated Burger"
	})).Return(expected, nil)

	w := doRequest(r, http.MethodPut, "/menu/1", reqBody)

	assert.Equal(t, http.StatusOK, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestMenuHandler_UpdateMenuItem_NotFound(t *testing.T) {
	r, mockSvc := setupMenuHandler(t)

	reqBody := map[string]interface{}{
		"category_id": 1,
		"name":        "Ghost Burger",
		"price":       "400.00",
	}

	mockSvc.On("UpdateMenuItem", mock.Anything, mock.Anything).
		Return(db.MenuItem{}, services.ErrMenuItemNotFound)

	w := doRequest(r, http.MethodPut, "/menu/999", reqBody)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestMenuHandler_UpdateMenuItem_InvalidID_String(t *testing.T) {
	r, mockSvc := setupMenuHandler(t)

	reqBody := map[string]interface{}{
		"category_id": 1,
		"name":        "Burger",
		"price":       "400.00",
	}

	w := doRequest(r, http.MethodPut, "/menu/abc", reqBody)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockSvc.AssertNotCalled(t, "UpdateMenuItem")
}

func TestMenuHandler_UpdateMenuItem_InvalidBody(t *testing.T) {
	r, mockSvc := setupMenuHandler(t)

	w := doRequest(r, http.MethodPut, "/menu/1", nil)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockSvc.AssertNotCalled(t, "UpdateMenuItem")
}

// DELETE /menu/:id

func TestMenuHandler_DeleteMenuItem_HappyPath(t *testing.T) {
	r, mockSvc := setupMenuHandler(t)

	mockSvc.On("DeleteMenuItem", mock.Anything, int32(1)).Return(nil)

	w := doRequest(r, http.MethodDelete, "/menu/1", nil)

	// 204 No Content — successful delete has no body
	assert.Equal(t, http.StatusNoContent, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestMenuHandler_DeleteMenuItem_NotFound(t *testing.T) {
	r, mockSvc := setupMenuHandler(t)

	mockSvc.On("DeleteMenuItem", mock.Anything, int32(999)).
		Return(services.ErrMenuItemNotFound)

	w := doRequest(r, http.MethodDelete, "/menu/999", nil)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestMenuHandler_DeleteMenuItem_InvalidID_String(t *testing.T) {
	r, mockSvc := setupMenuHandler(t)

	w := doRequest(r, http.MethodDelete, "/menu/abc", nil)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockSvc.AssertNotCalled(t, "DeleteMenuItem")
}

func TestMenuHandler_DeleteMenuItem_ServiceError(t *testing.T) {
	r, mockSvc := setupMenuHandler(t)

	mockSvc.On("DeleteMenuItem", mock.Anything, int32(1)).
		Return(errors.New("db error"))

	w := doRequest(r, http.MethodDelete, "/menu/1", nil)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockSvc.AssertExpectations(t)
}
