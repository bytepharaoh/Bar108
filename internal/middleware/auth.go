package middleware

import (
	"bar108/internal/apperror"
	"bar108/internal/jwt"
	"strings"

	"github.com/gin-gonic/gin"
)

type contextKey string

const (
	ContextUserID contextKey = "user_id"
	ContextRole   contextKey = "role"
)

// AuthMiddleware verifies the JWT token on every protected request.
// Flow:
// 1. Read Authorization header
// 2. Extract "Bearer <token>"
// 3. Verify token signature + expiry
// 4. Inject user_id and role into Gin context
// 5. Call next handler
//
// If anything fails → 401 Unauthorized, request stops here.

func AuthMiddleware(jwtManager *jwt.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Read the Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			apperror.Respond(c, apperror.ErrUnauthorized)
			c.Abort() // stop processing — don't call next handlers
			return
		} // Header format must be "Bearer <token>"
		// Split on space — we expect exactly 2 parts
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			apperror.Respond(c, apperror.ErrUnauthorized)
			c.Abort()
			return
		}
		tokenStr := parts[1]
		// Verify the token — checks signature + expiry
		claims, err := jwtManager.Verify(tokenStr)
		if err != nil {
			apperror.Respond(c, apperror.ErrUnauthorized)
			c.Abort()
			return
		}
		// Inject claims into context so handlers can read them
		// without knowing anything about JWT
		c.Set(string(ContextUserID), claims.UserID)
		c.Set(string(ContextRole), claims.Role)

		// All good — proceed to the actual handler
		c.Next()
	}
}

// AdminMiddleware ensures the user has the admin role.
func AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get(string(ContextRole))
		if !exists {
			apperror.Respond(c, apperror.ErrUnauthorized)
			c.Abort()
			return
		}

		if role != "admin" {
			apperror.Respond(c, apperror.ErrForbidden)
			c.Abort()
			return
		}

		c.Next()
	}
}
func GetUserID(c *gin.Context) (int32, bool) {
	val, exists := c.Get(string(ContextUserID))
	if !exists {
		return 0, false
	}
	id, ok := val.(int32)
	return id, ok
}

// GetRole is a helper to read the current user's role from context.
func GetRole(c *gin.Context) (string, bool) {
	val, exists := c.Get(string(ContextRole))
	if !exists {
		return "", false
	}
	role, ok := val.(string)
	return role, ok
}
