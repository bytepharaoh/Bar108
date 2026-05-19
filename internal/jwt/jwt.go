package jwt

import (
	"bar108/config"
	"errors"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Claims struct {
	UserID int32  `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

type Manager struct {
	secret      []byte
	expiryHours int
}

func New(cfg config.JWTConfig) *Manager {
	return &Manager{
		secret:      []byte(cfg.Secret),
		expiryHours: cfg.ExpiryHours,
	}
}

func (m *Manager) Generate(userID int32, role string) (string, error) {
	now := time.Now()
	expiresAt := now.Add(time.Duration(m.expiryHours) * time.Hour)

	claims := Claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			// jti — unique ID for this specific token.
			// This is what we store in Redis when blacklisting.
			ID:        uuid.NewString(),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.secret)
}

func (m *Manager) Verify(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(
		tokenStr,
		&Claims{},
		func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("unexpected signing method")
			}
			return m.secret, nil
		},
	)
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}
func (m *Manager) GetClaimsFromContext(c *gin.Context) (*Claims, error) {
	// re-parse from Authorization header
	authHeader := c.GetHeader("Authorization")
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 {
		return nil, errors.New("no token")
	}
	return m.Verify(parts[1])
}
