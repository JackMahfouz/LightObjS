package auth

import (
	"LightObjS/internal/domain"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3" // SQLite driver
)

// UserRepository defines the interface for user data operations.
type UserRepository interface {
	CreateUser(user *domain.User) error
	GetUserByUsername(username string) (*domain.User, error)
	GetUserByID(id string) (*domain.User, error)
	GetAllUsers() ([]domain.User, error)
}

// GetAllUsers retrieves all users from the database.
func (r *SQLiteRepository) GetAllUsers() ([]domain.User, error) {
	rows, err := r.db.Query(
		"SELECT id, username, hashed_password, created_at, is_admin FROM users",
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get all users: %w", err)
	}
	defer rows.Close()

	var users []domain.User
	for rows.Next() {
		user := domain.User{}
		if err := rows.Scan(&user.ID, &user.Username, &user.Password, &user.CreatedAt, &user.IsAdmin); err != nil {
			return nil, fmt.Errorf("failed to scan user row: %w", err)
		}
		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return users, nil
}

// SQLiteRepository implements UserRepository for SQLite.
type SQLiteRepository struct {
	db *sql.DB
}

// NewSQLiteRepository creates a new SQLiteRepository instance.
func NewSQLiteRepository(dbPath string) (*SQLiteRepository, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Ping the database to verify the connection
	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	repo := &SQLiteRepository{db: db}

	return repo, nil
}

// Close closes the database connection.
func (r *SQLiteRepository) Close() error {
	return r.db.Close()
}

// DB returns the underlying *sql.DB connection.
func (r *SQLiteRepository) DB() *sql.DB {
	return r.db
}

// CreateUser inserts a new user into the database.
func (r *SQLiteRepository) CreateUser(user *domain.User) error {
	if user.ID == "" {
		user.ID = uuid.New().String()
	}
	if user.CreatedAt.IsZero() {
		user.CreatedAt = time.Now()
	}

	_, err := r.db.Exec(
		"INSERT INTO users (id, username, hashed_password, created_at, is_admin) VALUES (?, ?, ?, ?, ?)",
		user.ID, user.Username, user.Password, user.CreatedAt, user.IsAdmin,
	)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	return nil
}

// GetUserByUsername retrieves a user by their username.
func (r *SQLiteRepository) GetUserByUsername(username string) (*domain.User, error) {
	user := &domain.User{}
	err := r.db.QueryRow(
		"SELECT id, username, hashed_password, created_at, is_admin FROM users WHERE username = ?", username,
	).Scan(&user.ID, &user.Username, &user.Password, &user.CreatedAt, &user.IsAdmin)

	if err == sql.ErrNoRows {
		return nil, nil // User not found
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user by username: %w", err)
	}
	return user, nil
}

// GetUserByID retrieves a user by their ID.
func (r *SQLiteRepository) GetUserByID(id string) (*domain.User, error) {
	user := &domain.User{}
	err := r.db.QueryRow(
		"SELECT id, username, hashed_password, created_at, is_admin FROM users WHERE id = ?", id,
	).Scan(&user.ID, &user.Username, &user.Password, &user.CreatedAt, &user.IsAdmin)

	if err == sql.ErrNoRows {
		return nil, nil // User not found
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}
	return user, nil
}
