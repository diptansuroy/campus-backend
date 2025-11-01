package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func RequireRole(role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		r, exists := c.Get("role")
		if !exists || r != role {
			c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden - insufficient permissions"})
			c.Abort()
			return
		}
		c.Next()
	}
}
