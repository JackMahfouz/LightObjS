package auth

import (
	"LightObjS/internal/domain"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ClientContextKey is the key to store the authenticated client in gin.Context
const ClientContextKey = "authenticatedClient"

// ClientAuthMiddleware struct holds dependencies for client authentication middleware
type ClientAuthMiddleware struct {
	apiKeyService APIKeyService
}

// NewClientAuthMiddleware creates a new instance of ClientAuthMiddleware
func NewClientAuthMiddleware(apiKeyService APIKeyService) *ClientAuthMiddleware {
	return &ClientAuthMiddleware{
		apiKeyService: apiKeyService,
	}
}

// ClientAuthRequired is a Gin middleware to protect routes requiring API key authentication
func (m *ClientAuthMiddleware) ClientAuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader("X-API-Key")
		apiSecret := c.GetHeader("X-API-Secret")

		if apiKey == "" || apiSecret == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "API Key and Secret required"})
			c.Abort()
			return
		}

		client, err := m.apiKeyService.AuthenticateClient(apiKey, apiSecret)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid API Key or Secret"})
			c.Abort()
			return
		}
		if client == nil || !client.IsActive {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Client not found or inactive"})
			c.Abort()
			return
		}

		// Store client in context
		c.Set(ClientContextKey, client)
		c.Next()
	}
}

// GetClientFromContext retrieves the authenticated client from the Gin context.
func GetClientFromContext(c *gin.Context) *domain.Client {
	if client, exists := c.Get(ClientContextKey); exists {
		if authClient, ok := client.(*domain.Client); ok {
			return authClient
		}
	}
	return nil
}
