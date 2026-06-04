package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/onecare/backend/internal/config"
	"github.com/onecare/backend/pkg/response"
)

type Claims struct {
	UserID int    `json:"user_id"`
	Role   string `json:"role"` // "doctor" | "patient" | "admin"
	jwt.RegisteredClaims
}

// Auth validates the Bearer token and injects claims into context.
func Auth(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			response.Unauthorized(c, "missing or malformed authorization header")
			c.Abort()
			return
		}

		tokenStr := strings.TrimPrefix(header, "Bearer ")
		claims := &Claims{}

		token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(cfg.JWTSecret), nil
		})

		if err != nil || !token.Valid {
			response.Unauthorized(c, "invalid or expired token")
			c.Abort()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("role", claims.Role)
		c.Next()
	}
}

// RequireRole restricts a route to specific roles.
func RequireRole(roles ...string) gin.HandlerFunc {
	allowed := make(map[string]bool, len(roles))
	for _, r := range roles {
		allowed[r] = true
	}
	return func(c *gin.Context) {
		role, _ := c.Get("role")
		if !allowed[role.(string)] {
			response.Unauthorized(c, "you do not have permission to access this resource")
			c.Abort()
			return
		}
		c.Next()
	}
}
