package auth

import (
	"LightObjS/internal/domain"
)

// AuthUser represents a user for authentication purposes.
// It embeds domain.User to leverage existing fields.
type AuthUser domain.User

// RegisterRequest represents the payload for user registration.
type RegisterRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginRequest represents the payload for user login.
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse represents the response after a successful login.
type LoginResponse struct {
	Token string `json:"token"`
}
