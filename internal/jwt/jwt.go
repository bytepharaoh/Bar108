package jwt

import (
	"bar108/config"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
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
	claims := Claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			// ExpiresAt — after this time the token is rejected
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(
				time.Duration(m.expiryHours) * time.Hour,
			)),
			// IssuedAt — when the token was created
			IssuedAt: jwt.NewNumericDate(time.Now()),
		},
	}

	// Create the token with our claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign it with our secret key — this creates the signature
	return token.SignedString(m.secret)
}
func (m *Manager) Verify(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(
		tokenStr,
		&Claims{},
		// This function returns the key used to verify the signature.
		// We verify the algorithm first — prevents algorithm switching attacks
		// where an attacker changes HS256 to none to bypass verification.
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
