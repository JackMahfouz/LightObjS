package auth

import (
	"LightObjS/internal/domain"
	"net/http"
	"strings"

	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// UserContextKey is the key to store the user in gin.Context
const UserContextKey = "authUser"

// Middleware struct holds dependencies for authentication middleware
type Middleware struct {
	jwtSecret []byte
	userRepo  UserRepository
}

// NewAuthMiddleware creates a new instance of AuthMiddleware
func NewAuthMiddleware(jwtSecret string, userRepo UserRepository) *Middleware {
	return &Middleware{
		jwtSecret: []byte(jwtSecret),
		userRepo:  userRepo,
	}
}

// AuthRequired is a Gin middleware to protect routes
func (m *Middleware) AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := m.extractToken(c)
		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization token required"})
			c.Abort()
			return
		}

		claims := jwt.MapClaims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return m.jwtSecret, nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		userID, ok := claims["userID"].(string)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			c.Abort()
			return
		}

		user, err := m.userRepo.GetUserByID(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve user"})
			c.Abort()
			return
		}
		if user == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
			c.Abort()
			return
		}

		// Store user in context
		c.Set(UserContextKey, user)
		c.Next()
	}
}

// extractToken extracts the JWT token from the Authorization header
func (m *Middleware) extractToken(c *gin.Context) string {
	bearerToken := c.GetHeader("Authorization")
	if strings.HasPrefix(bearerToken, "Bearer ") {
		return strings.TrimPrefix(bearerToken, "Bearer ")
	}
	return ""
}

// GetUserFromContext retrieves the authenticated user from the Gin context.
func GetUserFromContext(c *gin.Context) *domain.User {
	if user, exists := c.Get(UserContextKey); exists {
		if authUser, ok := user.(*domain.User); ok {
			return authUser
		}
	}
	return nil
}
