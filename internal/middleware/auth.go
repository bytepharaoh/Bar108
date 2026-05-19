package middleware

import (
	"bar108/internal/apperror"
	"bar108/internal/cache"
	"bar108/internal/jwt"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type contextKey string

const (
	ContextUserID contextKey = "user_id"
	ContextRole   contextKey = "role"
	ContextJTI    contextKey = "jti" // ← add this
)

// AuthMiddleware now also checks token blacklist in Redis.
// The cache client is injected — same pattern as everything else.
func AuthMiddleware(jwtManager *jwt.Manager, cache *cache.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			apperror.Respond(c, apperror.ErrUnauthorized)
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			apperror.Respond(c, apperror.ErrUnauthorized)
			c.Abort()
			return
		}

		claims, err := jwtManager.Verify(parts[1])
		if err != nil {
			apperror.Respond(c, apperror.ErrUnauthorized)
			c.Abort()
			return
		}

		// Check if this specific token has been blacklisted (logged out).
		// This is what makes logout actually work with stateless JWTs.
		blacklisted, err := cache.IsTokenBlacklisted(c.Request.Context(), claims.ID)
		if err != nil {
			// Redis is down — fail open or closed?
			// We fail CLOSED (reject the request) because security
			// is more important than availability here.
			apperror.Respond(c, apperror.ErrInternal)
			c.Abort()
			return
		}
		if blacklisted {
			apperror.Respond(c, apperror.ErrUnauthorized)
			c.Abort()
			return
		}

		c.Set(string(ContextUserID), claims.UserID)
		c.Set(string(ContextRole), claims.Role)
		c.Set(string(ContextJTI), claims.ID) // ← store JTI for logout
		c.Next()
	}
}

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

// RateLimitMiddleware limits requests per IP per window.
// limit = max requests, window = time window duration.
func RateLimitMiddleware(cache *cache.Client, limit int64, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()

		count, err := cache.IncrementRateLimit(c.Request.Context(), ip, window)
		if err != nil {
			// Redis down — fail open (allow the request)
			// Rate limiting is not worth breaking the app for
			c.Next()
			return
		}

		if count > limit {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "too many requests — slow down",
			})
			c.Abort()
			return
		}

		// Tell the client how many requests they have left
		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", limit))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", limit-count))

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

func GetRole(c *gin.Context) (string, bool) {
	val, exists := c.Get(string(ContextRole))
	if !exists {
		return "", false
	}
	role, ok := val.(string)
	return role, ok
}

func GetJTI(c *gin.Context) (string, bool) {
	val, exists := c.Get(string(ContextJTI))
	if !exists {
		return "", false
	}
	jti, ok := val.(string)
	return jti, ok
}
