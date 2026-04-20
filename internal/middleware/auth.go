package middleware

import (
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID   string `json:"user_id"`
	PYCCode  string `json:"pyccode"`
	Role     string `json:"role"`
	IsBanned bool   `json:"is_banned"`
	jwt.RegisteredClaims
}

func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenStr := ""

		// Try cookie first, then Authorization header
		if cookie, err := c.Cookie("token"); err == nil {
			tokenStr = cookie
		} else if h := c.GetHeader("Authorization"); strings.HasPrefix(h, "Bearer ") {
			tokenStr = strings.TrimPrefix(h, "Bearer ")
		}

		if tokenStr == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "未登入"})
			return
		}

		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
			return []byte(os.Getenv("JWT_SECRET")), nil
		})
		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "登入已過期，請重新登入"})
			return
		}

		if claims.IsBanned {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "帳號已被封禁"})
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("pyccode", claims.PYCCode)
		c.Set("role", claims.Role)
		c.Next()
	}
}
