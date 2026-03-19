package domain

import "time"

// User represents a system user.
type User struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	Password  string    `json:"-"` // Store hashed password
	CreatedAt time.Time `json:"created_at"`
	IsAdmin   bool      `json:"is_admin"` // Indicates if the user has admin privileges
}
