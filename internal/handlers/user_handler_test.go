// package handlers_test

// import (
// 	"bar108/internal/db"
// 	"bar108/internal/handlers"
// 	"bar108/internal/services"
// 	servicemocks "bar108/internal/services/mocks"
// 	"errors"
// 	"net/http"
// 	"testing"

// 	"github.com/gin-gonic/gin"
// 	"github.com/stretchr/testify/assert"
// 	"github.com/stretchr/testify/mock"
// )

// func setupUserHandler(t *testing.T) (*gin.Engine, *servicemocks.MockUserService) {
// 	gin.SetMode(gin.TestMode)
// 	mockSvc := new(servicemocks.MockUserService)
// 	h := handlers.NewUserHandler(mockSvc)

// 	r := gin.New()
// 	r.GET("/users", h.GetAllUsers)
// 	r.GET("/users/active", h.GetActiveUsers)
// 	r.GET("/users/:id", h.GetUserByID)
// 	r.POST("/users", h.CreateUser)
// 	r.PUT("/users/:id", h.UpdateUser)
// 	r.PATCH("/users/:id/bonus", h.UpdateUserBonusPoints)
// 	r.PATCH("/users/:id/activate", h.ActivateUser)
// 	r.PATCH("/users/:id/deactivate", h.DeactivateUser)

// 	return r, mockSvc
// }

// // GET /users

// func TestUserHandler_GetAllUsers_HappyPath(t *testing.T) {
// 	r, mockSvc := setupUserHandler(t)

// 	expected := []db.User{{ID: 1, Name: "Ahmed"}, {ID: 2, Name: "Ali"}}
// 	mockSvc.On("GetAllUsers", mock.Anything).Return(expected, nil)

// 	w := doRequest(r, http.MethodGet, "/users", nil)

// 	assert.Equal(t, http.StatusOK, w.Code)
// 	mockSvc.AssertExpectations(t)
// }

// func TestUserHandler_GetAllUsers_Empty(t *testing.T) {
// 	r, mockSvc := setupUserHandler(t)

// 	mockSvc.On("GetAllUsers", mock.Anything).Return([]db.User{}, nil)

// 	w := doRequest(r, http.MethodGet, "/users", nil)

// 	assert.Equal(t, http.StatusOK, w.Code)
// 	mockSvc.AssertExpectations(t)
// }

// func TestUserHandler_GetAllUsers_ServiceError(t *testing.T) {
// 	r, mockSvc := setupUserHandler(t)

// 	mockSvc.On("GetAllUsers", mock.Anything).Return([]db.User{}, errors.New("db error"))

// 	w := doRequest(r, http.MethodGet, "/users", nil)

// 	assert.Equal(t, http.StatusInternalServerError, w.Code)
// 	mockSvc.AssertExpectations(t)
// }

// // GET /users/:id

// func TestUserHandler_GetUserByID_HappyPath(t *testing.T) {
// 	r, mockSvc := setupUserHandler(t)

// 	mockSvc.On("GetUserByID", mock.Anything, int32(1)).
// 		Return(db.User{ID: 1, Name: "Ahmed"}, nil)

// 	w := doRequest(r, http.MethodGet, "/users/1", nil)

// 	assert.Equal(t, http.StatusOK, w.Code)
// 	mockSvc.AssertExpectations(t)
// }

// func TestUserHandler_GetUserByID_NotFound(t *testing.T) {
// 	r, mockSvc := setupUserHandler(t)

// 	mockSvc.On("GetUserByID", mock.Anything, int32(999)).
// 		Return(db.User{}, services.ErrUserNotFound)

// 	w := doRequest(r, http.MethodGet, "/users/999", nil)

// 	assert.Equal(t, http.StatusNotFound, w.Code)
// 	mockSvc.AssertExpectations(t)
// }

// func TestUserHandler_GetUserByID_InvalidID_String(t *testing.T) {
// 	r, mockSvc := setupUserHandler(t)

// 	w := doRequest(r, http.MethodGet, "/users/abc", nil)

// 	assert.Equal(t, http.StatusBadRequest, w.Code)
// 	mockSvc.AssertNotCalled(t, "GetUserByID")
// }

// // POST /users

// func TestUserHandler_CreateUser_HappyPath(t *testing.T) {
// 	r, mockSvc := setupUserHandler(t)

