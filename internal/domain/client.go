package domain

import "time"

// Client represents an API client that can access the storage system.
type Client struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`     // User who owns/created this client credential
	APIKey    string    `json:"api_key"`     // Publicly exposed part of the credential
	APISecret string    `json:"api_secret"`  // Hashed secret, not exposed
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"` // Optional expiry
	IsActive  bool      `json:"is_active"`   // Whether the client credential is active
}
