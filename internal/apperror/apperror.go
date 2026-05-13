package apperror

import "net/http"

type AppError struct {
	Code    int
	Message string
}

func (e *AppError) Error() string {
	return e.Message
}
func New(code int, message string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
	}
}

var (
	// 400 — client sent bad data
	ErrInvalidInput    = &AppError{Code: http.StatusBadRequest, Message: "invalid input"}
	ErrInvalidID       = &AppError{Code: http.StatusBadRequest, Message: "invalid id"}
	ErrZeroPrice       = &AppError{Code: http.StatusBadRequest, Message: "price must be greater than zero"}
	ErrNegativePrice   = &AppError{Code: http.StatusBadRequest, Message: "price cannot be negative"}
	ErrPromoExpired    = &AppError{Code: http.StatusBadRequest, Message: "promo code has expired"}
	ErrPromoExhausted  = &AppError{Code: http.StatusBadRequest, Message: "promo code usage limit reached"}
	ErrPromoNotFound   = &AppError{Code: http.StatusBadRequest, Message: "promo code not found"}
	ErrItemUnavailable = &AppError{Code: http.StatusBadRequest, Message: "menu item is not available"}

	// 404 — resource doesn't exist
	ErrNotFound      = &AppError{Code: http.StatusNotFound, Message: "resource not found"}
	ErrUserNotFound  = &AppError{Code: http.StatusNotFound, Message: "user not found"}
	ErrMenuNotFound  = &AppError{Code: http.StatusNotFound, Message: "menu item not found"}
	ErrOrderNotFound = &AppError{Code: http.StatusNotFound, Message: "order not found"}

	// 409 — conflict with current state
	ErrAlreadyActive   = &AppError{Code: http.StatusConflict, Message: "user is already active"}
	ErrAlreadyInactive = &AppError{Code: http.StatusConflict, Message: "user is already inactive"}
	ErrHasActiveOrders = &AppError{Code: http.StatusConflict, Message: "user has active orders and cannot be deactivated"}
	ErrAlreadyExists   = &AppError{Code: http.StatusConflict, Message: "resource already exists"}

	// 401 / 403 — auth errors (used in Phase 5)
	ErrUnauthorized = &AppError{Code: http.StatusUnauthorized, Message: "unauthorized"}
	ErrForbidden    = &AppError{Code: http.StatusForbidden, Message: "forbidden"}

	// 500 — our fault, never expose details
	ErrInternal = &AppError{Code: http.StatusInternalServerError, Message: "internal server error"}
)
