package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// AdminAuthRequired is a Gin middleware to protect routes requiring admin privileges.
func AdminAuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		authUser := GetUserFromContext(c)
		if authUser == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: User not authenticated"})
			c.Abort()
			return
		}

		if !authUser.IsAdmin {
			c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden: Admin privileges required"})
			c.Abort()
			return
		}

		c.Next()
	}
}
