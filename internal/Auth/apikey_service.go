package auth

import (
	"LightObjS/internal/domain"
	"fmt"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrClientNotFound = fmt.Errorf("client not found")
)

// APIKeyService defines the interface for API key management and authentication.
type APIKeyService interface {
	GenerateClientCredentials(userID string) (*domain.Client, string, error) // Returns client and plain-text secret
	AuthenticateClient(apiKey, apiSecret string) (*domain.Client, error)
	GetClientsForUser(userID string) ([]domain.Client, error)
	RevokeClient(clientID string) error
}

type apiKeyService struct {
	clientRepo ClientRepository
}

// NewAPIKeyService creates a new APIKeyService instance.
func NewAPIKeyService(clientRepo ClientRepository) APIKeyService {
	return &apiKeyService{
		clientRepo: clientRepo,
	}
}

// GenerateClientCredentials generates a new API key and secret pair for a given user.
func (s *apiKeyService) GenerateClientCredentials(userID string) (*domain.Client, string, error) {
	apiKey := uuid.New().String()
	plainTextSecret := uuid.New().String() // Generate a random string for the secret

	hashedSecret, err := bcrypt.GenerateFromPassword([]byte(plainTextSecret), bcrypt.DefaultCost)
	if err != nil {
		return nil, "", fmt.Errorf("failed to hash API secret: %w", err)
	}

	client := &domain.Client{
		UserID:    userID,
		APIKey:    apiKey,
		APISecret: string(hashedSecret),
		CreatedAt: time.Now(),
		IsActive:  true,
	}

	if err := s.clientRepo.CreateClient(client); err != nil {
		return nil, "", fmt.Errorf("failed to create client in repository: %w", err)
	}

	return client, plainTextSecret, nil
}

// AuthenticateClient authenticates a client using their API key and secret.
func (s *apiKeyService) AuthenticateClient(apiKey, apiSecret string) (*domain.Client, error) {
	client, err := s.clientRepo.GetClientByAPIKey(apiKey)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve client: %w", err)
	}
	if client == nil || !client.IsActive {
		return nil, ErrClientNotFound
	}

	if err := bcrypt.CompareHashAndPassword([]byte(client.APISecret), []byte(apiSecret)); err != nil {
		return nil, ErrClientNotFound // Return generic not found for security
	}

	// Check if client has expired
	if client.ExpiresAt != nil && client.ExpiresAt.Before(time.Now()) {
		return nil, ErrClientNotFound
	}

	return client, nil
}

// GetClientsForUser retrieves all API clients associated with a specific user.
func (s *apiKeyService) GetClientsForUser(userID string) ([]domain.Client, error) {
	return s.clientRepo.GetClientsByUserID(userID)
}

// RevokeClient deactivates an API client.
func (s *apiKeyService) RevokeClient(clientID string) error {
	// A more robust implementation would retrieve the client, mark it inactive, and then update.
	// For simplicity, this assumes direct update or delete if `clientRepo` supports it.
	// For now, let's assume `UpdateClient` can be used to set IsActive to false.
	// This would require fetching the client first, modifying it, then saving.
	// For this example, let's implement a direct delete for simplicity, but a real system
	// should prefer deactivation.
	return s.clientRepo.DeleteClient(clientID)
}