// 	reqBody := map[string]interface{}{
// 		"name":          "Ahmed",
// 		"phone":         "+79001234567",
// 		"email":         "ahmed@bar108.com",
// 		"password_hash": "hashed123",
// 	}
// 	mockSvc.On("CreateUser", mock.Anything, mock.Anything).
// 		Return(db.User{ID: 1, Name: "Ahmed"}, nil)

// 	w := doRequest(r, http.MethodPost, "/users", reqBody)

// 	assert.Equal(t, http.StatusCreated, w.Code)
// 	mockSvc.AssertExpectations(t)
// }

// func TestUserHandler_CreateUser_MissingFields(t *testing.T) {
// 	r, mockSvc := setupUserHandler(t)

// 	// Missing required fields — Gin binding catches this
// 	reqBody := map[string]interface{}{
// 		"name": "Ahmed",
// 		// phone, email, password_hash missing
// 	}

// 	w := doRequest(r, http.MethodPost, "/users", reqBody)

// 	assert.Equal(t, http.StatusBadRequest, w.Code)
// 	mockSvc.AssertNotCalled(t, "CreateUser")
// }

// func TestUserHandler_CreateUser_InvalidEmail(t *testing.T) {
// 	r, mockSvc := setupUserHandler(t)

// 	reqBody := map[string]interface{}{
// 		"name":          "Ahmed",
// 		"phone":         "+79001234567",
// 		"email":         "notanemail",
// 		"password_hash": "hashed123",
// 	}
// 	mockSvc.On("CreateUser", mock.Anything, mock.Anything).
// 		Return(db.User{}, services.ErrInvalidUserEmail)

// 	w := doRequest(r, http.MethodPost, "/users", reqBody)

// 	assert.Equal(t, http.StatusBadRequest, w.Code)
// 	mockSvc.AssertExpectations(t)
// }

// func TestUserHandler_CreateUser_EmptyBody(t *testing.T) {
// 	r, mockSvc := setupUserHandler(t)

// 	w := doRequest(r, http.MethodPost, "/users", nil)

// 	assert.Equal(t, http.StatusBadRequest, w.Code)
// 	mockSvc.AssertNotCalled(t, "CreateUser")
// }

// func TestUserHandler_CreateUser_ServiceError(t *testing.T) {
// 	r, mockSvc := setupUserHandler(t)

// 	reqBody := map[string]interface{}{
// 		"name":          "Ahmed",
// 		"phone":         "+79001234567",
// 		"email":         "ahmed@bar108.com",
// 		"password_hash": "hashed123",
// 	}
// 	mockSvc.On("CreateUser", mock.Anything, mock.Anything).
// 		Return(db.User{}, errors.New("db error"))

// 	w := doRequest(r, http.MethodPost, "/users", reqBody)

// 	assert.Equal(t, http.StatusInternalServerError, w.Code)
// 	mockSvc.AssertExpectations(t)
// }

// // PATCH /users/:id/activate

// func TestUserHandler_ActivateUser_HappyPath(t *testing.T) {
// 	r, mockSvc := setupUserHandler(t)

// 	mockSvc.On("ActivateUser", mock.Anything, int32(1)).
// 		Return(db.User{ID: 1, IsActive: true}, nil)

// 	w := doRequest(r, http.MethodPatch, "/users/1/activate", nil)

// 	assert.Equal(t, http.StatusOK, w.Code)
// 	mockSvc.AssertExpectations(t)
// }

// func TestUserHandler_ActivateUser_AlreadyActive(t *testing.T) {
// 	r, mockSvc := setupUserHandler(t)

// 	mockSvc.On("ActivateUser", mock.Anything, int32(1)).
// 		Return(db.User{}, services.ErrUserAlreadyActive)

// 	w := doRequest(r, http.MethodPatch, "/users/1/activate", nil)

// 	// 409 Conflict
// 	assert.Equal(t, http.StatusConflict, w.Code)
// 	mockSvc.AssertExpectations(t)
// }

// func TestUserHandler_ActivateUser_NotFound(t *testing.T) {
// 	r, mockSvc := setupUserHandler(t)

// 	mockSvc.On("ActivateUser", mock.Anything, int32(999)).
// 		Return(db.User{}, services.ErrUserNotFound)

// 	w := doRequest(r, http.MethodPatch, "/users/999/activate", nil)

// 	assert.Equal(t, http.StatusNotFound, w.Code)
// 	mockSvc.AssertExpectations(t)
// }

// func TestUserHandler_ActivateUser_InvalidID(t *testing.T) {
// 	r, mockSvc := setupUserHandler(t)

