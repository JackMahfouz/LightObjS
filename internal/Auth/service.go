package auth

import (
	"LightObjS/internal/domain"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserAlreadyExists = errors.New("user with this username already exists")
	ErrUserNotFound      = errors.New("user not found")
	ErrInvalidCredentials  = errors.New("invalid credentials")
)

// AuthService defines the interface for authentication operations.
type AuthService interface {
	Register(username, password string, isAdmin bool) (*domain.User, error)
	Login(username, password string) (string, error) // Returns JWT token
}

// authService implements AuthService.
type authService struct {
	userRepo UserRepository
	jwtSecret  []byte
}

// NewAuthService creates a new AuthService instance.
func NewAuthService(userRepo UserRepository, jwtSecret string) AuthService {
	return &authService{
		userRepo: userRepo,
		jwtSecret:  []byte(jwtSecret),
	}
}

// Register registers a new user with hashed password.
func (s *authService) Register(username, password string, isAdmin bool) (*domain.User, error) {
	// Check if user already exists
	existingUser, err := s.userRepo.GetUserByUsername(username)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing user: %w", err)
	}
	if existingUser != nil {
		return nil, ErrUserAlreadyExists
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user := &domain.User{
		Username:  username,
		Password:  string(hashedPassword), // Store hashed password
		CreatedAt: time.Now(),
		IsAdmin:   isAdmin,
	}

	if err := s.userRepo.CreateUser(user); err != nil {
		return nil, fmt.Errorf("failed to create user in repository: %w", err)
	}

	return user, nil
}

// Login authenticates a user and returns a JWT token.
func (s *authService) Login(username, password string) (string, error) {
	user, err := s.userRepo.GetUserByUsername(username)
	if err != nil {
		return "", fmt.Errorf("failed to retrieve user: %w", err)
	}
	if user == nil {
		return "", ErrUserNotFound
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", ErrInvalidCredentials
	}

	// Generate JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userID":   user.ID,
		"username": user.Username,
		"exp":      time.Now().Add(time.Hour * 24).Unix(), // Token expires in 24 hours
	})

	tokenString, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}

// HashPassword hashes a plain-text password using bcrypt.
func HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(hashedPassword), nil
}
