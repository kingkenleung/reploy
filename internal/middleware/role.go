package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func TeacherRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, _ := c.Get("role")
		if role != "teacher" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "需要教師權限"})
			return
		}
		c.Next()
	}
}