// 	w := doRequest(r, http.MethodPatch, "/users/abc/activate", nil)

// 	assert.Equal(t, http.StatusBadRequest, w.Code)
// 	mockSvc.AssertNotCalled(t, "ActivateUser")
// }

// // PATCH /users/:id/deactivate

// func TestUserHandler_DeactivateUser_HappyPath(t *testing.T) {
// 	r, mockSvc := setupUserHandler(t)

// 	mockSvc.On("DeactivateUser", mock.Anything, int32(1)).
// 		Return(db.User{ID: 1, IsActive: false}, nil)

// 	w := doRequest(r, http.MethodPatch, "/users/1/deactivate", nil)

// 	assert.Equal(t, http.StatusOK, w.Code)
// 	mockSvc.AssertExpectations(t)
// }

// func TestUserHandler_DeactivateUser_AlreadyInactive(t *testing.T) {
// 	r, mockSvc := setupUserHandler(t)

// 	mockSvc.On("DeactivateUser", mock.Anything, int32(1)).
// 		Return(db.User{}, services.ErrUserAlreadyInactive)

// 	w := doRequest(r, http.MethodPatch, "/users/1/deactivate", nil)

// 	assert.Equal(t, http.StatusConflict, w.Code)
// 	mockSvc.AssertExpectations(t)
// }

// func TestUserHandler_DeactivateUser_HasActiveOrders(t *testing.T) {
// 	r, mockSvc := setupUserHandler(t)

// 	mockSvc.On("DeactivateUser", mock.Anything, int32(1)).
// 		Return(db.User{}, services.ErrUserHasActiveOrders)

// 	w := doRequest(r, http.MethodPatch, "/users/1/deactivate", nil)

// 	// 422 Unprocessable Entity
// 	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
// 	mockSvc.AssertExpectations(t)
// }

// func TestUserHandler_DeactivateUser_NotFound(t *testing.T) {
// 	r, mockSvc := setupUserHandler(t)

// 	mockSvc.On("DeactivateUser", mock.Anything, int32(999)).
// 		Return(db.User{}, services.ErrUserNotFound)

// 	w := doRequest(r, http.MethodPatch, "/users/999/deactivate", nil)

// 	assert.Equal(t, http.StatusNotFound, w.Code)
// 	mockSvc.AssertExpectations(t)
// }

// // PATCH /users/:id/bonus

// func TestUserHandler_UpdateBonusPoints_HappyPath(t *testing.T) {
// 	r, mockSvc := setupUserHandler(t)

// 	mockSvc.On("UpdateUserBonusPoints", mock.Anything, int32(1), int32(100)).
// 		Return(db.User{ID: 1, BonusPoints: 100}, nil)

// 	w := doRequest(r, http.MethodPatch, "/users/1/bonus", map[string]interface{}{
// 		"bonus_points": 100,
// 	})

// 	assert.Equal(t, http.StatusOK, w.Code)
// 	mockSvc.AssertExpectations(t)
// }

// func TestUserHandler_UpdateBonusPoints_NegativePoints(t *testing.T) {
// 	r, mockSvc := setupUserHandler(t)

// 	mockSvc.On("UpdateUserBonusPoints", mock.Anything, int32(1), int32(-50)).
// 		Return(db.User{}, services.ErrNegativeBonusPoints)

// 	w := doRequest(r, http.MethodPatch, "/users/1/bonus", map[string]interface{}{
// 		"bonus_points": -50,
// 	})

// 	assert.Equal(t, http.StatusBadRequest, w.Code)
// 	mockSvc.AssertExpectations(t)
// }

// func TestUserHandler_UpdateBonusPoints_MissingBody(t *testing.T) {
// 	r, mockSvc := setupUserHandler(t)

// 	w := doRequest(r, http.MethodPatch, "/users/1/bonus", nil)

// 	assert.Equal(t, http.StatusBadRequest, w.Code)
// 	mockSvc.AssertNotCalled(t, "UpdateUserBonusPoints")
// }

// func TestUserHandler_UpdateBonusPoints_InvalidID(t *testing.T) {
// 	r, mockSvc := setupUserHandler(t)

// 	w := doRequest(r, http.MethodPatch, "/users/abc/bonus", map[string]interface{}{
// 		"bonus_points": 100,
// 	})

// 	assert.Equal(t, http.StatusBadRequest, w.Code)
// 	mockSvc.AssertNotCalled(t, "UpdateUserBonusPoints")
// }
