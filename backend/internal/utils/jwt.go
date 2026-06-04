package utils

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/onecare/backend/internal/middleware"
)

func GenerateToken(userID int, role, secret string, expiryHours int) (string, int64, error) {
	expiry := time.Now().Add(time.Duration(expiryHours) * time.Hour)
	claims := &middleware.Claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiry),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(secret))
	return signed, expiry.Unix(), err
}
